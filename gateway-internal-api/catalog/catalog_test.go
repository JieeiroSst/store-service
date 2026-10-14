package catalog

import (
	"encoding/json"
	"io/fs"
	"path"
	"testing"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

func TestExamplesValidateAgainstGatewaySchema(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	names, _ := fs.Glob(files, "schema/*.graphqls")
	var sources []*ast.Source
	for _, n := range names {
		b, _ := files.ReadFile(n)
		sources = append(sources, &ast.Source{Name: path.Base(n), Input: string(b)})
	}
	schema, gerr := gqlparser.LoadSchema(sources...)
	if gerr != nil {
		t.Fatal(gerr)
	}
	if len(c.Operations) == 0 {
		t.Fatal("no operations loaded")
	}
	for _, op := range c.Operations {
		if _, errs := gqlparser.LoadQuery(schema, op.Example.Query); len(errs) > 0 {
			t.Errorf("%s: %v", op.ID, errs)
		}
		var v map[string]any
		if err := json.Unmarshal([]byte(op.Example.Variables), &v); err != nil {
			t.Errorf("%s variables: %v", op.ID, err)
		}
	}
}
