package finding

import (
	"fmt"
	"strings"
)

// Severity describes the impact assigned to a security finding.
type Severity string

const (
	Critical Severity = "CRITICAL"
	High     Severity = "HIGH"
	Medium   Severity = "MEDIUM"
	Low      Severity = "LOW"
)

var severityRanks = map[Severity]int{
	Critical: 0,
	High:     1,
	Medium:   2,
	Low:      3,
}

func ParseSeverity(value string) (Severity, error) {
	severity := Severity(strings.ToUpper(strings.TrimSpace(value)))
	if _, ok := severityRanks[severity]; !ok {
		return "", fmt.Errorf("invalid severity %q", value)
	}
	return severity, nil
}

func (s Severity) Valid() bool {
	_, ok := severityRanks[s]
	return ok
}

func (s Severity) Rank() int {
	if rank, ok := severityRanks[s]; ok {
		return rank
	}
	return len(severityRanks)
}

func (s Severity) AtLeast(threshold Severity) bool {
	return s.Valid() && threshold.Valid() && s.Rank() <= threshold.Rank()
}
