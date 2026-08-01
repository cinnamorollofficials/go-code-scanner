package release

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWriteProvenanceHashesSortedArtifacts(t *testing.T) {
	directory := t.TempDir()
	for name, content := range map[string]string{"z-binary": "z", "a-binary": "a", "SHA256SUMS": "ignored"} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	output := filepath.Join(directory, "provenance.json")
	options := ProvenanceOptions{Version: "v1.2.3", Commit: "abc123", BuildDate: time.Unix(0, 0), Builder: "ci/example"}
	if err := WriteProvenance(directory, output, options); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(output)
	var document Provenance
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	if len(document.Subjects) != 2 || document.Subjects[0].Name != "a-binary" || document.Subjects[0].SHA256 == "" || document.Subjects[1].Name != "z-binary" {
		t.Fatalf("unexpected provenance subjects: %+v", document.Subjects)
	}
	info, _ := os.Stat(output)
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("unexpected provenance permissions: %o", info.Mode().Perm())
	}
}

func TestVerifyProvenanceSubjects(t *testing.T) {
	directory := t.TempDir()
	artifact := filepath.Join(directory, "security-review.tar.gz")
	if err := os.WriteFile(artifact, []byte("artifact"), 0o600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "provenance.json")
	options := ProvenanceOptions{Version: "v1.2.3", Commit: "abc123", BuildDate: time.Unix(1, 0), Builder: "ci/example"}
	if err := WriteProvenance(directory, path, options); err != nil {
		t.Fatal(err)
	}
	if err := VerifyProvenance(path, directory); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(artifact, []byte("tampered"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := VerifyProvenance(path, directory); err == nil || !strings.Contains(err.Error(), "mismatch") {
		t.Fatalf("expected provenance mismatch, got %v", err)
	}
}

func TestVerifyProvenanceRejectsUnknownFields(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "provenance.json")
	data := `{"schema":"go-code-scanner/provenance/v1","unknown":true}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := VerifyProvenance(path, directory); err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Fatalf("expected strict decode error, got %v", err)
	}
}

func TestProvenanceErrorsDoNotExposeArtifactContents(t *testing.T) {
	directory := t.TempDir()
	artifact := filepath.Join(directory, "release.tar.gz")
	const secret = "CANARY-SECRET-DO-NOT-LEAK"
	if err := os.WriteFile(artifact, []byte(secret), 0o600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "provenance.json")
	options := ProvenanceOptions{Version: "v1.2.3", Commit: "abc123", BuildDate: time.Unix(1, 0), Builder: "test"}
	if err := WriteProvenance(directory, path, options); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(artifact, []byte("tampered"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := VerifyProvenance(path, directory)
	if err == nil {
		t.Fatal("expected provenance mismatch")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("provenance error exposed artifact content: %v", err)
	}
}
