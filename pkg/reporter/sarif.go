package reporter

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
)

type sarifDocument struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name           string      `json:"name"`
	Version        string      `json:"version,omitempty"`
	InformationURI string      `json:"informationUri,omitempty"`
	Rules          []sarifRule `json:"rules"`
}

type sarifRule struct {
	ID               string         `json:"id"`
	ShortDescription sarifMessage   `json:"shortDescription"`
	Help             sarifMessage   `json:"help,omitempty"`
	HelpURI          string         `json:"helpUri,omitempty"`
	Properties       map[string]any `json:"properties,omitempty"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifResult struct {
	RuleID              string            `json:"ruleId"`
	Level               string            `json:"level"`
	Message             sarifMessage      `json:"message"`
	Locations           []sarifLocation   `json:"locations"`
	BaselineState       string            `json:"baselineState,omitempty"`
	PartialFingerprints map[string]string `json:"partialFingerprints,omitempty"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           sarifRegion           `json:"region"`
}

type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine int `json:"startLine"`
}

func WriteSARIF(path string, report *finding.Report) error {
	if report == nil {
		return fmt.Errorf("report is required")
	}
	rulesByID := make(map[string]sarifRule)
	results := make([]sarifResult, 0, len(report.Findings))
	for _, item := range report.Findings {
		help := item.Recommendation
		if help == "" {
			help = item.Description
		}
		rulesByID[item.RuleID] = sarifRule{
			ID: item.RuleID, ShortDescription: sarifMessage{Text: item.Description},
			Help: sarifMessage{Text: help}, HelpURI: item.Documentation,
			Properties: map[string]any{"tags": item.Tags, "fixable": item.Fixable, "domain": item.Domain},
		}
		line := item.Location.Line
		if line < 1 {
			line = 1
		}
		result := sarifResult{
			RuleID: item.RuleID, Level: sarifLevel(item.Severity),
			Message: sarifMessage{Text: item.Description},
			Locations: []sarifLocation{{PhysicalLocation: sarifPhysicalLocation{
				ArtifactLocation: sarifArtifactLocation{URI: item.Location.File},
				Region:           sarifRegion{StartLine: line},
			}}},
			PartialFingerprints: map[string]string{"securityReviewFingerprint/v" + report.FingerprintVersion: item.Fingerprint},
		}
		switch item.BaselineState {
		case finding.BaselineNew:
			result.BaselineState = "new"
		case finding.BaselineExisting:
			result.BaselineState = "unchanged"
		}
		results = append(results, result)
	}
	ruleIDs := make([]string, 0, len(rulesByID))
	for id := range rulesByID {
		ruleIDs = append(ruleIDs, id)
	}
	sort.Strings(ruleIDs)
	rules := make([]sarifRule, 0, len(ruleIDs))
	for _, id := range ruleIDs {
		rules = append(rules, rulesByID[id])
	}
	document := sarifDocument{
		Version: "2.1.0", Schema: "https://json.schemastore.org/sarif-2.1.0.json",
		Runs: []sarifRun{{Tool: sarifTool{Driver: sarifDriver{
			Name: "go-code-scanner", Version: report.ToolVersion,
			InformationURI: "https://github.com/cinnamorollofficials/go-code-scanner", Rules: rules,
		}}, Results: results}},
	}
	data, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return fmt.Errorf("encode SARIF report: %w", err)
	}
	return writeAtomic(path, append(data, '\n'), "SARIF report")
}

func sarifLevel(severity finding.Severity) string {
	switch severity {
	case finding.Critical, finding.High:
		return "error"
	case finding.Medium:
		return "warning"
	default:
		return "note"
	}
}
