package cache

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const defaultStaleLockAge = 5 * time.Minute

// WithLock serializes work for a cache key across processes. Lock acquisition
// is cancellation-aware and abandoned locks are recovered after staleAge.
func (s Store) WithLock(ctx context.Context, key string, staleAge time.Duration, work func() error) error {
	if work == nil {
		return fmt.Errorf("cache lock work is required")
	}
	if _, err := s.entryPath(key); err != nil {
		return err
	}
	if staleAge <= 0 {
		staleAge = defaultStaleLockAge
	}
	if err := os.MkdirAll(s.Directory, 0o700); err != nil {
		return fmt.Errorf("create cache directory: %w", err)
	}
	lockPath := filepath.Join(s.Directory, key+".lock")
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		err := os.Mkdir(lockPath, 0o700)
		if err == nil {
			defer os.Remove(lockPath)
			return work()
		}
		if !errors.Is(err, os.ErrExist) {
			return fmt.Errorf("acquire cache lock: %w", err)
		}
		if info, statErr := os.Stat(lockPath); statErr == nil && time.Since(info.ModTime()) > staleAge {
			_ = os.Remove(lockPath)
			continue
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("acquire cache lock: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}
