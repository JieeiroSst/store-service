package infrastructure

import (
	"testing"

	"go.uber.org/fx"
)

func TestModuleGraphIsComplete(t *testing.T) {
	t.Setenv("CONSUL_ADDR", "")
	t.Setenv("AUTHORIZE_KEY", "test-key")
	if err := fx.ValidateApp(Module); err != nil {
		t.Fatalf("fx graph: %v", err)
	}
}
