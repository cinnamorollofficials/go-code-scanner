package discovery

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cinnamorollofficials/go-code-scanner/config"
)

func Files(ctx context.Context, cfg config.Config) ([]string, error) {
	switch cfg.Mode {
	case config.ModeFull:
		return walk(ctx, cfg)
	case config.ModeChanged:
		return gitFiles(ctx, cfg, "diff", "--name-only", "--diff-filter=ACMR", "HEAD")
	case config.ModeStaged:
		return gitFiles(ctx, cfg, "diff", "--cached", "--name-only", "--diff-filter=ACMR")
	default:
		return nil, fmt.Errorf("unsupported mode %q", cfg.Mode)
	}
}

func walk(ctx context.Context, cfg config.Config) ([]string, error) {
	var files []string
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
		if entry.Type().IsRegular() && allowed(path, cfg) {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

func gitFiles(ctx context.Context, cfg config.Config, args ...string) ([]string, error) {
	cmdArgs := append([]string{"-C", cfg.Root}, args...)
	output, err := exec.CommandContext(ctx, "git", cmdArgs...).Output()
	if err != nil {
		return nil, fmt.Errorf("list git files: %w", err)
	}
	var files []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		path := filepath.Join(cfg.Root, filepath.FromSlash(strings.TrimSpace(scanner.Text())))
		if info, statErr := os.Stat(path); statErr == nil && info.Mode().IsRegular() && allowed(path, cfg) {
			files = append(files, path)
		}
	}
	sort.Strings(files)
	return files, scanner.Err()
}

func allowed(path string, cfg config.Config) bool {
	base := filepath.Base(path)
	if contains(cfg.ExcludeFiles, base) {
		return false
	}
	ext := strings.ToLower(filepath.Ext(path))
	return contains(cfg.IncludeExtensions, ext)
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}
