package release

import (
	"encoding/json"
	"os"
	"path/filepath"
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
