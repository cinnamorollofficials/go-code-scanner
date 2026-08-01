package reporter

import (
	"encoding/json"
	"fmt"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

func WriteJSON(path string, report *finding.Report) error {
	if report == nil {
		return fmt.Errorf("report is required")
	}
	sanitized := *report
	sanitized.Findings = redactFindingSnippets(report.Findings)
	sanitized.SuppressionsApplied = redactFindingSnippets(report.SuppressionsApplied)
	data, err := json.MarshalIndent(&sanitized, "", "  ")
	if err != nil {
		return fmt.Errorf("encode report: %w", err)
	}
	return writeAtomic(path, append(data, '\n'), "report")
}

func redactFindingSnippets(input []finding.Finding) []finding.Finding {
	if input == nil {
		return nil
	}
	output := append([]finding.Finding(nil), input...)
	for index := range output {
		output[index].Snippet = ""
	}
	return output
}
