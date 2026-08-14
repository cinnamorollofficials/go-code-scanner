package sqltaint

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/scanner"
)

func TestSQLTaintScannerComprehensive(t *testing.T) {
	code := `package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
)

type ORM struct{}
func (o *ORM) Where(query string, args ...any) *ORM { return o }
func (o *ORM) Raw(query string, args ...any) *ORM { return o }

func unsafeConcatenation(db *sql.DB, id string) {
	query := "SELECT * FROM users WHERE id = " + id
	db.Query(query)
}

func unsafeSprintf(db *sql.DB, name string) {
	db.Query(fmt.Sprintf("SELECT * FROM users WHERE name = '%s'", name))
}

func unsafeIdentifier(db *sql.DB, tableName string) {
	db.Query(fmt.Sprintf("SELECT * FROM %s WHERE active = 1", tableName))
}

func unsafeORMWhere(orm *ORM, role string) {
	orm.Where(fmt.Sprintf("role = '%s'", role))
}

func unsafeBindMismatch(db *sql.DB, id string) {
	db.Query("SELECT * FROM users WHERE id = ? AND tenant_id = ?", id)
}

func unsafeListExpansion(db *sql.DB, ids []string) {
	query := fmt.Sprintf("SELECT * FROM users WHERE id IN (%s)", strings.Join(ids, ","))
	db.Query(query)
}

func unsafePreparedTemplate(db *sql.DB, filter string) {
	stmt, err := db.Prepare("SELECT * FROM users WHERE status = " + filter)
	_ = stmt
	_ = err
}

func httpHandlerSink(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	userID := r.URL.Query().Get("id")
	db.Query(fmt.Sprintf("SELECT * FROM profiles WHERE user_id = '%s'", userID))
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

	foundRules := make(map[string]int)
	for _, f := range res.Findings {
		foundRules[f.RuleID]++
		if f.FindingState != finding.FindingConfirmed {
			t.Errorf("expected FindingConfirmed for %s, got %v", f.RuleID, f.FindingState)
		}
	}

	// SQLI-001 (unsafeConcatenation + unsafeSprintf + httpHandlerSink) -> 3
	if got := foundRules["SQLI-001"]; got != 3 {
		t.Errorf("got %d SQLI-001 findings, want 3", got)
	}

	// SQLI-002 (unsafeIdentifier) -> 1
	if got := foundRules["SQLI-002"]; got != 1 {
		t.Errorf("got %d SQLI-002 findings, want 1", got)
	}

	// SQLI-004 (unsafeORMWhere) -> 1
	if got := foundRules["SQLI-004"]; got != 1 {
		t.Errorf("got %d SQLI-004 findings, want 1", got)
	}

	// SQLI-008 (unsafeBindMismatch) -> 1
	if got := foundRules["SQLI-008"]; got != 1 {
		t.Errorf("got %d SQLI-008 findings, want 1", got)
	}

	// SQLI-011 (unsafeListExpansion) -> 1
	if got := foundRules["SQLI-011"]; got != 1 {
		t.Errorf("got %d SQLI-011 findings, want 1", got)
	}

	// SQLI-012 (unsafePreparedTemplate) -> 1
	if got := foundRules["SQLI-012"]; got != 1 {
		t.Errorf("got %d SQLI-012 findings, want 1", got)
	}

	// SQLSAFE-001 (unsafeDelete) -> 1
	if got := foundRules["SQLSAFE-001"]; got != 1 {
		t.Errorf("got %d SQLSAFE-001 findings, want 1", got)
	}
}
