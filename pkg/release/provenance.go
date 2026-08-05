package release

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const ProvenanceSchema = "go-code-scanner/provenance/v1"

type ProvenanceOptions struct {
	Version   string
	Commit    string
	BuildDate time.Time
	Builder   string
}

type Subject struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
}

type Provenance struct {
	Schema    string    `json:"schema"`
	Version   string    `json:"version"`
	Commit    string    `json:"commit"`
	BuildDate time.Time `json:"build_date"`
	Builder   string    `json:"builder"`
	Subjects  []Subject `json:"subjects"`
}

func WriteProvenance(directory, output string, options ProvenanceOptions) error {
	if strings.TrimSpace(options.Version) == "" || strings.TrimSpace(options.Commit) == "" || options.BuildDate.IsZero() || strings.TrimSpace(options.Builder) == "" {
		return fmt.Errorf("version, commit, build date, and builder are required")
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("read release directory: %w", err)
	}
	var subjects []Subject
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == filepath.Base(output) || entry.Name() == "SHA256SUMS" || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		file, err := os.Open(filepath.Join(directory, entry.Name()))
		if err != nil {
			return err
		}
		digest := sha256.New()
		_, copyErr := io.Copy(digest, file)
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		subjects = append(subjects, Subject{Name: entry.Name(), SHA256: hex.EncodeToString(digest.Sum(nil))})
	}
	sort.Slice(subjects, func(i, j int) bool { return subjects[i].Name < subjects[j].Name })
	document := Provenance{Schema: ProvenanceSchema, Version: options.Version, Commit: options.Commit, BuildDate: options.BuildDate.UTC(), Builder: options.Builder, Subjects: subjects}
	data, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(output), ".provenance-*")
	if err != nil {
		return err
	}
	name := temporary.Name()
	defer os.Remove(name)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(append(data, '\n')); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(name, output)
}

func VerifyProvenance(path, directory string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read provenance: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var document Provenance
	if err := decoder.Decode(&document); err != nil {
		return fmt.Errorf("decode provenance: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("decode provenance: trailing JSON value")
	}
	if document.Schema != ProvenanceSchema {
		return fmt.Errorf("unsupported provenance schema %q", document.Schema)
	}
	if strings.TrimSpace(document.Version) == "" || strings.TrimSpace(document.Commit) == "" || document.BuildDate.IsZero() || strings.TrimSpace(document.Builder) == "" {
		return fmt.Errorf("provenance metadata is incomplete")
	}
	if len(document.Subjects) == 0 {
		return fmt.Errorf("provenance contains no subjects")
	}
	previous := ""
	for index, subject := range document.Subjects {
		if filepath.Base(subject.Name) != subject.Name || subject.Name == "." || subject.Name == ".." {
			return fmt.Errorf("subject %d has unsafe name %q", index, subject.Name)
		}
		if subject.Name <= previous {
			return fmt.Errorf("provenance subjects must be unique and sorted")
		}
		previous = subject.Name
		if err := verifyChecksumFile(filepath.Join(directory, subject.Name), subject.SHA256); err != nil {
			return fmt.Errorf("verify provenance subject %s: %w", subject.Name, err)
		}
	}
	return nil
}
