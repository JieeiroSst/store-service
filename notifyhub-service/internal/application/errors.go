package application

import (
	"fmt"

	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
)

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", port.ErrInvalidInput, fmt.Sprintf(format, args...))
}
