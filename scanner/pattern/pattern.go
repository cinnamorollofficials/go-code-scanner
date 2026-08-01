package pattern

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/rules"
	"github.com/cinnamorollofficials/go-code-scanner/scanner"
)

type Scanner struct {
	genericRules  []rules.Compiled
	rulesBySuffix map[string][]rules.Compiled
	workers       int
	limits        Limits
}

type Limits struct {
	MaxFileBytes         int64
	MaxLineBytes         int
	QualityMaxFileBytes  int64
	QualityMaxLineLength int
}

func New(compiled []rules.Compiled, workers int, configured ...Limits) *Scanner {
	limits := Limits{MaxFileBytes: 2 * 1024 * 1024, MaxLineBytes: 1024 * 1024}
	if len(configured) > 0 {
		limits = configured[0]
	}
	s := &Scanner{rulesBySuffix: make(map[string][]rules.Compiled), workers: max(workers, 1), limits: limits}
	for _, rule := range compiled {
		if len(rule.Extensions) == 0 {
			s.genericRules = append(s.genericRules, rule)
			continue
		}
		for _, suffix := range rule.Extensions {
			suffix = strings.ToLower(suffix)
			s.rulesBySuffix[suffix] = append(s.rulesBySuffix[suffix], rule)
		}
	}
	return s
}

func (s *Scanner) ID() string { return "pattern" }

func (s *Scanner) Describe() scanner.Descriptor {
	return scanner.Descriptor{
		Capabilities:   []string{"built-in-rules", "line-patterns", "redaction"},
		SupportedModes: []string{"full", "changed", "staged"},
	}
}

func (s *Scanner) Scan(ctx context.Context, request scanner.Request) scanner.Result {
	started := time.Now()
	result := scanner.Result{State: finding.ScannerClean}
	fileFindings, fileErrors := s.scanFilePolicies(ctx, request)
	result.Findings = append(result.Findings, fileFindings...)
	for _, err := range fileErrors {
		result.Message = appendMessage(result.Message, err.Error())
	}
	if len(request.Sources) == 0 {
		if result.Message != "" {
			result.State, result.Failure = finding.ScannerPartial, scanner.FailurePartial
		} else if len(result.Findings) > 0 {
			result.State = finding.ScannerFindings
		}
		result.Duration = time.Since(started)
		return result
	}
	jobs := make(chan scanner.Source)
	outcomes := make(chan outcome, len(request.Sources))
	workerCount := min(s.workers, len(request.Sources))
	for range workerCount {
		go func() {
			for source := range jobs {
				items, err := s.scanSource(ctx, source, request.Root)
				outcomes <- outcome{findings: items, err: err}
			}
		}()
	}
	go feedSources(ctx, request.Sources, jobs)
	for completed := 0; completed < len(request.Sources); completed++ {
		select {
		case item := <-outcomes:
			result.Findings = append(result.Findings, item.findings...)
			if item.err != nil {
				result.Message = appendMessage(result.Message, item.err.Error())
			}
		case <-ctx.Done():
			result.State, result.Message, result.Failure = finding.ScannerFailed, ctx.Err().Error(), scanner.FailureCanceled
			result.Duration = time.Since(started)
			return result
		}
	}
	if result.Message != "" {
		result.State = finding.ScannerPartial
		result.Failure = scanner.FailurePartial
	} else if len(result.Findings) > 0 {
		result.State = finding.ScannerFindings
	}
	result.Duration = time.Since(started)
	return result
}

