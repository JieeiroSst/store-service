package infrastructure

import (
	"testing"

	"github.com/JIeeiroSst/catalogues-service/config"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

// TestGraph checks that every constructor's dependencies are satisfied,
// without needing a database.
func TestGraph(t *testing.T) {
	err := fx.ValidateApp(
		fx.Supply(&config.Config{}, (*gorm.DB)(nil)),
		Core,
	)
	if err != nil {
		t.Fatal(err)
	}
}
