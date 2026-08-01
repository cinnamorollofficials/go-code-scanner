package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/scanner"
)

var validKey = regexp.MustCompile(`^v[0-9]+-[a-f0-9]{64}$`)

type Entry struct {
	CreatedAt time.Time      `json:"created_at"`
	Result    scanner.Result `json:"result"`
}

type Store struct {
	Directory string
	Now       func() time.Time
}

func (s Store) Get(key string) (scanner.Result, bool, error) {
	path, err := s.entryPath(key)
	if err != nil {
		return scanner.Result{}, false, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return scanner.Result{}, false, nil
	}
	if err != nil {
		return scanner.Result{}, false, fmt.Errorf("read cache entry: %w", err)
	}
	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		return scanner.Result{}, false, fmt.Errorf("decode cache entry: %w", err)
	}
	return entry.Result, true, nil
}

func (s Store) Put(key string, result scanner.Result) error {
	path, err := s.entryPath(key)
	if err != nil {
		return err
	}
	result = Sanitize(result)
	now := time.Now
	if s.Now != nil {
		now = s.Now
	}
	data, err := json.Marshal(Entry{CreatedAt: now().UTC(), Result: result})
	if err != nil {
		return fmt.Errorf("encode cache entry: %w", err)
	}
	if err := os.MkdirAll(s.Directory, 0o700); err != nil {
		return fmt.Errorf("create cache directory: %w", err)
	}
	temporary, err := os.CreateTemp(s.Directory, ".entry-*")
	if err != nil {
		return fmt.Errorf("create cache entry: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace cache entry: %w", err)
	}
	return nil
}

// Sanitize returns a cache-safe deep copy without source snippets.
func Sanitize(result scanner.Result) scanner.Result {
	result.Findings = append([]finding.Finding(nil), result.Findings...)
	for index := range result.Findings {
		result.Findings[index].Snippet = ""
	}
	return result
}

func (s Store) entryPath(key string) (string, error) {
	if !validKey.MatchString(key) {
		return "", fmt.Errorf("invalid cache key")
	}
	if s.Directory == "" {
		return "", fmt.Errorf("cache directory is required")
	}
	return filepath.Join(s.Directory, key+".json"), nil
}
