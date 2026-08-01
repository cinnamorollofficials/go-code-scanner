package adapters

import (
	"encoding/json"
	"fmt"
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
