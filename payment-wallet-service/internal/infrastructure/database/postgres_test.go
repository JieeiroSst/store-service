package database

import (
	"reflect"
	"testing"
)

func TestSplitStatementsIgnoresSemicolonsInComments(t *testing.T) {
	script := `-- header; with a semicolon
CREATE TABLE a (
    x INT,
    -- 0 = unlimited. Gate Withdraw only;
    -- top-up limits are not our concern.
    y INT
);

CREATE INDEX i ON a(x);
-- trailing comment; here
`
	got := splitStatements(script)
	want := []string{"CREATE TABLE a (\n    x INT,\n    y INT\n)", "CREATE INDEX i ON a(x)"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q\nwant %q", got, want)
	}
}
