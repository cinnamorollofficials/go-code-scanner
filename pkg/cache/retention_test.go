package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/scanner"
)

func TestPruneEnforcesAgeAndSizeWithoutTouchingForeignFiles(t *testing.T) {
	directory := t.TempDir()
	store := Store{Directory: directory}
	keys := make([]string, 3)
	for index := range keys {
		keys[index], _ = Key(KeyInput{ScannerID: "scanner", ScannerVersion: string(rune('1' + index))})
		if err := store.Put(keys[index], scanner.Result{Message: "entry"}); err != nil {
			t.Fatal(err)
		}
		stamp := time.Now().Add(time.Duration(index-3) * time.Hour)
		if err := os.Chtimes(filepath.Join(directory, keys[index]+".json"), stamp, stamp); err != nil {
			t.Fatal(err)
		}
	}
	foreign := filepath.Join(directory, "README")
	if err := os.WriteFile(foreign, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	if removed, err := store.Prune(2*time.Hour, 1); err != nil || removed != 3 {
		t.Fatalf("unexpected prune: removed=%d err=%v", removed, err)
	}
	if _, err := os.Stat(foreign); err != nil {
		t.Fatal("prune removed foreign file")
	}
	stats, err := store.Stats()
	if err != nil || stats.Entries != 0 || stats.Bytes != 0 {
		t.Fatalf("unexpected stats: %+v err=%v", stats, err)
	}
}
