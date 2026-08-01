package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Stats struct {
	Entries int
	Bytes   int64
}

type cacheFile struct {
	path     string
	size     int64
	modified time.Time
}

func (s Store) Stats() (Stats, error) {
	files, err := s.cacheFiles()
	if err != nil {
		return Stats{}, err
	}
	stats := Stats{Entries: len(files)}
	for _, file := range files {
		stats.Bytes += file.size
	}
	return stats, nil
}

// Prune removes expired entries first, then the oldest entries until maxBytes
// is satisfied. Non-cache files and active lock directories are untouched.
func (s Store) Prune(maxAge time.Duration, maxBytes int64) (int, error) {
	files, err := s.cacheFiles()
	if err != nil {
		return 0, err
	}
	now := time.Now()
	removed := 0
	var retained []cacheFile
	var total int64
	for _, file := range files {
		if maxAge > 0 && now.Sub(file.modified) > maxAge {
			if err := os.Remove(file.path); err != nil {
				return removed, fmt.Errorf("remove expired cache entry: %w", err)
			}
			removed++
			continue
		}
		retained = append(retained, file)
		total += file.size
	}
	sort.Slice(retained, func(i, j int) bool {
		if !retained[i].modified.Equal(retained[j].modified) {
			return retained[i].modified.Before(retained[j].modified)
		}
		return retained[i].path < retained[j].path
	})
	for _, file := range retained {
		if maxBytes <= 0 || total <= maxBytes {
			break
		}
		if err := os.Remove(file.path); err != nil {
			return removed, fmt.Errorf("remove oversized cache entry: %w", err)
		}
		total -= file.size
		removed++
	}
	return removed, nil
}

func (s Store) cacheFiles() ([]cacheFile, error) {
	entries, err := os.ReadDir(s.Directory)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read cache directory: %w", err)
	}
	var files []cacheFile
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") || !validKey.MatchString(strings.TrimSuffix(entry.Name(), ".json")) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		files = append(files, cacheFile{path: filepath.Join(s.Directory, entry.Name()), size: info.Size(), modified: info.ModTime()})
	}
	return files, nil
}