func (s *Scanner) scanFilePolicies(ctx context.Context, request scanner.Request) ([]finding.Finding, []error) {
	var findings []finding.Finding
	var failures []error
	knownFiles := make(map[string]struct{}, len(request.Files))
	for _, source := range request.Files {
		relative, err := filepath.Rel(request.Root, source.Path)
		if err == nil {
			knownFiles[strings.ToLower(filepath.ToSlash(relative))] = struct{}{}
		}
	}
	for _, source := range request.Files {
		if err := ctx.Err(); err != nil {
			return findings, append(failures, err)
		}
		relative, err := filepath.Rel(request.Root, source.Path)
		if err != nil {
			relative = source.Path
		}
		relative = filepath.ToSlash(relative)
		base := strings.ToLower(filepath.Base(relative))
		extension := strings.ToLower(filepath.Ext(base))
		if temporaryArtifact(base, extension) {
			findings = append(findings, fileFinding("temporary-artifact", finding.Quality, "repository_hygiene", finding.Medium,
				"Temporary atau build artifact ikut dalam perubahan", "Hapus artefak dan tambahkan pola yang sesuai ke .gitignore", relative, 1))
		}
		if base == "package.json" && !hasJavaScriptLockfile(knownFiles, filepath.ToSlash(filepath.Dir(relative))) {
			findings = append(findings, fileFinding("manifest-without-lockfile", finding.SupplyChain, "dependency_lock", finding.High,
				"JavaScript dependency manifest tidak memiliki lockfile pada directory yang sama", "Commit lockfile package manager untuk resolution dependency yang reproducible", relative, 1))
		}
		if base != "dockerfile" && !strings.HasPrefix(base, "dockerfile.") && request.Mode == "full" {
			if base != "go.mod" && base != "package.json" && !isGitHubWorkflow(relative) {
				continue
			}
		}
		reader, openErr := source.Open(ctx)
		if openErr != nil {
			failures = append(failures, fmt.Errorf("inspect %s: %w", relative, openErr))
			continue
		}
		content, readErr := io.ReadAll(io.LimitReader(reader, 64*1024+1))
		if readErr != nil {
			_ = reader.Close()
			failures = append(failures, fmt.Errorf("inspect %s: %w", relative, readErr))
			continue
		}
		lineScanner := bufio.NewScanner(bytes.NewReader(content))
		line := 0
		for lineScanner.Scan() {
			line++
			text := lineScanner.Text()
			if (base == "dockerfile" || strings.HasPrefix(base, "dockerfile.")) && strings.EqualFold(strings.TrimSpace(text), "USER root") {
				findings = append(findings, fileFinding("docker-root-user", finding.Hardening, "container_security", finding.High,
					"Docker container dikonfigurasi berjalan sebagai root", "Gunakan USER non-root dengan permission minimum", relative, line))
			}
			if (base == "dockerfile" || strings.HasPrefix(base, "dockerfile.")) && dockerLatest(text) {
				findings = append(findings, fileFinding("docker-latest-tag", finding.SupplyChain, "unpinned_dependency", finding.High,
					"Docker base image menggunakan mutable tag latest", "Pin base image ke digest sha256 atau version tag yang immutable", relative, line))
			}
			if base == "go.mod" && localGoReplace(text) {
				findings = append(findings, fileFinding("go-local-replace", finding.SupplyChain, "unpinned_dependency", finding.High,
					"go.mod menggunakan replace ke path lokal", "Gunakan module version yang dapat direproduksi atau workspace file yang tidak di-commit", relative, line))
			}
			if isGitHubWorkflow(relative) && mutableActionReference(text) {
				findings = append(findings, fileFinding("github-action-mutable-ref", finding.SupplyChain, "ci_dependency", finding.High,
					"GitHub Action menggunakan ref yang dapat berubah", "Pin action ke full commit SHA dan catat version tag sebagai komentar", relative, line))
			}
			if request.Mode != "full" && strings.Contains(text, "Code generated") && strings.Contains(text, "DO NOT EDIT") {
				findings = append(findings, fileFinding("generated-file-changed", finding.Quality, "generated_code", finding.Low,
					"Generated file termasuk dalam perubahan", "Pastikan file dihasilkan ulang dari source generator, bukan diedit manual", relative, line))
				break
			}
		}
		if base == "package.json" {
			findings = append(findings, unpinnedPackageDependencies(content, relative)...)
		}
		if scanErr := lineScanner.Err(); scanErr != nil {
			failures = append(failures, fmt.Errorf("inspect %s: %w", relative, scanErr))
		}
		if closeErr := reader.Close(); closeErr != nil {
			failures = append(failures, fmt.Errorf("close %s: %w", relative, closeErr))
		}
	}
	return findings, failures
}

