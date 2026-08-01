package release

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func VerifyChecksums(manifestPath, directory string) error {
	manifest, err := os.Open(manifestPath)
	if err != nil {
		return fmt.Errorf("open checksum manifest: %w", err)
	}
	defer manifest.Close()

	seen := make(map[string]struct{})
	scanner := bufio.NewScanner(manifest)
	line := 0
	for scanner.Scan() {
		line++
		text := scanner.Text()
		parts := strings.SplitN(text, "  ", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("checksum manifest line %d: expected '<sha256>  <filename>'", line)
		}
		digest, name := parts[0], parts[1]
		decoded, err := hex.DecodeString(digest)
		if err != nil || len(decoded) != sha256.Size {
			return fmt.Errorf("checksum manifest line %d: invalid SHA-256", line)
		}
		if filepath.Base(name) != name || name == "." || name == ".." {
			return fmt.Errorf("checksum manifest line %d: filename must be a basename", line)
		}
		if _, exists := seen[name]; exists {
			return fmt.Errorf("checksum manifest line %d: duplicate filename %q", line, name)
		}
		seen[name] = struct{}{}
		if err := verifyChecksumFile(filepath.Join(directory, name), digest); err != nil {
			return fmt.Errorf("verify %s: %w", name, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read checksum manifest: %w", err)
	}
	if len(seen) == 0 {
		return fmt.Errorf("checksum manifest contains no artifacts")
	}
	return nil
}

func verifyChecksumFile(path, expected string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("artifact must be a regular file")
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}
	actual := hex.EncodeToString(hash.Sum(nil))
	if actual != strings.ToLower(expected) {
		return fmt.Errorf("checksum mismatch")
	}
	return nil
}
