package auth

import (
	"GoTasker/internal/config"
	"GoTasker/internal/domain"
	"GoTasker/pkg/utils"
	"context"
	"fmt"
	"log/slog"
)

type UserPostgresRepo interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
}

type UserAuthUseCase struct {
	userRepo UserPostgresRepo
	cfg      *config.Config
}

func NewAuthUseCase(userRepo UserPostgresRepo, cfg *config.Config) *UserAuthUseCase {
	return &UserAuthUseCase{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

// Register регистрирует нового пользователя и генерирует JWT токены
func (uc *UserAuthUseCase) Register(ctx context.Context, user *domain.User) error {
	const op = "internal.useCase.auth.Register"

	if err := validateUser(user); err != nil {
		slog.Error(op,
			"ошибка валидации данных пользователся",
			slog.String("err", err.Error()),
		)
		return err
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		slog.Error(op, "ошибка при хэшировании пароля", slog.String("error", err.Error()))
		return fmt.Errorf("ошибка при хэшировании пароля: %w", err)
	}
	user.Password = hashedPassword

	err = uc.userRepo.Create(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

func (uc *UserAuthUseCase) Login(ctx context.Context, email, password string) (string, string, error) {
	const op = "internal.useCase.auth.login"

	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", "", err
	}

	err = utils.CheckPasswordHash(user.Password, password)
	if err != nil {
		slog.Error(op,
			"неверный пароль",
			slog.String("error", err.Error()),
			slog.String("email", email),
		)
		return "", "", fmt.Errorf("неверный пароль")
	}

	accessToken, err := utils.GenerateAccessToken(user.Username,
		user.Email,
		uc.cfg.Server.JWTSecret,
		uc.cfg.Server.AccessDuration,
	)
	if err != nil {
		slog.Error(op, "ошибка при генерации access токена", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("ошибка при генерации токенов: %w", err)
	}

	refreshToken, err := utils.GenerateRefreshToken(user.Email,
		uc.cfg.Server.JWTSecret,
		uc.cfg.Server.RefreshDuration,
	)
	if err != nil {
		slog.Error(op, "ошибка при генерации refresh токена", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("ошибка при генерации токенов: %w", err)
	}

	return accessToken, refreshToken, nil

}

func validateUser(user *domain.User) error {
	if user.Username == "" {
		return fmt.Errorf("имя не может быть пустым")
	}
	if user.Email == "" {
		return fmt.Errorf("email не может быть пустым")
	}
	if user.Password == "" {
		return fmt.Errorf("не указан пароль")
	}

	return nil
}