func hasJavaScriptLockfile(files map[string]struct{}, directory string) bool {
	if directory == "." {
		directory = ""
	} else if directory != "" {
		directory += "/"
	}
	for _, name := range []string{"package-lock.json", "npm-shrinkwrap.json", "pnpm-lock.yaml", "yarn.lock", "bun.lock", "bun.lockb"} {
		if _, ok := files[strings.ToLower(directory+name)]; ok {
			return true
		}
	}
	return false
}

func dockerLatest(line string) bool {
	fields := strings.Fields(strings.TrimSpace(line))
	if len(fields) < 2 || !strings.EqualFold(fields[0], "FROM") {
		return false
	}
	image := strings.ToLower(fields[1])
	return strings.HasSuffix(image, ":latest")
}

func localGoReplace(line string) bool {
	parts := strings.SplitN(strings.TrimSpace(line), "=>", 2)
	if len(parts) != 2 {
		return false
	}
	target := strings.Fields(strings.TrimSpace(parts[1]))
	return len(target) > 0 && (strings.HasPrefix(target[0], "./") || strings.HasPrefix(target[0], "../") || filepath.IsAbs(target[0]))
}

func isGitHubWorkflow(path string) bool {
	path = strings.ToLower(filepath.ToSlash(path))
	return strings.HasPrefix(path, ".github/workflows/") && (strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml"))
}

func mutableActionReference(line string) bool {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "uses:") && !strings.HasPrefix(line, "- uses:") {
		return false
	}
	at := strings.LastIndex(line, "@")
	if at < 0 {
		return true
	}
	if strings.TrimSpace(line[at+1:]) == "" {
		return true
	}
	ref := strings.Fields(line[at+1:])[0]
	if len(ref) != 40 {
		return true
	}
	for _, character := range ref {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f') || (character >= 'A' && character <= 'F')) {
			return true
		}
	}
	return false
}

func unpinnedPackageDependencies(content []byte, path string) []finding.Finding {
	var manifest struct {
		Dependencies         map[string]string `json:"dependencies"`
		DevDependencies      map[string]string `json:"devDependencies"`
		OptionalDependencies map[string]string `json:"optionalDependencies"`
		PeerDependencies     map[string]string `json:"peerDependencies"`
	}
	if json.Unmarshal(content, &manifest) != nil {
		return nil
	}
	var findings []finding.Finding
	groups := []map[string]string{manifest.Dependencies, manifest.DevDependencies, manifest.OptionalDependencies, manifest.PeerDependencies}
	for _, group := range groups {
		for name, version := range group {
			value := strings.ToLower(strings.TrimSpace(version))
			if value != "*" && value != "latest" && !strings.HasPrefix(value, "git+") && !strings.HasPrefix(value, "github:") {
				continue
			}
			item := fileFinding("javascript-unpinned-dependency", finding.SupplyChain, "unpinned_dependency", finding.High,
				fmt.Sprintf("Dependency %s menggunakan version reference yang tidak dikunci", name), "Pin dependency ke version range yang terkontrol dan commit lockfile", path, 1)
			item.Metadata = map[string]string{"dependency": name, "version": version}
			findings = append(findings, item)
		}
	}
	return findings
}

func temporaryArtifact(base, extension string) bool {
	switch extension {
	case ".bak", ".class", ".dump", ".exe", ".orig", ".out", ".swp", ".tmp":
		return true
	}
	return strings.HasSuffix(base, "~") || base == "core"
}

func fileFinding(id string, domain finding.Domain, category string, severity finding.Severity, description, recommendation, path string, line int) finding.Finding {
	return finding.Finding{RuleID: id, Tool: "pattern", Domain: domain, Category: category, Severity: severity,
		Description: description, Recommendation: recommendation, Location: finding.Location{File: path, Line: line}}
}

type outcome struct {
	findings []finding.Finding
	err      error
}

func feedSources(ctx context.Context, sources []scanner.Source, jobs chan<- scanner.Source) {
	defer close(jobs)
	for _, source := range sources {
		select {
		case jobs <- source:
		case <-ctx.Done():
			return
		}
	}
}

