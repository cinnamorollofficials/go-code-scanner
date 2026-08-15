package viewer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	securityreview "github.com/cinnamorollofficials/go-code-scanner"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/buildinfo"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/config"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/fixer"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/rules"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/suppression"
)

type Server struct {
	mu           sync.RWMutex
	root         string
	configPath   string
	latestReport *finding.Report
	lastScanTime time.Time
	staticFS     http.FileSystem
}

type ServerOptions struct {
	Root       string
	ConfigPath string
	StaticFS   http.FileSystem
}

func NewServer(opts ServerOptions) *Server {
	root := opts.Root
	if root == "" {
		root = "."
	}
	return &Server{
		root:       root,
		configPath: opts.ConfigPath,
		staticFS:   opts.StaticFS,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/report", s.handleGetReport)
	mux.HandleFunc("/api/scan", s.handleScan)
	mux.HandleFunc("/api/fix", s.handleFix)
	mux.HandleFunc("/api/suppress", s.handleSuppress)
	mux.HandleFunc("/api/file", s.handleFileContent)
	mux.HandleFunc("/api/rules", s.handleRules)

	if s.staticFS != nil {
		fileServer := http.FileServer(s.staticFS)
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.NotFound(w, r)
				return
			}
			fileServer.ServeHTTP(w, r)
		})
	} else {
		mux.HandleFunc("/", s.handleDefaultUI)
	}

	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type StatusResponse struct {
	ToolVersion  string    `json:"tool_version"`
	Root         string    `json:"root"`
	HasReport    bool      `json:"has_report"`
	LastScanTime time.Time `json:"last_scan_time,omitempty"`
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	absRoot, _ := filepath.Abs(s.root)
	resp := StatusResponse{
		ToolVersion:  buildinfo.String(),
		Root:         absRoot,
		HasReport:    s.latestReport != nil,
		LastScanTime: s.lastScanTime,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleGetReport(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	report := s.latestReport
	s.mu.RUnlock()

	if report == nil {
		var err error
		report, err = s.executeScan(r.Context(), ScanRequest{})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}
	writeJSON(w, http.StatusOK, report)
}

type ScanRequest struct {
	Mode    string `json:"mode"`     // "full", "changed", "staged"
	Profile string `json:"profile"`  // e.g. "default", "strict"
	Scope   string `json:"scope"`    // "all", "client", "server"
	FailOn  string `json:"fail_on"`  // "critical", "high", etc.
	AutoFix bool   `json:"auto_fix"` // apply deterministic fixes
}

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ScanRequest
	if r.Method == http.MethodPost && r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	report, err := s.executeScan(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, report)
}

func (s *Server) executeScan(ctx context.Context, req ScanRequest) (*finding.Report, error) {
	cfg, err := config.Load(s.configPath)
	if err != nil {
		cfg = config.Default()
	}
	cfg.Root = s.root

	switch strings.ToLower(req.Mode) {
	case "changed":
		cfg.Mode = config.ModeChanged
	case "staged":
		cfg.Mode = config.ModeStaged
	default:
		cfg.Mode = config.ModeFull
	}

	if req.Profile != "" {
		cfg.SelectedProfile = req.Profile
	}
	if req.Scope != "" {
		if parsedScope, scopeErr := config.ParseScanScope(req.Scope); scopeErr == nil {
			cfg.ScanScope = parsedScope
		}
	}
	if req.FailOn != "" {
		if sev, sevErr := finding.ParseSeverity(req.FailOn); sevErr == nil {
			cfg.FailOn = sev
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	reviewer, err := securityreview.New(cfg, securityreview.WithToolVersion(buildinfo.String()))
	if err != nil {
		return nil, fmt.Errorf("initialize reviewer: %w", err)
	}

	report, err := reviewer.Run(ctx)
	if report == nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	if req.AutoFix && cfg.Mode == config.ModeFull {
		changes, fixErr := fixer.Apply(cfg.Root, report.Findings, false)
		if fixErr == nil && len(changes) > 0 {
			if updatedReport, rescanErr := reviewer.Run(ctx); updatedReport != nil && rescanErr == nil {
				report = updatedReport
			}
		}
	}

	s.mu.Lock()
	s.latestReport = report
	s.lastScanTime = time.Now().UTC()
	s.mu.Unlock()

	return report, nil
}

type FixRequest struct {
	FindingID string `json:"finding_id,omitempty"`
	File      string `json:"file,omitempty"`
	Line      int    `json:"line,omitempty"`
	RuleID    string `json:"rule_id,omitempty"`
}

func (s *Server) handleFix(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	report := s.latestReport
	s.mu.RUnlock()

	if report == nil {
		http.Error(w, "no scan report available to fix", http.StatusBadRequest)
		return
	}

	var req FixRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	var targetFindings []finding.Finding
	if req.RuleID != "" && req.File != "" {
		for _, f := range report.Findings {
			if f.RuleID == req.RuleID && f.Location.File == req.File && (req.Line == 0 || f.Location.Line == req.Line) {
				targetFindings = append(targetFindings, f)
			}
		}
	} else {
		targetFindings = report.Findings
	}

	changes, err := fixer.Apply(s.root, targetFindings, false)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	updatedReport, _ := s.executeScan(r.Context(), ScanRequest{})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"changes": len(changes),
		"report":  updatedReport,
	})
}

type SuppressRequest struct {
	RuleID      string `json:"rule_id"`
	Fingerprint string `json:"fingerprint"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Reason      string `json:"reason"`
	Expires     string `json:"expires"`
	ApprovedBy  string `json:"approved_by,omitempty"`
	Ticket      string `json:"ticket,omitempty"`
}

func (s *Server) handleSuppress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SuppressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json payload: " + err.Error()})
		return
	}

	rule := suppression.Rule{
		RuleID:      req.RuleID,
		Fingerprint: req.Fingerprint,
		File:        req.File,
		Line:        req.Line,
		Reason:      req.Reason,
		Expires:     req.Expires,
		ApprovedBy:  req.ApprovedBy,
		Ticket:      req.Ticket,
	}

	suppressionPath := filepath.Join(s.root, ".security-review-suppressions.json")
	file, err := suppression.Add(suppressionPath, rule, false)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	updatedReport, _ := s.executeScan(r.Context(), ScanRequest{})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":      "suppression added successfully",
		"suppressions": len(file.Suppressions),
		"report":       updatedReport,
	})
}

func (s *Server) handleFileContent(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		http.Error(w, "missing path query parameter", http.StatusBadRequest)
		return
	}

	cleanPath := filepath.Clean(filePath)
	if filepath.IsAbs(cleanPath) || strings.HasPrefix(cleanPath, "..") {
		http.Error(w, "invalid file path", http.StatusBadRequest)
		return
	}

	fullPath := filepath.Join(s.root, cleanPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		http.Error(w, "file not found or unreadable: "+err.Error(), http.StatusNotFound)
		return
	}

	lines := strings.Split(string(data), "\n")
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"path":        cleanPath,
		"total_lines": len(lines),
		"lines":       lines,
	})
}

func (s *Server) handleRules(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load(s.configPath)
	if err != nil {
		cfg = config.Default()
	}
	cfg.Root = s.root

	ruleFiles := cfg.RuleFiles
	if len(ruleFiles) == 0 {
		ruleFiles = []string{".security-review-rules.json"}
	}

	var resolvedPaths []string
	for _, f := range ruleFiles {
		resolvedPaths = append(resolvedPaths, filepath.Join(cfg.Root, f))
	}

	compiled, _ := rules.Load(resolvedPaths)
	ruleDefs := make([]rules.Rule, len(compiled))
	for i, cr := range compiled {
		ruleDefs[i] = cr.Rule
	}
	if len(ruleDefs) == 0 {
		ruleDefs = rules.Default()
	}

	writeJSON(w, http.StatusOK, ruleDefs)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
