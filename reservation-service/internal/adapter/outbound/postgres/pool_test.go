package postgres

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

// Errors that mean "the work was not done, try again" must reach the caller as ErrBusy (429), never as a 500.
func TestDatabaseOverloadIsBusyNotAnInternalError(t *testing.T) {
	for code, name := range map[string]string{
		"53300": "too many clients: more replicas than the server has connections for",
		"55P03": "lock wait timeout on a hot row",
		"57014": "statement timeout",
		"40001": "serialization failure",
		"40P01": "deadlock",
		"57P03": "server is starting up",
	} {
		// pgx wraps server errors, sometimes several times, when a connection cannot be made.
		err := fmt.Errorf("failed to connect: %w", &pgconn.PgError{Code: code, Message: name})
		if got := mapErr(err); !errors.Is(got, domain.ErrBusy) {
			t.Errorf("%s (%s): got %v, want ErrBusy", code, name, got)
		}
	}
	if got := mapErr(&pgconn.PgError{Code: "23505"}); !errors.Is(got, domain.ErrConflict) {
		t.Errorf("unique violation must stay a conflict: %v", got)
	}
	if got := mapErr(errors.New("boom")); errors.Is(got, domain.ErrBusy) {
		t.Errorf("an unknown error must not be reported as busy")
	}
}
