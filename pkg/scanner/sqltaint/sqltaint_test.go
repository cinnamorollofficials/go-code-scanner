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

func TestInterproceduralAndRules(t *testing.T) {
	codeHandler := `package main

import (
	"database/sql"
	"net/http"
)

type GinContext struct{}
func (c *GinContext) Param(key string) string { return "123" }

func UserHandler(c *GinContext, repo *UserRepo) {
	id := c.Param("id")
	repo.FindUserUnsafe(id)
}
`

	codeRepo := `package main

import (
	"database/sql"
)

type UserRepo struct {
	db *sql.DB
}

func (r *UserRepo) FindUserUnsafe(userID string) {
	query := "SELECT * FROM users WHERE id = " + userID
	r.db.Query(query)
}

func GetTenantAccountsUnsafe(db *sql.DB) {
	db.Query("SELECT * FROM accounts WHERE status = 'active'")
}

func GetOrderUnscopedUnsafe(db *sql.DB, id string) {
	db.Query("SELECT * FROM orders WHERE id = $1", id)
}

func RawBypassAuthUnsafe(db *sql.DB) {
	db.Raw("SELECT * FROM users")
}

func RLSBypassUnsafe(db *sql.DB) {
	db.Exec("SET ROLE postgres")
}

func NonAtomicBalanceUnsafe(db *sql.DB, id string) {
	var bal int
	db.QueryRow("SELECT balance FROM accounts WHERE id = $1", id).Scan(&bal)
	bal += 50
	db.Exec("UPDATE accounts SET balance = $1 WHERE id = $2", bal, id)
}

func TxEscapeUnsafe(tx *sql.Tx, db *sql.DB) {
	db.Exec("DELETE FROM logs WHERE id = 1")
}

func LogicPrecedenceUnsafe(db *sql.DB) {
	db.Query("SELECT * FROM orders WHERE tenant_id = 1 AND status = 'active' OR is_admin = true")
}

func SoftDeleteBypassUnsafe(db *sql.DB) {
	db.Query("SELECT * FROM users WHERE status = 'active'")
}

func UnboundedQueryUnsafe(db *sql.DB) {
	db.Query("SELECT * FROM events WHERE created_at > 1000")
}

func NPlusOneLoopUnsafe(db *sql.DB, userIDs []string) {
	for _, id := range userIDs {
		db.QueryRow("SELECT * FROM profiles WHERE id = $1", id)
	}
}

func ErrorLeakedToResponseUnsafe(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	_, err := db.Query("SELECT 1")
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}
`

	s := New()
	req := scanner.Request{
		Sources: []scanner.Source{
			{
				Path: "handler.go",
				Open: func(ctx context.Context) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader(codeHandler)), nil
				},
			},
			{
				Path: "repo.go",
				Open: func(ctx context.Context) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader(codeRepo)), nil
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
	}

	expectedRules := []string{
		"SQLI-001",
		"SQLAUTH-001",
		"SQLAUTH-002",
		"SQLAUTH-003",
		"SQLAUTH-004",
		"SQLSAFE-003",
		"SQLSAFE-004",
		"SQLSAFE-005",
		"SQLSAFE-006",
		"DBPERF-001",
		"DBPERF-002",
		"DBSEC-003",
	}

	for _, ruleID := range expectedRules {
		if foundRules[ruleID] == 0 {
			t.Errorf("expected rule %s to trigger, findings: %v", ruleID, res.Findings)
		}
	}
}


