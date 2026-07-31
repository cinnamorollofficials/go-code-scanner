package config

import "testing"

func TestDefaultConfigIsValid(t *testing.T) {
	cfg := Default()
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	if !filepathIsAbsolute(cfg.Root) {
		t.Fatalf("expected absolute root, got %q", cfg.Root)
	}
}

func TestConfigRejectsInvalidWorkerCount(t *testing.T) {
	cfg := Default()
	cfg.Workers = 0
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected worker validation error")
	}
}

func filepathIsAbsolute(path string) bool {
	if len(path) >= 3 && path[1] == ':' {
		return true
	}
	return len(path) > 0 && path[0] == '/'
}
