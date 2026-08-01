package cache

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/scanner"
)

func TestStoreRoundTripRedactsSnippetsAndUsesSafePermissions(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "cache")
	store := Store{Directory: directory, Now: func() time.Time { return time.Unix(0, 0) }}
	key, err := Key(KeyInput{ScannerID: "pattern", ScannerVersion: "1"})
	if err != nil {
		t.Fatal(err)
	}
	result := scanner.Result{State: finding.ScannerFindings, Findings: []finding.Finding{{RuleID: "secret", Snippet: "TOP-SECRET"}}}
	if err := store.Put(key, result); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(directory, key+".json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "TOP-SECRET") {
		t.Fatal("cache persisted a source snippet")
	}
	info, _ := os.Stat(filepath.Join(directory, key+".json"))
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("unexpected cache permissions: %o", info.Mode().Perm())
	}
	loaded, found, err := store.Get(key)
	if err != nil || !found || loaded.State != finding.ScannerFindings || loaded.Findings[0].Snippet != "" {
		t.Fatalf("unexpected cache round trip: result=%+v found=%t err=%v", loaded, found, err)
	}
}

func TestStoreRejectsTraversalKey(t *testing.T) {
	store := Store{Directory: t.TempDir()}
	if err := store.Put("../../escape", scanner.Result{}); err == nil {
		t.Fatal("unsafe cache key accepted")
	}
}
