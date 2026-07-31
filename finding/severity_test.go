package finding

import "testing"

func TestParseSeverity(t *testing.T) {
	got, err := ParseSeverity(" high ")
	if err != nil {
		t.Fatal(err)
	}
	if got != High {
		t.Fatalf("got %q, want %q", got, High)
	}
	if !Critical.AtLeast(High) || Low.AtLeast(High) {
		t.Fatal("unexpected severity ordering")
	}
}

func TestParseSeverityRejectsUnknownValue(t *testing.T) {
	if _, err := ParseSeverity("urgent"); err == nil {
		t.Fatal("expected validation error")
	}
}
