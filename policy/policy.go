package policy

import (
	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

func Violations(report *finding.Report, threshold finding.Severity) []finding.Finding {
	var result []finding.Finding
	for _, item := range report.Findings {
		if item.Severity.AtLeast(threshold) {
			result = append(result, item)
		}
	}
	return result
}
