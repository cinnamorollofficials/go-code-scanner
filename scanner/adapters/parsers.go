package adapters

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/scanner/command"
)

func parseGosec(data []byte) ([]command.ParsedFinding, error) {
	var report struct {
		Issues []struct {
			Severity   string `json:"severity"`
			Confidence string `json:"confidence"`
			RuleID     string `json:"rule_id"`
			Details    string `json:"details"`
			File       string `json:"file"`
			Line       string `json:"line"`
		} `json:"Issues"`
	}
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}
	result := make([]command.ParsedFinding, 0, len(report.Issues))
	for _, issue := range report.Issues {
		line, _ := strconv.Atoi(strings.Split(issue.Line, "-")[0])
		result = append(result, command.ParsedFinding{
			RuleID: issue.RuleID, Severity: adapterSeverity(issue.Severity), Description: issue.Details,
			File: issue.File, Line: line, Metadata: map[string]string{"confidence": issue.Confidence},
		})
	}
	return result, nil
}

func parseGitleaks(data []byte) ([]command.ParsedFinding, error) {
	var leaks []struct {
		RuleID      string `json:"RuleID"`
		Description string `json:"Description"`
		File        string `json:"File"`
		StartLine   int    `json:"StartLine"`
		Fingerprint string `json:"Fingerprint"`
		Commit      string `json:"Commit"`
	}
	if err := json.Unmarshal(data, &leaks); err != nil {
		return nil, err
	}
	result := make([]command.ParsedFinding, 0, len(leaks))
	for _, leak := range leaks {
		metadata := map[string]string{"fingerprint": leak.Fingerprint}
		if leak.Commit != "" {
			metadata["commit"] = leak.Commit
		}
		result = append(result, command.ParsedFinding{
			RuleID: leak.RuleID, Description: leak.Description, File: leak.File, Line: leak.StartLine, Metadata: metadata,
		})
	}
	return result, nil
}

func parseTrivy(data []byte) ([]command.ParsedFinding, error) {
	var report struct {
		Results []struct {
			Target          string `json:"Target"`
			Vulnerabilities []struct {
				ID               string `json:"VulnerabilityID"`
				Package          string `json:"PkgName"`
				InstalledVersion string `json:"InstalledVersion"`
				FixedVersion     string `json:"FixedVersion"`
				Severity         string `json:"Severity"`
				Title            string `json:"Title"`
				PrimaryURL       string `json:"PrimaryURL"`
			} `json:"Vulnerabilities"`
		} `json:"Results"`
	}
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}
	var result []command.ParsedFinding
	for _, target := range report.Results {
		for _, vulnerability := range target.Vulnerabilities {
			description := vulnerability.Title
			if description == "" {
				description = fmt.Sprintf("Vulnerable dependency %s %s", vulnerability.Package, vulnerability.InstalledVersion)
			}
			result = append(result, command.ParsedFinding{
				RuleID: vulnerability.ID, Severity: adapterSeverity(vulnerability.Severity), Description: description,
				Documentation: vulnerability.PrimaryURL, File: target.Target,
				Metadata: map[string]string{"package": vulnerability.Package, "installed_version": vulnerability.InstalledVersion, "fixed_version": vulnerability.FixedVersion},
			})
		}
	}
	return result, nil
}

func parseSemgrep(data []byte) ([]command.ParsedFinding, error) {
	var report struct {
		Results []struct {
			CheckID string `json:"check_id"`
			Path    string `json:"path"`
			Start   struct {
				Line int `json:"line"`
			} `json:"start"`
			Extra struct {
				Message  string `json:"message"`
				Severity string `json:"severity"`
			} `json:"extra"`
		} `json:"results"`
	}
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}
	result := make([]command.ParsedFinding, 0, len(report.Results))
	for _, item := range report.Results {
		result = append(result, command.ParsedFinding{
			RuleID: item.CheckID, Severity: adapterSeverity(item.Extra.Severity), Description: item.Extra.Message,
			File: item.Path, Line: item.Start.Line,
		})
	}
	return result, nil
}

