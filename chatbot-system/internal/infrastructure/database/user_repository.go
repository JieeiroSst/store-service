package database

import (
	"context"
	"database/sql"
	"fmt"

	"chatbot-system/internal/domain"
)

type MySQLUserRepository struct {
	db *sql.DB
}

func NewMySQLUserRepository(db *sql.DB) *MySQLUserRepository {
	return &MySQLUserRepository{db: db}
}

func (r *MySQLUserRepository) Create(ctx context.Context, user *domain.User) error {
	// user.ID is the user-service id of the account joining the chat.
	query := `
		INSERT INTO users (id, role, manager_id, advisor_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())
	`
	if _, err := r.db.ExecContext(ctx, query, user.ID, user.Role, user.ManagerID, user.AdvisorID); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *MySQLUserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `
		SELECT id, role, manager_id, advisor_id, created_at, updated_at
		FROM users
		WHERE id = ?
	`
	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Role,
		&user.ManagerID,
		&user.AdvisorID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *MySQLUserRepository) GetManagedUsers(ctx context.Context, managerID int64) ([]*domain.User, error) {
	query := `
		SELECT id, role, manager_id, advisor_id, created_at, updated_at
		FROM users
		WHERE manager_id = ?
	`
	rows, err := r.db.QueryContext(ctx, query, managerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get managed users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		if err := rows.Scan(
			&user.ID,
			&user.Role,
			&user.ManagerID,
			&user.AdvisorID,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *MySQLUserRepository) GetAdvisedUser(ctx context.Context, advisorID int64) (*domain.User, error) {
	query := `
		SELECT id, role, manager_id, advisor_id, created_at, updated_at
		FROM users
		WHERE advisor_id = ?
		LIMIT 1
	`
	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, advisorID).Scan(
		&user.ID,
		&user.Role,
		&user.ManagerID,
		&user.AdvisorID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no advised user found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get advised user: %w", err)
	}

	return user, nil
}

func (r *MySQLUserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET role = ?, manager_id = ?, advisor_id = ?, updated_at = NOW()
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, user.Role, user.ManagerID, user.AdvisorID, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}
