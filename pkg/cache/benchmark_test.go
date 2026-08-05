package cache

import (
	"fmt"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/scanner"
)

func BenchmarkKey(b *testing.B) {
	files := make(map[string]string, 1000)
	for index := range 1000 {
		files[fmt.Sprintf("internal/package%04d/file.go", index)] = fmt.Sprintf("hash-%04d", index)
	}
	input := KeyInput{ScannerID: "pattern", ScannerVersion: "1", ConfigHash: "config", RuleSetHash: "rules", Files: files}
	b.ReportAllocs()
	for b.Loop() {
		if _, err := Key(input); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStoreRoundTrip(b *testing.B) {
	store := Store{Directory: b.TempDir()}
	key, err := Key(KeyInput{ScannerID: "pattern", ScannerVersion: "1"})
	if err != nil {
		b.Fatal(err)
	}
	result := scanner.Result{Message: "cached scanner result"}
	b.ReportAllocs()
	for b.Loop() {
		if err := store.Put(key, result); err != nil {
			b.Fatal(err)
		}
		if _, found, err := store.Get(key); err != nil || !found {
			b.Fatalf("cache read failed: found=%t err=%v", found, err)
		}
	}
}
