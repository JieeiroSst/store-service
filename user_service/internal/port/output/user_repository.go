package output

import (
	"context"

	"github.com/JIeeiroSst/user-service/internal/domain"
)

type UserRepository interface {
	CheckAccount(ctx context.Context, user domain.User) (int, string, string, error)
	CheckAccountExists(ctx context.Context, user domain.User) error
	CreateAccount(ctx context.Context, user domain.User) (domain.User, error)
	FindUser(ctx context.Context, userID int) (domain.User, error)
	LockAccount(ctx context.Context, id int) error
	UpdateProfile(ctx context.Context, user domain.User) (domain.User, error)
}
