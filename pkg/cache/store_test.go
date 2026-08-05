package cache

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/scanner"
)

func TestStoreRoundTripRedactsSnippetsAndUsesSafePermissions(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "cache")
	store := Store{Directory: directory, Now: func() time.Time { return time.Unix(0, 0) }}
	key, err := Key(KeyInput{ScannerID: "pattern", ScannerVersion: "1"})
	if err != nil {
		t.Fatal(err)
	}
	result := scanner.Result{State: finding.ScannerFindings, Findings: []finding.Finding{{RuleID: "secret", Snippet: "CANARY-SECRET-DO-NOT-LEAK"}}}
	if err := store.Put(key, result); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(directory, key+".json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "CANARY-SECRET-DO-NOT-LEAK") {
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

func TestStoreQuarantinesCorruptEntryAsCacheMiss(t *testing.T) {
	directory := t.TempDir()
	store := Store{Directory: directory}
	key, _ := Key(KeyInput{ScannerID: "pattern", ScannerVersion: "1"})
	path := filepath.Join(directory, key+".json")
	if err := os.WriteFile(path, []byte(`{"created_at":`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, found, err := store.Get(key); err != nil || found {
		t.Fatalf("corrupt entry was not treated as miss: found=%t err=%v", found, err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("corrupt entry remained active")
	}
	matches, err := filepath.Glob(filepath.Join(directory, ".corrupt-"+key+".json-*"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("corrupt entry was not quarantined: matches=%v err=%v", matches, err)
	}
	if err := store.Put(key, scanner.Result{State: finding.ScannerClean}); err != nil {
		t.Fatal(err)
	}
	if _, found, err := store.Get(key); err != nil || !found {
		t.Fatalf("healthy replacement was not readable: found=%t err=%v", found, err)
	}
}

func TestStoreRejectsSymlinkDirectory(t *testing.T) {
	outside := t.TempDir()
	link := filepath.Join(t.TempDir(), "cache")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	store := Store{Directory: link}
	key, _ := Key(KeyInput{ScannerID: "pattern", ScannerVersion: "1"})
	if err := store.Put(key, scanner.Result{}); err == nil {
		t.Fatal("cache store accepted symlink directory")
	}
	if _, err := store.Stats(); err == nil {
		t.Fatal("cache stats accepted symlink directory")
	}
	entries, err := os.ReadDir(outside)
	if err != nil || len(entries) != 0 {
		t.Fatalf("outside directory was modified: entries=%v err=%v", entries, err)
	}
}
