package reporter

import (
	"fmt"
	"io"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

func WriteTerminal(writer io.Writer, report *finding.Report) error {
	if _, err := fmt.Fprintf(writer, "Security review: %s (%s)\n", report.Project, report.ScanMode); err != nil {
		return err
	}
	for _, status := range report.Scanners {
		if _, err := fmt.Fprintf(writer, "  scanner %-16s %s", status.ID, status.State); err != nil {
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
	}
	_, err := fmt.Fprintf(writer,
		"Findings: %d | critical=%d high=%d medium=%d low=%d | suppressed=%d stale=%d\n",
		report.Summary.Total, report.Summary.Critical, report.Summary.High,
		report.Summary.Medium, report.Summary.Low, report.Summary.Suppressed,
		report.Summary.StaleSuppressions,
	)
	return err
}
