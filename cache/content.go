package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cinnamorollofficials/go-code-scanner/scanner"
)

// HashSources returns normalized path-to-content hashes without retaining
// source bytes. Duplicate paths are read once and output is deterministic.
func HashSources(ctx context.Context, root string, groups ...[]scanner.Source) (map[string]string, error) {
	unique := make(map[string]scanner.Source)
	for _, sources := range groups {
		for _, source := range sources {
			relative, err := filepath.Rel(root, source.Path)
			if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
				return nil, fmt.Errorf("cache source %q escapes root", source.Path)
			}
			unique[filepath.ToSlash(relative)] = source
		}
	}
	paths := make([]string, 0, len(unique))
	for path := range unique {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	hashes := make(map[string]string, len(paths))
	for _, path := range paths {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		reader, err := unique[path].Open(ctx)
		if err != nil {
			return nil, fmt.Errorf("hash cache source %s: %w", path, err)
		}
		digest := sha256.New()
		_, copyErr := io.Copy(digest, reader)
		closeErr := reader.Close()
		if copyErr != nil {
			return nil, fmt.Errorf("hash cache source %s: %w", path, copyErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("close cache source %s: %w", path, closeErr)
		}
		hashes[path] = hex.EncodeToString(digest.Sum(nil))
	}
	return hashes, nil
}
