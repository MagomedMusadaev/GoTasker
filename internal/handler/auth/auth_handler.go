package auth

import (
	"GoTasker/internal/domain"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	ErrMsgEmptyUsername   = "имя не может быть пустым"
	ErrMsgEmptyEmail      = "email не может быть пустым"
	ErrMsgEmptyPassword   = "не указан пароль"
	ErrMsgEmailExists     = "email уже используется"
	ErrMsgUserNotFound    = "пользователь не найден"
	ErrMsgInvalidPassword = "неверный пароль"
	ErrMsgServerError     = "Не удалось выполнить операцию. Попробуйте позже."
)

type UserAuthUseCase interface {
	Register(ctx context.Context, user *domain.User) error
	Login(ctx context.Context, email, password string) (string, string, error)
}

type UserAuthHandler struct {
	authUseCase UserAuthUseCase
}

func NewUserAuthHandler(authUseCase UserAuthUseCase) *UserAuthHandler {
	return &UserAuthHandler{
		authUseCase: authUseCase,
	}
}

// @Summary Регистрация нового пользователя
// @Description Регистрирует нового пользователя в системе
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param user body domain.User true "Данные пользователя"
// @Success 201 {object} map[string]string "Пользователь успешно зарегистрирован"
// @Failure 400 {object} map[string]string "Ошибка валидации"
// @Failure 409 {object} map[string]string "Email уже используется"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /auth/register [post]
func (h *UserAuthHandler) Register(c *gin.Context) {
	const op = "internal.handler.auth.Register"

	var user domain.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Невалидные JSON"})
		return
	}
	ctx := c.Request.Context()

	err := h.authUseCase.Register(ctx, &user)
	if err != nil {
		switch err.Error() {
		case ErrMsgEmptyUsername:
			c.AbortWithStatusJSON(http.StatusBadRequest,
				gin.H{"error": "Имя не может быть пустым"})
		case ErrMsgEmptyEmail:
			c.AbortWithStatusJSON(http.StatusBadRequest,
				gin.H{"error": "Email не может быть пустым"})
		case ErrMsgEmptyPassword:
			c.AbortWithStatusJSON(http.StatusBadRequest,
				gin.H{"error": "Пароль не может быть пустым"})
		case ErrMsgEmailExists:
			c.AbortWithStatusJSON(http.StatusConflict,
				gin.H{"error": "Пользователь с таким email уже существует"})
		case ErrMsgUserNotFound:
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": "Пользователь не найден"})
		case ErrMsgInvalidPassword:
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": "Неверный пароль"})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError,
				gin.H{"error": ErrMsgServerError})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Пользователь успешно зарегистрирован"})
}

// @Summary Авторизация пользователя
// @Description Авторизует пользователя и возвращает токены доступа
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param credentials body domain.User true "Учетные данные пользователя"
// @Success 200 {object} map[string]string "Токены доступа"
// @Failure 400 {object} map[string]string "Ошибка валидации"
// @Failure 401 {object} map[string]string "Неверные учетные данные"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /auth/login [post]
func (h *UserAuthHandler) Login(c *gin.Context) {
	const op = "internal.handler.auth.login"

	var user domain.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Невалидные данные"},
		)

		return
	}
	ctx := c.Request.Context()

	accessToken, refreshToken, err := h.authUseCase.Login(ctx, user.Email, user.Password)
	if err != nil {
		switch err.Error() {
		case ErrMsgUserNotFound, ErrMsgInvalidPassword:
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": "Неверный email или пароль"})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError,
				gin.H{"error": "Не удалось выполнить операцию. Попробуйте позже."},
			)
		}
		return
	}

	c.JSON(http.StatusOK,
		gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
	)
}
