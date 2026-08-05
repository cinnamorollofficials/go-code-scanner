package fixer

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
)

type Change struct {
	File   string
	RuleID string
	Line   int
}

func Apply(root string, findings []finding.Finding, dryRun bool) ([]Change, error) {
	byFile := make(map[string][]finding.Finding)
	for _, item := range findings {
		if item.Fixable && item.RuleID == "trailing-whitespace" {
			byFile[item.Location.File] = append(byFile[item.Location.File], item)
		}
	}
	paths := make([]string, 0, len(byFile))
	for path := range byFile {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	var changes []Change
	for _, relative := range paths {
		path, err := safePath(root, relative)
		if err != nil {
			return nil, err
		}
		info, err := os.Lstat(path)
		if err != nil {
			return nil, fmt.Errorf("inspect fixable file %s: %w", relative, err)
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("refusing to fix non-regular file %s", relative)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read fixable file %s: %w", relative, err)
		}
		updated := append([]byte(nil), data...)
		items := append([]finding.Finding(nil), byFile[relative]...)
		sort.Slice(items, func(i, j int) bool { return items[i].Location.Line > items[j].Location.Line })
		for _, item := range items {
			var changed bool
			updated, changed = trimLine(updated, item.Location.Line)
			if changed {
				changes = append(changes, Change{File: relative, RuleID: item.RuleID, Line: item.Location.Line})
			}
		}
		if dryRun || bytes.Equal(data, updated) {
			continue
		}
		if err := writeAtomic(path, updated, info.Mode().Perm()); err != nil {
			return nil, err
		}
	}
	return changes, nil
}

func safePath(root, relative string) (string, error) {
	if filepath.IsAbs(relative) {
		return "", fmt.Errorf("refusing to fix absolute path %s", relative)
	}
	path := filepath.Clean(filepath.Join(root, filepath.FromSlash(relative)))
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("refusing to fix path outside root %s", relative)
	}
	return path, nil
}

func trimLine(data []byte, lineNumber int) ([]byte, bool) {
	if lineNumber < 1 {
		return data, false
	}
	lines := bytes.SplitAfter(data, []byte{'\n'})
	if lineNumber > len(lines) {
		return data, false
	}
	line := lines[lineNumber-1]
	ending := []byte(nil)
	if bytes.HasSuffix(line, []byte{'\n'}) {
		line, ending = line[:len(line)-1], []byte{'\n'}
	}
	if bytes.HasSuffix(line, []byte{'\r'}) {
		line, ending = line[:len(line)-1], []byte{'\r', '\n'}
	}
	trimmed := bytes.TrimRight(line, " \t")
	if len(trimmed) == len(line) {
		return data, false
	}
	lines[lineNumber-1] = append(append([]byte(nil), trimmed...), ending...)
	return bytes.Join(lines, nil), true
}

func writeAtomic(path string, data []byte, mode os.FileMode) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".security-review-fix-*")
	if err != nil {
		return fmt.Errorf("create fix file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(mode); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("install fixed file: %w", err)
	}
	return nil
}
