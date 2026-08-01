package reporter

import (
	"encoding/json"
	"fmt"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

func WriteJSON(path string, report *finding.Report) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("encode report: %w", err)
	}
	return writeAtomic(path, append(data, '\n'), "report")
}
