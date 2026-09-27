package dbcheck

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestValidateReadOnly(t *testing.T) {
	ok := []string{
		"SELECT 1",
		"select o.id from orders o where o.total <> (select sum(price) from items) ;",
		"WITH x AS (SELECT 1) SELECT * FROM x",
		"SELECT 'update; delete' AS s",
	}
	for _, q := range ok[:3] {
		if err := ValidateReadOnly(q); err != nil {
			t.Errorf("%q should pass: %v", q, err)
		}
	}
	bad := map[string]string{
		"":                                      "empty",
		"DELETE FROM orders":                    "SELECT/WITH",
		"SELECT 1; DROP TABLE orders":           "single statement",
		"SELECT * FROM a INTO OUTFILE '/tmp/x'": "into",
		"SELECT pg_sleep(100)":                  "pg_sleep",
		"SELECT nextval('s')":                   "nextval",
		"WITH d AS (DELETE FROM t RETURNING *) SELECT * FROM d": "delete",
		"/* hi */ UPDATE t SET a=1":                             "SELECT/WITH",
		"SELECT 1 -- ok\n; DROP TABLE x":                        "single statement",
	}
	for q, want := range bad {
		err := ValidateReadOnly(q)
		if err == nil || !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(want)) {
			t.Errorf("%q: err=%v, want contains %q", q, err, want)
		}
	}
}

func TestEqualScalar(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want bool
	}{{"10.50", "10.5", true}, {"100", "100.00", true}, {"59.97", "59.96", false}, {"abc", "abc", true}, {"abc", "abd", false}} {
		if equalScalar(c.a, c.b) != c.want {
			t.Errorf("equalScalar(%q,%q) != %v", c.a, c.b, c.want)
		}
	}
}

func runIntegration(t *testing.T, dialect, dsn string) {
	db, err := Open(dialect, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	stmts := []string{
		"DROP TABLE IF EXISTS qa_items", "DROP TABLE IF EXISTS qa_orders", "DROP TABLE IF EXISTS qa_users", "DROP TABLE IF EXISTS qa_log",
		"CREATE TABLE qa_users (id INT PRIMARY KEY, name VARCHAR(50))",
		"CREATE TABLE qa_orders (id INT PRIMARY KEY, qa_user_id INT, total DECIMAL(10,2))",
		"CREATE TABLE qa_items (id INT PRIMARY KEY, order_id INT, price DOUBLE PRECISION)",
		"CREATE TABLE qa_log (msg VARCHAR(20))",
		"INSERT INTO qa_users VALUES (1,'a')",
		"INSERT INTO qa_orders VALUES (1,1,59.97),(2,99,-5.00)",
		"INSERT INTO qa_items VALUES (1,1,19.99),(2,1,19.99)",
	}
	if dialect == "mysql" {
		stmts[6] = "CREATE TABLE qa_items (id INT PRIMARY KEY, order_id INT, price DOUBLE)"
	}
	for _, s := range stmts {
		if _, err := db.ExecContext(ctx, s); err != nil {
			t.Fatalf("setup %q: %v", s, err)
		}
	}
	defer func() {
		for _, s := range []string{"DROP TABLE qa_items", "DROP TABLE qa_orders", "DROP TABLE qa_users", "DROP TABLE qa_log"} {
			db.ExecContext(context.Background(), s)
		}
	}()

	rep, err := Audit(ctx, db, dialect, []Assertion{
		{Name: "order total equals sum of items", SQL: "SELECT o.id FROM qa_orders o WHERE o.total <> (SELECT SUM(i.price) FROM qa_items i WHERE i.order_id = o.id)"},
		{Name: "user count", SQL: "SELECT COUNT(*) FROM qa_users", Expect: "value", Value: "1"},
		{Name: "no negative totals", SQL: "SELECT id FROM qa_orders WHERE total < 0"},
		{Name: "must be rejected", SQL: "DELETE FROM qa_users"},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, f := range rep.Findings {
		got[f.ID+"|"+f.Title] = true
	}
	has := func(id, sub string) bool {
		for _, f := range rep.Findings {
			if f.ID == id && strings.Contains(f.Title, sub) {
				return true
			}
		}
		return false
	}
	for id, sub := range map[string]string{
		"DB-PK-001": "qa_log", "DB-MONEY-001": "qa_items.price", "DB-MONEY-002": "qa_orders.total", "DB-ORPHAN-001": "qa_orders.qa_user_id",
	} {
		if !has(id, sub) {
			t.Errorf("missing finding %s %q; got %v", id, sub, rep.Findings)
		}
	}
	want := map[string]bool{"order total equals sum of items": false, "user count": true, "no negative totals": false, "must be rejected": false}
	for _, a := range rep.Assertions {
		if a.Passed != want[a.Name] {
			t.Errorf("assertion %q passed=%v want %v (%s)", a.Name, a.Passed, want[a.Name], a.Detail)
		}
	}

	var n int
	db.QueryRow("SELECT COUNT(*) FROM qa_users").Scan(&n)
	if n != 1 {
		t.Fatal("assertion modified data")
	}
}

func TestPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN not set")
	}
	runIntegration(t, "postgres", dsn)
}

func TestMySQL(t *testing.T) {
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN not set")
	}
	runIntegration(t, "mysql", dsn)
}
