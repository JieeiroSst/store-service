package infrastructure

import (
	"testing"

	"go.uber.org/fx"
)

// TestModuleGraph checks every dependency in the fx graph is provided,
// without running constructors (so no database is needed).
func TestModuleGraph(t *testing.T) {
	if err := fx.ValidateApp(Module); err != nil {
		t.Fatal(err)
	}
}
