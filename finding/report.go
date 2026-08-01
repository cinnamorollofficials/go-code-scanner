package finding

import "time"

type Location struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

type Finding struct {
	ID                string            `json:"id"`
	Fingerprint       string            `json:"fingerprint,omitempty"`
	RuleID            string            `json:"rule_id,omitempty"`
	Tool              string            `json:"tool"`
	Domain            Domain            `json:"domain"`
	Category          string            `json:"category"`
	Severity          Severity          `json:"severity"`
	Description       string            `json:"description"`
	Snippet           string            `json:"snippet,omitempty"`
	Recommendation    string            `json:"recommendation,omitempty"`
	Location          Location          `json:"location"`
	Suppressed        bool              `json:"suppressed"`
	SuppressionReason string            `json:"suppression_reason,omitempty"`
	Metadata          map[string]string `json:"metadata,omitempty"`
}

type Summary struct {
	Total             int `json:"total"`
	Critical          int `json:"critical"`
	High              int `json:"high"`
	Medium            int `json:"medium"`
	Low               int `json:"low"`
	Suppressed        int `json:"suppressed"`
	StaleSuppressions int `json:"stale_suppressions"`
}

type ScannerState string

const (
	ScannerClean    ScannerState = "clean"
	ScannerFindings ScannerState = "findings"
	ScannerPartial  ScannerState = "partial"
	ScannerSkipped  ScannerState = "skipped"
	ScannerFailed   ScannerState = "failed"
)

type ScannerStatus struct {
	ID       string        `json:"id"`
	State    ScannerState  `json:"state"`
	Required bool          `json:"required"`
	Duration time.Duration `json:"duration_ns"`
	Message  string        `json:"message,omitempty"`
	Version  string        `json:"version,omitempty"`
}

type Report struct {
	SchemaVersion         string          `json:"schema_version"`
	FingerprintVersion    string          `json:"fingerprint_version"`
	Timestamp             time.Time       `json:"timestamp"`
	Duration              time.Duration   `json:"duration_ns"`
	ScanMode              string          `json:"scan_mode"`
	Project               string          `json:"project"`
	Summary               Summary         `json:"summary"`
	Findings              []Finding       `json:"findings"`
	SuppressionsApplied   []Finding       `json:"suppressions_applied,omitempty"`
	StaleSuppressionFiles []string        `json:"stale_suppression_files,omitempty"`
	Scanners              []ScannerStatus `json:"scanner_status"`
	Warnings              []string        `json:"warnings,omitempty"`
}
