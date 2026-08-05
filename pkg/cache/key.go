package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

const KeyVersion = "1"

type KeyInput struct {
	ScannerID      string
	ScannerVersion string
	ConfigHash     string
	RuleSetHash    string
	Files          map[string]string
}

type keyFile struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
}

// Key returns a versioned deterministic cache identity. Files contains only
// normalized paths and content hashes; source content must never be supplied.
func Key(input KeyInput) (string, error) {
	if strings.TrimSpace(input.ScannerID) == "" || strings.TrimSpace(input.ScannerVersion) == "" {
		return "", fmt.Errorf("scanner ID and version are required")
	}
	paths := make([]string, 0, len(input.Files))
	for path, hash := range input.Files {
		if strings.TrimSpace(path) == "" || strings.TrimSpace(hash) == "" {
			return "", fmt.Errorf("cache file path and hash are required")
		}
		paths = append(paths, path)
	}
	sort.Strings(paths)
	files := make([]keyFile, 0, len(paths))
	for _, path := range paths {
		files = append(files, keyFile{Path: path, Hash: input.Files[path]})
	}
	payload := struct {
		Version        string    `json:"version"`
		ScannerID      string    `json:"scanner_id"`
		ScannerVersion string    `json:"scanner_version"`
		ConfigHash     string    `json:"config_hash"`
		RuleSetHash    string    `json:"rule_set_hash"`
		Files          []keyFile `json:"files"`
	}{KeyVersion, input.ScannerID, input.ScannerVersion, input.ConfigHash, input.RuleSetHash, files}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode cache key: %w", err)
	}
	digest := sha256.Sum256(encoded)
	return "v" + KeyVersion + "-" + hex.EncodeToString(digest[:]), nil
}
