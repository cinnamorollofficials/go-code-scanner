package gitrepo

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Repository provides the Git paths needed by hook and workspace operations.
type Repository struct {
	root string
}

// Open resolves the repository containing start.
func Open(ctx context.Context, start string) (*Repository, error) {
	if strings.TrimSpace(start) == "" {
		start = "."
	}
	output, err := exec.CommandContext(ctx, "git", "-C", start, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return nil, fmt.Errorf("resolve Git repository from %s: %w", start, err)
	}
	root := strings.TrimSpace(string(output))
	if root == "" {
		return nil, fmt.Errorf("resolve Git repository from %s: empty root", start)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve Git root: %w", err)
	}
	return &Repository{root: filepath.Clean(absRoot)}, nil
}

func (r *Repository) Root() string {
	return r.root
}

// GitPath resolves a repository-specific path while respecting worktrees and
// configuration such as core.hooksPath.
func (r *Repository) GitPath(ctx context.Context, name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("Git path name is required")
	}
	output, err := exec.CommandContext(ctx, "git", "-C", r.root, "rev-parse", "--git-path", name).Output()
	if err != nil {
		return "", fmt.Errorf("resolve Git path %s: %w", name, err)
	}
	path := strings.TrimSpace(string(output))
	if path == "" {
		return "", fmt.Errorf("resolve Git path %s: empty path", name)
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(r.root, path)
	}
	return filepath.Clean(path), nil
}

func (r *Repository) HooksDir(ctx context.Context) (string, error) {
	return r.GitPath(ctx, "hooks")
}
