package http

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeiroSst/automatic-payment-service/pkg/authn"
	"gofr.dev/pkg/gofr"
)

func TestAdminDecorator(t *testing.T) {
	called := false
	h := admin(func(*gofr.Context) (interface{}, error) { called = true; return nil, nil })
	run := func(p *authn.Principal) error {
		ctx := context.Background()
		if p != nil {
			ctx = authn.WithPrincipal(ctx, *p)
		}
		_, err := h(&gofr.Context{Context: ctx})
		return err
	}
	if err := run(nil); !errors.Is(err, authn.ErrUnauthenticated) || called {
		t.Fatalf("anonymous: %v called=%v", err, called)
	}
	if err := run(&authn.Principal{UserID: "7", Role: "user"}); !errors.Is(err, authn.ErrForbidden) || called {
		t.Fatalf("user: %v called=%v", err, called)
	}
	if err := run(&authn.Principal{Service: true}); err != nil || !called {
		t.Fatalf("service: %v called=%v", err, called)
	}
}
