package policy

import (
	"sort"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

type Reason struct {
	Domain    finding.Domain   `json:"domain"`
	Threshold finding.Severity `json:"threshold"`
	Count     int              `json:"count"`
}

type Decision struct {
	Allowed    bool              `json:"allowed"`
	NewOnly    bool              `json:"new_only"`
	Violations []finding.Finding `json:"violations,omitempty"`
	Reasons    []Reason          `json:"reasons,omitempty"`
}

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

func Evaluate(report *finding.Report, fallback finding.Severity, overrides map[finding.Domain]finding.Severity, newOnly bool) Decision {
	violations := violationsByDomain(report, fallback, overrides, newOnly)
	counts := make(map[finding.Domain]int)
	for _, item := range violations {
		counts[item.Domain]++
	}
	domains := make([]finding.Domain, 0, len(counts))
	for domain := range counts {
		domains = append(domains, domain)
	}
	sort.Slice(domains, func(i, j int) bool { return domains[i] < domains[j] })
	reasons := make([]Reason, 0, len(domains))
	for _, domain := range domains {
		threshold := fallback
		if configured, ok := overrides[domain]; ok {
			threshold = configured
		}
		reasons = append(reasons, Reason{Domain: domain, Threshold: threshold, Count: counts[domain]})
	}
	return Decision{Allowed: len(violations) == 0, NewOnly: newOnly, Violations: violations, Reasons: reasons}
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
