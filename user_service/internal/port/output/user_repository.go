package output

import (
	"context"

	"github.com/JIeeiroSst/user-service/internal/domain"
)

type UserRepository interface {
	CheckAccount(ctx context.Context, user domain.User) (int, string, string, error)
	CheckAccountExists(ctx context.Context, user domain.User) error
	CreateAccount(ctx context.Context, user domain.User) (domain.User, error)
	DeleteAccount(ctx context.Context, id int) error
	FindUser(ctx context.Context, userID int) (domain.User, error)
	SearchUsers(ctx context.Context, filter domain.UserFilter) ([]domain.User, int64, error)
	LockAccount(ctx context.Context, id int) error
	UpdateProfile(ctx context.Context, user domain.User) (domain.User, error)
}
