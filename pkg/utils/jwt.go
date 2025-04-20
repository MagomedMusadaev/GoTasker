package utils

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

// AccessTokenClaim предоставляет payload для access токена
type AccessTokenClaim struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// RefreshTokenClaim предоставляет payload для refresh токена
type RefreshTokenClaim struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateAccessToken генерирует access токен
func GenerateAccessToken(username, email, secretKey string, ttl time.Duration) (string, error) {
	claims := AccessTokenClaim{
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secretKey))
}

// GenerateRefreshToken генерирует refresh токен
func GenerateRefreshToken(email, secretKey string, ttl time.Duration) (string, error) {
	claims := RefreshTokenClaim{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secretKey))
}

// ValidateToken валидирует токен и возвращает распарсенный jwt.Token
func ValidateToken(encodedToken string, secretKey []byte) (*jwt.Token, error) {
	token, err := jwt.Parse(encodedToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неизвестный метод подписания")
		}
		return secretKey, nil
	})
	if err != nil {
		return nil, errors.New("невалидный токен")
	}

	return token, nil
}
