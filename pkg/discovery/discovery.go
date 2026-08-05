package discovery

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/config"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/scanner"
)

func Sources(ctx context.Context, cfg config.Config) ([]scanner.Source, error) {
	return discover(ctx, cfg, false)
}

// Files returns every repository file eligible for metadata checks, including
// extensionless manifests and artifact formats that must not be regex-scanned.
func Files(ctx context.Context, cfg config.Config) ([]scanner.Source, error) {
	return discover(ctx, cfg, true)
}

// RepositoryFiles returns the complete repository/index inventory for
// governance checks that must not mistake an unchanged file for a missing one.
func RepositoryFiles(ctx context.Context, cfg config.Config) ([]scanner.Source, error) {
	if cfg.Mode == config.ModeFull {
		return Files(ctx, cfg)
	}
	return gitSources(ctx, cfg, true, true, "ls-files", "-z")
}

func discover(ctx context.Context, cfg config.Config, allFiles bool) ([]scanner.Source, error) {
	switch cfg.Mode {
	case config.ModeFull:
		return walk(ctx, cfg, allFiles)
	case config.ModeChanged:
		if !hasHEAD(ctx, cfg.Root) {
			return gitSources(ctx, cfg, true, allFiles, "diff", "--cached", "--name-only", "-z", "--diff-filter=ACMR")
		}
		return gitSources(ctx, cfg, false, allFiles, "diff", "--name-only", "-z", "--diff-filter=ACMR", "HEAD")
	case config.ModeStaged:
		return gitSources(ctx, cfg, true, allFiles, "diff", "--cached", "--name-only", "-z", "--diff-filter=ACMR")
	default:
		return nil, fmt.Errorf("unsupported mode %q", cfg.Mode)
	}
}

func hasHEAD(ctx context.Context, root string) bool {
	return exec.CommandContext(ctx, "git", "-C", root, "rev-parse", "--verify", "HEAD").Run() == nil
}

func walk(ctx context.Context, cfg config.Config, allFiles bool) ([]scanner.Source, error) {
	var sources []scanner.Source
	err := filepath.WalkDir(cfg.Root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if entry.IsDir() && path != cfg.Root && contains(cfg.ExcludeDirectories, entry.Name()) {
			return filepath.SkipDir
		}
		if entry.Type().IsRegular() && pathAllowed(path, cfg, allFiles) && (allFiles || allowed(path, cfg)) {
			sources = append(sources, fileSource(path))
		}
		return nil
	})
	sort.Slice(sources, func(i, j int) bool { return sources[i].Path < sources[j].Path })
	return sources, err
}

func gitSources(ctx context.Context, cfg config.Config, staged, allFiles bool, args ...string) ([]scanner.Source, error) {
	cmdArgs := append([]string{"-C", cfg.Root}, args...)
	output, err := exec.CommandContext(ctx, "git", cmdArgs...).Output()
	if err != nil {
		return nil, fmt.Errorf("list git files: %w", err)
	}
	var sources []scanner.Source
	for _, name := range bytes.Split(output, []byte{0}) {
		if len(name) == 0 {
			continue
		}
		relative := filepath.ToSlash(string(name))
		path := filepath.Join(cfg.Root, filepath.FromSlash(relative))
		if !pathAllowed(path, cfg, allFiles) || (!allFiles && !allowed(path, cfg)) {
			continue
		}
		if staged {
			sources = append(sources, stagedSource(cfg.Root, relative))
			continue
		}
		// Lstat deliberately rejects working-tree symlinks. Staged sources are
		// read from the Git object and therefore never follow the link target.
		if info, statErr := os.Lstat(path); statErr == nil && info.Mode().IsRegular() {
			sources = append(sources, fileSource(path))
		}
	}
	sort.Slice(sources, func(i, j int) bool { return sources[i].Path < sources[j].Path })
	return sources, nil
}

func fileSource(path string) scanner.Source {
	return scanner.Source{Path: path, Open: func(context.Context) (io.ReadCloser, error) {
		return os.Open(path)
	}}
}

func stagedSource(root, relative string) scanner.Source {
	path := filepath.Join(root, filepath.FromSlash(relative))
	return scanner.Source{Path: path, Open: func(ctx context.Context) (io.ReadCloser, error) {
		output, err := exec.CommandContext(ctx, "git", "-C", root, "show", ":"+relative).Output()
		if err != nil {
			return nil, fmt.Errorf("read staged file %s: %w", relative, err)
		}
		return io.NopCloser(bytes.NewReader(output)), nil
	}}
}

func allowed(path string, cfg config.Config) bool {
	base := filepath.Base(path)
	if contains(cfg.ExcludeFiles, base) {
		return false
	}
	ext := strings.ToLower(filepath.Ext(path))
	if contains(cfg.IncludeExtensions, ext) {
		return true
	}
	if cfg.Frontend.Enabled && contains(cfg.Frontend.IncludeExtensions, ext) {
		return true
	}
	return false
}

func pathAllowed(path string, cfg config.Config, includeExcludedFiles bool) bool {
	relative, err := filepath.Rel(cfg.Root, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return false
	}
	if cfg.Cache.Enabled {
		cacheDirectory := cfg.Cache.Directory
		if !filepath.IsAbs(cacheDirectory) {
			cacheDirectory = filepath.Join(cfg.Root, cacheDirectory)
		}
		cacheRelative, cacheErr := filepath.Rel(cacheDirectory, path)
		if cacheErr == nil && cacheRelative != ".." && !strings.HasPrefix(cacheRelative, ".."+string(filepath.Separator)) {
			return false
		}
	}
	for _, part := range strings.Split(filepath.Clean(relative), string(filepath.Separator)) {
		if contains(cfg.ExcludeDirectories, part) {
			return false
		}
	}
	return includeExcludedFiles || !contains(cfg.ExcludeFiles, filepath.Base(path))
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}
