package sqltaint

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/scanner"
)

func TestSQLTaintScannerDetection(t *testing.T) {
	code := `package main

import (
	"database/sql"
	"fmt"
)

func unsafeConcatenation(db *sql.DB, id string) {
	query := "SELECT * FROM users WHERE id = " + id
	db.Query(query)
}

func unsafeSprintf(db *sql.DB, name string) {
	db.Query(fmt.Sprintf("SELECT * FROM users WHERE name = '%s'", name))
}

func safeParameterized(db *sql.DB, id string) {
	db.Query("SELECT * FROM users WHERE id = $1", id)
}

func unsafeDelete(db *sql.DB) {
	db.Exec("DELETE FROM users")
}
`

	s := New()
	if s.ID() != "sqltaint" {
		t.Fatalf("got ID %q, want 'sqltaint'", s.ID())
	}

	desc := s.Describe()
	if desc.Domain != finding.Security {
		t.Fatalf("got domain %v, want Security", desc.Domain)
	}

	req := scanner.Request{
		Sources: []scanner.Source{
			{
				Path: "app.go",
				Open: func(ctx context.Context) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader(code)), nil
				},
			},
		},
	}

	res := s.Scan(context.Background(), req)
	if res.State != finding.ScannerFindings {
		t.Fatalf("got state %v, want ScannerFindings", res.State)
	}

	if len(res.Findings) != 3 {
		t.Fatalf("got %d findings, want 3 (2 SQLI-001, 1 SQLSAFE-001)", len(res.Findings))
	}

	sqliCount := 0
	sqlsafeCount := 0

	for _, f := range res.Findings {
		if f.RuleID == "SQLI-001" {
			sqliCount++
			if f.Confidence != finding.ConfidenceHigh {
				t.Errorf("got confidence %v, want High", f.Confidence)
			}
			if len(f.Dataflow) == 0 {
				t.Error("expected non-empty dataflow trace for SQLI-001")
			}
		} else if f.RuleID == "SQLSAFE-001" {
			sqlsafeCount++
		}
	}

	if sqliCount != 2 {
		t.Errorf("got %d SQLI-001 findings, want 2", sqliCount)
	}
	if sqlsafeCount != 1 {
		t.Errorf("got %d SQLSAFE-001 findings, want 1", sqlsafeCount)
	}
}
