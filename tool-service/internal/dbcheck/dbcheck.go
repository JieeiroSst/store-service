package dbcheck

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

const (
	High   = "high"
	Medium = "medium"
	Low    = "low"
	Info   = "info"
)

type Finding struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Title    string `json:"title"`
	Detail   string `json:"detail,omitempty"`
}

type Assertion struct {
	Name   string `json:"name"`
	SQL    string `json:"sql"`
	Expect string `json:"expect,omitempty"`
	Value  string `json:"value,omitempty"`
}

type AssertionResult struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Detail string `json:"detail,omitempty"`
}

type Report struct {
	Dialect    string            `json:"dialect"`
	LatencyMs  int64             `json:"latency_ms"`
	Findings   []Finding         `json:"findings"`
	Assertions []AssertionResult `json:"assertions,omitempty"`
}

var forbidden = regexp.MustCompile(`(?i)\b(insert|update|delete|drop|alter|create|truncate|grant|revoke|merge|call|exec|execute|copy|lock|set|into|pg_sleep|sleep|benchmark|nextval|setval)\b`)

func ValidateReadOnly(q string) error {
	s := stripComments(q)
	s = strings.TrimSpace(strings.TrimRight(strings.TrimSpace(s), ";"))
	if s == "" {
		return errors.New("empty query")
	}
	if strings.Contains(s, ";") {
		return errors.New("only a single statement is allowed")
	}
	low := strings.ToLower(s)
	if !strings.HasPrefix(low, "select") && !strings.HasPrefix(low, "with") {
		return errors.New("only SELECT/WITH queries are allowed")
	}
	if m := forbidden.FindString(stripStrings(s)); m != "" {
		return fmt.Errorf("keyword %q is not allowed in assertions", strings.ToLower(m))
	}
	return nil
}

var (
	lineComment  = regexp.MustCompile(`--[^\n]*`)
	blockComment = regexp.MustCompile(`(?s)/\*.*?\*/`)
	sqlString    = regexp.MustCompile(`'(?:[^']|'')*'`)
)

func stripComments(s string) string {
	return blockComment.ReplaceAllString(lineComment.ReplaceAllString(s, " "), " ")
}
func stripStrings(s string) string { return sqlString.ReplaceAllString(s, "''") }

func Open(dialect, dsn string) (*sql.DB, error) {
	driver := map[string]string{"postgres": "postgres", "mysql": "mysql"}[dialect]
	if driver == "" {
		return nil, fmt.Errorf("dialect must be postgres or mysql")
	}
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(2)
	db.SetConnMaxLifetime(time.Minute)
	return db, nil
}

type session struct {
	tx      *sql.Tx
	dialect string
}

func begin(ctx context.Context, db *sql.DB, dialect string) (*session, error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}

	switch dialect {
	case "postgres":
		_, err = tx.ExecContext(ctx, "SET LOCAL statement_timeout = 15000")
	case "mysql":
		_, err = tx.ExecContext(ctx, "SET SESSION max_execution_time = 15000")
	}
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	return &session{tx, dialect}, nil
}

func (s *session) quote(id string) string {
	if s.dialect == "mysql" {
		return "`" + strings.ReplaceAll(id, "`", "``") + "`"
	}
	return `"` + strings.ReplaceAll(id, `"`, `""`) + `"`
}

func (s *session) schemaFilter() string {
	if s.dialect == "mysql" {
		return "table_schema = DATABASE()"
	}
	return "table_schema NOT IN ('pg_catalog','information_schema')"
}

func (s *session) count(ctx context.Context, q string) (int64, error) {
	var n int64
	err := s.tx.QueryRowContext(ctx, q).Scan(&n)
	return n, err
}

func Audit(ctx context.Context, db *sql.DB, dialect string, asserts []Assertion) (*Report, error) {
	rep := &Report{Dialect: dialect}
	t0 := time.Now()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("cannot connect: %w", err)
	}
	rep.LatencyMs = time.Since(t0).Milliseconds()

	s, err := begin(ctx, db, dialect)
	if err != nil {
		return nil, err
	}
	defer s.tx.Rollback()

	add := func(id, sev, title, detail string) {
		rep.Findings = append(rep.Findings, Finding{id, sev, title, detail})
	}

	s.tablesWithoutPK(ctx, add)
	s.moneyColumns(ctx, add)
	s.implicitOrphans(ctx, add)

	for _, a := range asserts {
		rep.Assertions = append(rep.Assertions, s.assertion(ctx, a))
	}
	return rep, nil
}

func (s *session) tablesWithoutPK(ctx context.Context, add func(id, sev, title, detail string)) {
	q := `SELECT t.table_name FROM information_schema.tables t
	      WHERE t.table_type='BASE TABLE' AND ` + strings.ReplaceAll(s.schemaFilter(), "table_schema", "t.table_schema") + `
	      AND NOT EXISTS (SELECT 1 FROM information_schema.table_constraints c
	          WHERE c.table_schema=t.table_schema AND c.table_name=t.table_name AND c.constraint_type='PRIMARY KEY')
	      ORDER BY 1 LIMIT 50`
	rows, err := s.tx.QueryContext(ctx, q)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var t string
		if rows.Scan(&t) == nil {
			add("DB-PK-001", Medium, "table has no primary key: "+t, "Rows cannot be uniquely identified; duplicates and lost updates become possible.")
		}
	}
}

