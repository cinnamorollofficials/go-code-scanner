package reporter

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

func TestWriteJSONReplacesExistingReport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.json")
	if err := os.WriteFile(path, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	report := &finding.Report{SchemaVersion: "1.0", Timestamp: time.Unix(0, 0).UTC()}
	if err := WriteJSON(path, report); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) == "old" {
		t.Fatal("report was not replaced")
	}
	if _, err := os.Stat(path + ".previous"); !os.IsNotExist(err) {
		t.Fatal("backup should be removed after successful replacement")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected report mode 0600, got %o", info.Mode().Perm())
	}
}
