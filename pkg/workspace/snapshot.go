package workspace

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/gitrepo"
)

const (
	DefaultMaxFiles int64 = 100_000
	DefaultMaxBytes int64 = 2 << 30
)

type Limits struct {
	MaxFiles int64
	MaxBytes int64
}

func DefaultLimits() Limits {
	return Limits{MaxFiles: DefaultMaxFiles, MaxBytes: DefaultMaxBytes}
}

// Snapshot is a temporary worktree materialized from the Git index.
type Snapshot struct {
	root   string
	closed bool
}

// MaterializeIndex exports the complete Git index without copying unstaged
// working-tree content or repository metadata.
func MaterializeIndex(ctx context.Context, repository *gitrepo.Repository, limits Limits) (*Snapshot, error) {
	if repository == nil {
		return nil, fmt.Errorf("repository is required")
	}
	if limits.MaxFiles < 1 {
		return nil, fmt.Errorf("max files must be at least 1")
	}
	if limits.MaxBytes < 1 {
		return nil, fmt.Errorf("max bytes must be at least 1")
	}

	root, err := os.MkdirTemp("", "security-review-index-*")
	if err != nil {
		return nil, fmt.Errorf("create staged workspace: %w", err)
	}
	snapshot := &Snapshot{root: root}
	failed := true
	defer func() {
		if failed {
			_ = snapshot.Close()
		}
	}()

	prefix := root + string(filepath.Separator)
	command := exec.CommandContext(ctx, "git", "-C", repository.Root(), "checkout-index", "--all", "--force", "--prefix="+prefix)
	if output, commandErr := command.CombinedOutput(); commandErr != nil {
		return nil, fmt.Errorf("materialize Git index: %w: %s", commandErr, strings.TrimSpace(string(output)))
	}
	if err := validateTree(root, limits); err != nil {
		return nil, err
	}
	failed = false
	return snapshot, nil
}

func (s *Snapshot) Root() string {
	return s.root
}

// Close removes the staged workspace. It is safe to call more than once.
func (s *Snapshot) Close() error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	if err := os.RemoveAll(s.root); err != nil {
		return fmt.Errorf("remove staged workspace: %w", err)
	}
	return nil
}

func validateTree(root string, limits Limits) error {
	var files, bytes int64
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root || entry.IsDir() {
			return nil
		}
		files++
		if files > limits.MaxFiles {
			return fmt.Errorf("staged workspace exceeds file limit %d", limits.MaxFiles)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			bytes += info.Size()
			if bytes > limits.MaxBytes {
				return fmt.Errorf("staged workspace exceeds size limit %d bytes", limits.MaxBytes)
			}
		}
		if info.Mode()&os.ModeSymlink != 0 {
			if err := validateSymlink(root, path); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("validate staged workspace: %w", err)
	}
	return nil
}

func validateSymlink(root, path string) error {
	target, err := os.Readlink(path)
	if err != nil {
		return err
	}
	if filepath.IsAbs(target) {
		return fmt.Errorf("symlink %s has absolute target %s", relativePath(root, path), target)
	}
	candidate := filepath.Clean(filepath.Join(filepath.Dir(path), target))
	if resolved, resolveErr := filepath.EvalSymlinks(candidate); resolveErr == nil {
		candidate = resolved
	}
	relative, err := filepath.Rel(root, candidate)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("symlink %s escapes staged workspace", relativePath(root, path))
	}
	return nil
}

func relativePath(root, path string) string {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(relative)
}
