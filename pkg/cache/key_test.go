package cache

import "testing"

func TestKeyIsDeterministicAndInvalidatesEveryInput(t *testing.T) {
	base := KeyInput{ScannerID: "pattern", ScannerVersion: "1", ConfigHash: "config", RuleSetHash: "rules", Files: map[string]string{"b.go": "b", "a.go": "a"}}
	want, err := Key(base)
	if err != nil {
		t.Fatal(err)
	}
	reordered := base
	reordered.Files = map[string]string{"a.go": "a", "b.go": "b"}
	if got, _ := Key(reordered); got != want {
		t.Fatalf("map order changed cache key: %s != %s", got, want)
	}
	mutations := []KeyInput{
		{ScannerID: "other", ScannerVersion: "1", ConfigHash: "config", RuleSetHash: "rules", Files: base.Files},
		{ScannerID: "pattern", ScannerVersion: "2", ConfigHash: "config", RuleSetHash: "rules", Files: base.Files},
		{ScannerID: "pattern", ScannerVersion: "1", ConfigHash: "other", RuleSetHash: "rules", Files: base.Files},
		{ScannerID: "pattern", ScannerVersion: "1", ConfigHash: "config", RuleSetHash: "other", Files: base.Files},
		{ScannerID: "pattern", ScannerVersion: "1", ConfigHash: "config", RuleSetHash: "rules", Files: map[string]string{"a.go": "changed", "b.go": "b"}},
	}
	for index, mutation := range mutations {
		if got, err := Key(mutation); err != nil || got == want {
			t.Fatalf("mutation %d did not invalidate key: got=%q err=%v", index, got, err)
		}
	}
}

func TestKeyRejectsIncompleteIdentity(t *testing.T) {
	if _, err := Key(KeyInput{}); err == nil {
		t.Fatal("incomplete cache identity accepted")
	}
}
