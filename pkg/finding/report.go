package finding

import "time"

type Confidence string

const (
	ConfidenceHigh    Confidence = "high"
	ConfidenceMedium  Confidence = "medium"
	ConfidenceLow     Confidence = "low"
	ConfidenceUnknown Confidence = "unknown"
)

func (c Confidence) Valid() bool {
	return c == ConfidenceHigh || c == ConfidenceMedium || c == ConfidenceLow || c == ConfidenceUnknown
}

type Exploitability string

const (
	ExploitabilityLikely   Exploitability = "likely"
	ExploitabilityUnlikely Exploitability = "unlikely"
	ExploitabilityUnknown  Exploitability = "unknown"
)

func (e Exploitability) Valid() bool {
	return e == ExploitabilityLikely || e == ExploitabilityUnlikely || e == ExploitabilityUnknown
}

type FindingState string

const (
	FindingCandidate             FindingState = "candidate"
	FindingProbable              FindingState = "probable"
	FindingConfirmed             FindingState = "confirmed"
	FindingNeedsContext          FindingState = "needs_context"
	FindingDismissedWithEvidence FindingState = "dismissed_with_evidence"
	FindingFixedVerified         FindingState = "fixed_verified"
	FindingFixedNotVerified      FindingState = "fixed_not_verified"
)

func (f FindingState) Valid() bool {
	return f == FindingCandidate || f == FindingProbable || f == FindingConfirmed ||
		f == FindingNeedsContext || f == FindingDismissedWithEvidence ||
		f == FindingFixedVerified || f == FindingFixedNotVerified
}

type StepType string

const (
	StepSource     StepType = "source"
	StepPropagator StepType = "propagator"
	StepTransform  StepType = "transform"
	StepSanitizer  StepType = "sanitizer"
	StepSink       StepType = "sink"
	StepBarrier    StepType = "barrier"
	StepUnknown    StepType = "unknown"
)

type Location struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

type DataflowStep struct {
	Type        StepType `json:"type"`
	Location    Location `json:"location"`
	Label       string   `json:"label,omitempty"`
	Explanation string   `json:"explanation,omitempty"`
}

type Finding struct {
	ID                string            `json:"id"`
	Fingerprint       string            `json:"fingerprint,omitempty"`
	RuleID            string            `json:"rule_id,omitempty"`
	Tool              string            `json:"tool"`
	Domain            Domain            `json:"domain"`
	Category          string            `json:"category"`
	Severity          Severity          `json:"severity"`
	Confidence        Confidence        `json:"confidence,omitempty"`
	Exploitability    Exploitability    `json:"exploitability,omitempty"`
	FindingState      FindingState      `json:"finding_state,omitempty"`
	Description       string            `json:"description"`
	Snippet           string            `json:"snippet,omitempty"`
	Recommendation    string            `json:"recommendation,omitempty"`
	Documentation     string            `json:"documentation,omitempty"`
	Tags              []string          `json:"tags,omitempty"`
	Fixable           bool              `json:"fixable,omitempty"`
	Location          Location          `json:"location"`
	Dataflow          []DataflowStep    `json:"dataflow,omitempty"`
	Suppressed        bool              `json:"suppressed"`
	SuppressionReason string            `json:"suppression_reason,omitempty"`
	BaselineState     BaselineState     `json:"baseline_state,omitempty"`
	Metadata          map[string]string `json:"metadata,omitempty"`
}

type Summary struct {
	Total             int            `json:"total"`
	Critical          int            `json:"critical"`
	High              int            `json:"high"`
	Medium            int            `json:"medium"`
	Low               int            `json:"low"`
	Suppressed        int            `json:"suppressed"`
	StaleSuppressions int            `json:"stale_suppressions"`
	New               int            `json:"new"`
	Existing          int            `json:"existing"`
	Resolved          int            `json:"resolved"`
	ByDomain          map[Domain]int `json:"by_domain"`
}

type ScannerState string

const (
	ScannerClean    ScannerState = "clean"
	ScannerFindings ScannerState = "findings"
	ScannerPartial  ScannerState = "partial"
	ScannerSkipped  ScannerState = "skipped"
	ScannerFailed   ScannerState = "failed"
)

func (s ScannerState) Valid() bool {
	return s == ScannerClean || s == ScannerFindings || s == ScannerPartial || s == ScannerSkipped || s == ScannerFailed
}

type ScannerStatus struct {
	ID             string        `json:"id"`
	State          ScannerState  `json:"state"`
	Required       bool          `json:"required"`
	Duration       time.Duration `json:"duration_ns"`
	Message        string        `json:"message,omitempty"`
	Version        string        `json:"version,omitempty"`
	Domain         Domain        `json:"domain,omitempty"`
	Capabilities   []string      `json:"capabilities,omitempty"`
	SupportedModes []string      `json:"supported_modes,omitempty"`
	FailureKind    string        `json:"failure_kind,omitempty"`
}

type Report struct {
	SchemaVersion         string          `json:"schema_version"`
	FingerprintVersion    string          `json:"fingerprint_version"`
	ToolVersion           string          `json:"tool_version,omitempty"`
	ConfigHash            string          `json:"config_hash"`
	RuleSetHash           string          `json:"rule_set_hash"`
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