func parseGovulncheck(data []byte) ([]command.ParsedFinding, error) {
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	seen := make(map[string]struct{})
	var result []command.ParsedFinding
	for {
		var message struct {
			Finding *struct {
				OSV          string `json:"osv"`
				FixedVersion string `json:"fixed_version"`
				Trace        []struct {
					Module   string `json:"module"`
					Version  string `json:"version"`
					Package  string `json:"package"`
					Function string `json:"function"`
					Position *struct {
						Filename string `json:"filename"`
						Line     int    `json:"line"`
					} `json:"position"`
				} `json:"trace"`
			} `json:"finding"`
		}
		if err := decoder.Decode(&message); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}
		if message.Finding == nil || message.Finding.OSV == "" {
			continue
		}
		item := command.ParsedFinding{RuleID: message.Finding.OSV, Documentation: "https://pkg.go.dev/vuln/" + message.Finding.OSV,
			Description: "Reachable Go vulnerability " + message.Finding.OSV, Metadata: map[string]string{"fixed_version": message.Finding.FixedVersion}}
		for _, frame := range message.Finding.Trace {
			item.Metadata["module"], item.Metadata["version"], item.Metadata["package"] = frame.Module, frame.Version, frame.Package
			if frame.Position != nil && frame.Position.Line > 0 {
				item.File, item.Line = frame.Position.Filename, frame.Position.Line
				item.Metadata["symbol"] = frame.Function
			}
		}
		key := fmt.Sprintf("%s\x00%s\x00%d", item.RuleID, item.File, item.Line)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, item)
	}
	return result, nil
}

func parseOSVScanner(data []byte) ([]command.ParsedFinding, error) {
	var report struct {
		Results []struct {
			Source struct {
				Path string `json:"path"`
			} `json:"source"`
			Packages []struct {
				Package struct {
					Name, Version, Ecosystem string
				} `json:"package"`
				Vulnerabilities []struct {
					ID               string `json:"id"`
					Summary          string `json:"summary"`
					Details          string `json:"details"`
					DatabaseSpecific struct {
						Severity string `json:"severity"`
					} `json:"database_specific"`
				} `json:"vulnerabilities"`
			} `json:"packages"`
		} `json:"results"`
	}
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}
	var result []command.ParsedFinding
	for _, scan := range report.Results {
		for _, pkg := range scan.Packages {
			for _, vulnerability := range pkg.Vulnerabilities {
				description := vulnerability.Summary
				if description == "" {
					description = "Dependency vulnerability " + vulnerability.ID
				}
				result = append(result, command.ParsedFinding{
					RuleID: vulnerability.ID, Severity: adapterSeverity(vulnerability.DatabaseSpecific.Severity), Description: description,
					Documentation: "https://osv.dev/vulnerability/" + vulnerability.ID, File: scan.Source.Path,
					Metadata: map[string]string{"package": pkg.Package.Name, "version": pkg.Package.Version, "ecosystem": pkg.Package.Ecosystem},
				})
			}
		}
	}
	return result, nil
}

func adapterSeverity(value string) finding.Severity {
	switch strings.ToUpper(value) {
	case "CRITICAL":
		return finding.Critical
	case "HIGH", "ERROR":
		return finding.High
	case "MEDIUM", "MODERATE", "WARNING", "WARN":
		return finding.Medium
	case "LOW", "INFO":
		return finding.Low
	default:
		return finding.Medium
	}
}

func parseESLint(data []byte) ([]command.ParsedFinding, error) {
	var files []struct {
		FilePath string `json:"filePath"`
		Messages []struct {
			RuleID   string `json:"ruleId"`
			Severity int    `json:"severity"`
			Line     int    `json:"line"`
			Column   int    `json:"column"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(data, &files); err != nil {
		return nil, err
	}
	var result []command.ParsedFinding
	for _, file := range files {
		for _, msg := range file.Messages {
			ruleID := msg.RuleID
			if ruleID == "" {
				ruleID = "eslint/unknown"
			}
			sev := finding.Medium
			if msg.Severity == 1 {
				sev = finding.Low
			} else if msg.Severity == 2 {
				sev = finding.High
			}
			result = append(result, command.ParsedFinding{
				RuleID:      ruleID,
				Severity:    sev,
				Description: fmt.Sprintf("ESLint violation for rule %s", ruleID),
				File:        file.FilePath,
				Line:        msg.Line,
			})
		}
	}
	return result, nil
}
