package infrastructure

import (
	"testing"

	"go.uber.org/fx"
)

func TestDependencyGraph(t *testing.T) {
	if err := fx.ValidateApp(Options()...); err != nil {
		t.Fatal(err)
	}
}
