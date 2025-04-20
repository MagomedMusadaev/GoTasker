package users

import (
	"GoTasker/internal/domain"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"log/slog"
)

type UserPostgresRepo struct {
	db *sql.DB
}

func NewUserPostgresRepo(db *sql.DB) *UserPostgresRepo {
	return &UserPostgresRepo{
		db: db,
	}
}

// Create создает нового пользователя в базе данных
func (r *UserPostgresRepo) Create(ctx context.Context, user *domain.User) error {
	const op = "internal.repository.postgres.user_repo.Create"

	query := `INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3)`

	_, err := r.db.ExecContext(ctx, query, user.Username, user.Email, user.Password)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return fmt.Errorf("%w", errors.New("email уже используется"))
		}

		slog.Error(op,
			slog.String("email", user.Email),
			slog.String("error", err.Error()),
		)

		return errors.New("ошибка при вставке пользователя")
	}

	return nil
}

// FindByEmail находит пользователя по email в базе данных
func (r *UserPostgresRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	const op = "internal.repository.postgres.user_repo.FindByEmail"

	query := `SELECT id, username, email, password_hash FROM users WHERE email = $1`

	row := r.db.QueryRowContext(ctx, query, email)

	var user domain.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("пользователь не найден")
		}

		slog.Error(op,
			slog.String("email", email),
			slog.String("error", err.Error()),
		)

		return nil, fmt.Errorf("ошибка при поиске пользователя")
	}

	return &user, nil
}
