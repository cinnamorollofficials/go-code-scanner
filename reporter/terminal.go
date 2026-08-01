package reporter

import (
	"fmt"
	"io"
	"sort"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

const DefaultTerminalFindingLimit = 50

type TerminalOptions struct {
	MaxFindings     int
	Verbose         bool
	Color           bool
	Command         string
	SuppressionFile string
}

func WriteTerminal(writer io.Writer, report *finding.Report) error {
	return WriteTerminalWithOptions(writer, report, TerminalOptions{MaxFindings: DefaultTerminalFindingLimit})
}

func WriteTerminalWithOptions(writer io.Writer, report *finding.Report, options TerminalOptions) error {
	if report == nil {
		return fmt.Errorf("report is required")
	}
	if options.MaxFindings < 0 {
		return fmt.Errorf("max findings cannot be negative")
	}
	if options.Command == "" {
		options.Command = "security-review"
	}
	if options.SuppressionFile == "" {
		options.SuppressionFile = ".security-ignore"
	}
	if _, err := fmt.Fprintf(writer, "Code review: %s (%s)\n", report.Project, report.ScanMode); err != nil {
		return err
	}
	for _, status := range report.Scanners {
		state := string(status.State)
		if options.Color {
			state = colorState(status.State, state)
		}
		if _, err := fmt.Fprintf(writer, "  scanner %-16s %s", status.ID, state); err != nil {
			return err
		}
		if status.Message != "" {
			if _, err := fmt.Fprintf(writer, " (%s)", status.Message); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(writer); err != nil {
			return err
		}
		if options.Verbose {
			if _, err := fmt.Fprintf(writer, "    required=%t duration=%s", status.Required, status.Duration); err != nil {
				return err
			}
			if status.Version != "" {
				_, _ = fmt.Fprintf(writer, " version=%s", status.Version)
			}
			if status.Domain != "" {
				_, _ = fmt.Fprintf(writer, " domain=%s", status.Domain)
			}
			if status.FailureKind != "" {
				_, _ = fmt.Fprintf(writer, " failure=%s", status.FailureKind)
			}
			if len(status.Capabilities) > 0 {
				_, _ = fmt.Fprintf(writer, " capabilities=%v", status.Capabilities)
			}
			if _, err := fmt.Fprintln(writer); err != nil {
				return err
			}
		}
	}
	if _, err := fmt.Fprintf(writer,
		"Findings: %d | critical=%d high=%d medium=%d low=%d | suppressed=%d stale=%d\n",
		report.Summary.Total, report.Summary.Critical, report.Summary.High,
		report.Summary.Medium, report.Summary.Low, report.Summary.Suppressed,
		report.Summary.StaleSuppressions,
	); err != nil {
		return err
	}

	items := append([]finding.Finding(nil), report.Findings...)
	sort.SliceStable(items, func(i, j int) bool {
		if baselineRank(items[i].BaselineState) != baselineRank(items[j].BaselineState) {
			return baselineRank(items[i].BaselineState) < baselineRank(items[j].BaselineState)
		}
		if items[i].Severity.Rank() != items[j].Severity.Rank() {
			return items[i].Severity.Rank() < items[j].Severity.Rank()
		}
		if items[i].Location.File != items[j].Location.File {
			return items[i].Location.File < items[j].Location.File
		}
		return items[i].Location.Line < items[j].Location.Line
	})
	limit := len(items)
	if options.MaxFindings > 0 && limit > options.MaxFindings {
		limit = options.MaxFindings
	}
	for _, item := range items[:limit] {
		state := ""
		if item.BaselineState != "" {
			state = " [" + string(item.BaselineState) + "]"
		}
		severity := string(item.Severity)
		if options.Color {
			severity = colorSeverity(item.Severity, severity)
		}
		if _, err := fmt.Fprintf(writer, "\n[%s]%s %s/%s\n", severity, state, item.Domain, item.RuleID); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(writer, "  %s:%d %s\n", item.Location.File, item.Location.Line, item.Description); err != nil {
			return err
		}
		if item.Recommendation != "" {
			if _, err := fmt.Fprintf(writer, "  Fix: %s\n", item.Recommendation); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(writer, "  Explain: %s --explain %s\n", options.Command, item.RuleID); err != nil {
			return err
		}
		if item.Fixable {
			if _, err := fmt.Fprintf(writer, "  Fix: %s --fix\n", options.Command); err != nil {
				return err
			}
		}
		if item.Fingerprint != "" {
			if _, err := fmt.Fprintf(writer, "  Suppress: add fingerprint %s with reason and expiry to %s\n", item.Fingerprint, options.SuppressionFile); err != nil {
				return err
			}
		}
	}
	if hidden := len(items) - limit; hidden > 0 {
		if _, err := fmt.Fprintf(writer, "\n... %d additional findings omitted\n", hidden); err != nil {
			return err
		}
	}
	if len(report.Warnings) > 0 {
		if _, err := fmt.Fprintln(writer, "\nWarnings:"); err != nil {
			return err
		}
		for _, warning := range report.Warnings {
			if _, err := fmt.Fprintf(writer, "  - %s\n", warning); err != nil {
				return err
			}
		}
	}
	return nil
}

func colorSeverity(severity finding.Severity, value string) string {
	code := "36"
	switch severity {
	case finding.Critical:
		code = "1;31"
	case finding.High:
		code = "31"
	case finding.Medium:
		code = "33"
	}
	return "\x1b[" + code + "m" + value + "\x1b[0m"
}

func colorState(state finding.ScannerState, value string) string {
	code := "36"
	switch state {
	case finding.ScannerClean:
		code = "32"
	case finding.ScannerFailed:
		code = "31"
	case finding.ScannerPartial:
		code = "33"
	case finding.ScannerSkipped:
		code = "2"
	}
	return "\x1b[" + code + "m" + value + "\x1b[0m"
}

func baselineRank(state finding.BaselineState) int {
	switch state {
	case finding.BaselineNew:
		return 0
	case "":
		return 1
	case finding.BaselineExisting:
		return 2
	default:
		return 3
	}
}
