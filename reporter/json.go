package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

func WriteJSON(path string, report *finding.Report) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("encode report: %w", err)
	}
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	temporary, err := os.CreateTemp(directory, ".security-review-*.json")
	if err != nil {
		return fmt.Errorf("create temporary report: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err := temporary.Write(append(data, '\n')); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return fmt.Errorf("sync temporary report: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := replace(path, temporaryPath); err != nil {
		return err
	}
	return nil
}

func replace(path, temporaryPath string) error {
	backup := path + ".previous"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Rename(temporaryPath, path); err != nil {
			return fmt.Errorf("install report: %w", err)
		}
		return nil
	}
	_ = os.Remove(backup)
	if err := os.Rename(path, backup); err != nil {
		return fmt.Errorf("backup existing report: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		if restoreErr := os.Rename(backup, path); restoreErr != nil {
			return fmt.Errorf("install report: %v; restore previous report: %w", err, restoreErr)
		}
		return fmt.Errorf("install report: %w", err)
	}
	if err := os.Remove(backup); err != nil {
		return fmt.Errorf("remove report backup: %w", err)
	}
	return nil
}
