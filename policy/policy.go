package policy

import (
	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

func Violations(report *finding.Report, threshold finding.Severity) []finding.Finding {
	return ViolationsByDomain(report, threshold, nil)
}

// ViolationsByDomain returns active findings that meet their domain-specific
// threshold. Domains without an override use fallback.
func ViolationsByDomain(report *finding.Report, fallback finding.Severity, overrides map[finding.Domain]finding.Severity) []finding.Finding {
	return violationsByDomain(report, fallback, overrides, false)
}

func NewViolationsByDomain(report *finding.Report, fallback finding.Severity, overrides map[finding.Domain]finding.Severity) []finding.Finding {
	return violationsByDomain(report, fallback, overrides, true)
}

func violationsByDomain(report *finding.Report, fallback finding.Severity, overrides map[finding.Domain]finding.Severity, newOnly bool) []finding.Finding {
	var result []finding.Finding
	for _, item := range report.Findings {
		if newOnly && item.BaselineState != finding.BaselineNew {
			continue
		}
		threshold := fallback
		if configured, ok := overrides[item.Domain]; ok {
			threshold = configured
		}
		if item.Severity.AtLeast(threshold) {
			result = append(result, item)
		}
	}
	return result
}