var moneyName = regexp.MustCompile(`(?i)(amount|price|balance|total|fee|cost|charge|tax|discount|refund|payment|credit|debit|quantity|qty)`)
var numericType = regexp.MustCompile(`(?i)^(smallint|integer|int|bigint|tinyint|mediumint|numeric|decimal|real|double precision|double|float)$`)
var floatType = regexp.MustCompile(`(?i)^(real|double precision|double|float)$`)

type column struct{ table, name, typ string }

func (s *session) columns(ctx context.Context, where string) []column {
	q := `SELECT table_name, column_name, data_type FROM information_schema.columns WHERE ` + s.schemaFilter() + where + ` ORDER BY 1,2 LIMIT 2000`
	rows, err := s.tx.QueryContext(ctx, q)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []column
	for rows.Next() {
		var c column
		if rows.Scan(&c.table, &c.name, &c.typ) == nil {
			out = append(out, c)
		}
	}
	return out
}

func (s *session) moneyColumns(ctx context.Context, add func(id, sev, title, detail string)) {
	checked := 0
	for _, c := range s.columns(ctx, "") {
		if !moneyName.MatchString(c.name) || !numericType.MatchString(c.typ) || checked >= 100 {
			continue
		}
		checked++
		if floatType.MatchString(c.typ) {
			add("DB-MONEY-001", Medium, fmt.Sprintf("money-like column stored as %s: %s.%s", c.typ, c.table, c.name), "Binary floating point cannot represent decimal amounts exactly; use NUMERIC/DECIMAL or integer cents.")
		}
		n, err := s.count(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s < 0", s.quote(c.table), s.quote(c.name)))
		if err != nil || n == 0 {
			continue
		}
		sev := Medium
		if regexp.MustCompile(`(?i)balance|credit|debit|discount|refund`).MatchString(c.name) {
			sev = Low
		}
		add("DB-MONEY-002", sev, fmt.Sprintf("%d negative value(s) in %s.%s", n, c.table, c.name), "Verify negative values are legitimate for this column.")
	}
}

func (s *session) implicitOrphans(ctx context.Context, add func(id, sev, title, detail string)) {
	cols := s.columns(ctx, ` AND column_name LIKE '%\_id'`)
	tables := map[string]bool{}
	for _, c := range s.columns(ctx, ` AND column_name = 'id'`) {
		tables[c.table] = true
	}
	checked := 0
	for _, c := range cols {
		if checked >= 60 {
			break
		}
		base := strings.TrimSuffix(c.name, "_id")
		var ref string
		for _, cand := range []string{base, base + "s", base + "es", strings.TrimSuffix(base, "y") + "ies"} {
			if tables[cand] && cand != c.table {
				ref = cand
				break
			}
		}
		if ref == "" {
			continue
		}
		checked++
		q := fmt.Sprintf("SELECT COUNT(*) FROM %[1]s t WHERE t.%[2]s IS NOT NULL AND NOT EXISTS (SELECT 1 FROM %[3]s r WHERE r.id = t.%[2]s)",
			s.quote(c.table), s.quote(c.name), s.quote(ref))
		n, err := s.count(ctx, q)
		if err != nil || n == 0 {
			continue
		}
		add("DB-ORPHAN-001", High, fmt.Sprintf("%d orphan row(s): %s.%s has no matching %s.id", n, c.table, c.name, ref), "Rows reference a parent that does not exist. Add a foreign key or clean the data.")
	}
}

func (s *session) assertion(ctx context.Context, a Assertion) AssertionResult {
	res := AssertionResult{Name: a.Name}
	if err := ValidateReadOnly(a.SQL); err != nil {
		res.Detail = "rejected: " + err.Error()
		return res
	}
	q := strings.TrimRight(strings.TrimSpace(a.SQL), ";")
	switch a.Expect {
	case "", "no_rows":
		rows, err := s.tx.QueryContext(ctx, "SELECT * FROM ("+q+") AS v LIMIT 5")
		if err != nil {
			res.Detail = "query failed: " + err.Error()
			return res
		}
		defer rows.Close()
		cols, _ := rows.Columns()
		n := 0
		var sample []string
		for rows.Next() {
			vals := make([]sql.NullString, len(cols))
			ptrs := make([]any, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if rows.Scan(ptrs...) == nil {
				var parts []string
				for i, v := range vals {
					parts = append(parts, cols[i]+"="+v.String)
				}
				sample = append(sample, strings.Join(parts, ", "))
			}
			n++
		}
		res.Passed = n == 0
		if !res.Passed {
			res.Detail = fmt.Sprintf("%d violation(s) found, e.g. %s", n, strings.Join(sample, " | "))
		}
	case "value":
		var got sql.NullString
		if err := s.tx.QueryRowContext(ctx, q).Scan(&got); err != nil {
			res.Detail = "query failed: " + err.Error()
			return res
		}
		res.Passed = equalScalar(got.String, a.Value)
		if !res.Passed {
			res.Detail = fmt.Sprintf("want %s, got %s", a.Value, got.String)
		}
	default:
		res.Detail = `expect must be "no_rows" or "value"`
	}
	return res
}

func equalScalar(got, want string) bool {
	g, gok := new(rat).parse(got)
	w, wok := new(rat).parse(want)
	if gok && wok {
		return g.cmp(w)
	}
	return got == want
}
