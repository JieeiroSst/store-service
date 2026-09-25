package infrastructure

import (
	"testing"

	"go.uber.org/fx"
)

// The dependency graph must resolve: every port has exactly one adapter.
func TestModuleGraph(t *testing.T) {
	if err := fx.ValidateApp(Module); err != nil {
		t.Fatal(err)
	}
}
