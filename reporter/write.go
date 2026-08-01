package reporter

import (
	"fmt"
	"os"
	"path/filepath"
)

func writeAtomic(path string, data []byte, label string) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create %s directory: %w", label, err)
	}
	temporary, err := os.CreateTemp(directory, ".security-review-*")
	if err != nil {
		return fmt.Errorf("create temporary %s: %w", label, err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return fmt.Errorf("secure temporary %s: %w", label, err)
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return fmt.Errorf("write temporary %s: %w", label, err)
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return fmt.Errorf("sync temporary %s: %w", label, err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary %s: %w", label, err)
	}
	if err := replace(path, temporaryPath); err != nil {
		return err
	}
	return nil
}

func replace(path, temporaryPath string) error {
	backup := path + ".previous"
	if err := rejectSymlink(path); err != nil {
		return err
	}
	if err := rejectSymlink(backup); err != nil {
		return err
	}
	if _, err := os.Lstat(path); os.IsNotExist(err) {
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

func rejectSymlink(path string) error {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect report path: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to replace symlink %s", path)
	}
	return nil
}