func (s *Scanner) scanSource(ctx context.Context, source scanner.Source, root string) ([]finding.Finding, error) {
	file, err := source.Open(ctx)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	relative, err := filepath.Rel(root, source.Path)
	if err != nil {
		relative = source.Path
	}
	relative = filepath.ToSlash(relative)
	extension := strings.ToLower(filepath.Ext(source.Path))
	applicableRules := s.rulesFor(extension)
	var findings []finding.Finding
	lineNumber := 0
	counter := &countingReader{reader: io.LimitReader(file, s.limits.MaxFileBytes+1)}
	lineScanner := bufio.NewScanner(counter)
	lineScanner.Buffer(make([]byte, min(s.limits.MaxLineBytes, 64*1024)), s.limits.MaxLineBytes)
	for lineScanner.Scan() {
		lineNumber++
		line := lineScanner.Text()
		if s.limits.QualityMaxLineLength > 0 && len([]rune(line)) > s.limits.QualityMaxLineLength {
			findings = append(findings, fileFinding("line-length", finding.Quality, "maintainability", finding.Low,
				fmt.Sprintf("Baris melebihi batas %d karakter", s.limits.QualityMaxLineLength),
				"Pecah expression atau data panjang menjadi struktur yang lebih mudah ditinjau", relative, lineNumber))
		}
		if counter.bytes > s.limits.MaxFileBytes {
			return findings, fmt.Errorf("%s exceeds pattern file limit of %d bytes; increase pattern_max_file_bytes", relative, s.limits.MaxFileBytes)
		}
		for _, rule := range applicableRules {
			if !rule.Regex.MatchString(line) {
				continue
			}
			findings = append(findings, finding.Finding{
				RuleID: rule.ID, Tool: s.ID(), Domain: rule.Domain, Category: rule.Category,
				Severity: rule.Severity, Description: rule.Description,
				Recommendation: rule.Recommendation, Documentation: rule.Documentation,
				Tags: append([]string(nil), rule.Tags...), Fixable: rule.Fixable,
				Snippet:  redact(rule, line),
				Location: finding.Location{File: relative, Line: lineNumber},
			})
		}
	}
	if s.limits.QualityMaxFileBytes > 0 && counter.bytes > s.limits.QualityMaxFileBytes {
		findings = append(findings, fileFinding("source-file-size", finding.Quality, "maintainability", finding.Medium,
			fmt.Sprintf("Source file melebihi batas %d bytes", s.limits.QualityMaxFileBytes),
			"Pisahkan file berdasarkan tanggung jawab atau pindahkan data besar ke format yang sesuai", relative, 1))
	}
	if err := lineScanner.Err(); err != nil {
		if strings.Contains(err.Error(), "token too long") {
			return findings, fmt.Errorf("%s contains a line exceeding %d bytes; increase pattern_max_line_bytes", relative, s.limits.MaxLineBytes)
		}
		return findings, err
	}
	return findings, nil
}

type countingReader struct {
	reader io.Reader
	bytes  int64
}

func (r *countingReader) Read(buffer []byte) (int, error) {
	read, err := r.reader.Read(buffer)
	r.bytes += int64(read)
	return read, err
}

func (s *Scanner) rulesFor(extension string) []rules.Compiled {
	result := make([]rules.Compiled, 0, len(s.genericRules)+len(s.rulesBySuffix[extension]))
	result = append(result, s.genericRules...)
	result = append(result, s.rulesBySuffix[extension]...)
	return result
}

func redact(rule rules.Compiled, value string) string {
	if rule.Category == "secret_leak" {
		return "[REDACTED: " + rule.ID + "]"
	}
	if sensitiveRule(rule) {
		return "[REDACTED: potentially sensitive source line]"
	}
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) > 200 {
		value = string(runes[:200]) + "…"
	}
	return value
}

func sensitiveRule(rule rules.Compiled) bool {
	switch strings.ToLower(rule.Category) {
	case "authorization", "credential", "credentials", "personal_data", "secret", "secret_leak":
		return true
	}
	for _, tag := range rule.Tags {
		switch strings.ToLower(tag) {
		case "credential", "pii", "secret", "sensitive":
			return true
		}
	}
	return false
}

func appendMessage(current, addition string) string {
	if current == "" {
		return addition
	}
	return current + "; " + addition
}
