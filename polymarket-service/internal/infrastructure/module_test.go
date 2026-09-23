package infrastructure

import (
	"testing"

	"go.uber.org/fx"
)

// TestModuleGraph fails if any constructor's dependencies cannot be satisfied,
// without connecting to MySQL or any other service.
func TestModuleGraph(t *testing.T) {
	if err := fx.ValidateApp(Module); err != nil {
		t.Fatal(err)
	}
}
