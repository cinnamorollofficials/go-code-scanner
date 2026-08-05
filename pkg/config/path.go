package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolveProjectPath resolves target under root and rejects lexical traversal
// and existing symlink parents that escape the project boundary.
func ResolveProjectPath(root, target string) (string, error) {
	if strings.TrimSpace(root) == "" || strings.TrimSpace(target) == "" {
		return "", fmt.Errorf("root and path are required")
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	resolved := target
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(absRoot, resolved)
	}
	resolved, err = filepath.Abs(resolved)
	if err != nil {
		return "", err
	}
	if !within(absRoot, resolved) {
		return "", fmt.Errorf("path %q escapes project root", target)
	}
	realRoot, err := filepath.EvalSymlinks(absRoot)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if realRoot == "" {
		realRoot = absRoot
	}
	realParent, err := filepath.EvalSymlinks(filepath.Dir(resolved))
	if err == nil && !within(realRoot, realParent) {
		return "", fmt.Errorf("path %q resolves outside project root", target)
	}
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	return resolved, nil
}

func within(root, target string) bool {
	relative, err := filepath.Rel(root, target)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
