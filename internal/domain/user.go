package domain

// User предоставляет пользователя
type User struct {
	ID       int64
	Username string
	Email    string
	Password string
}

// NewUser конструктор для User
func NewUser(id int64, username, email, password string) *User {
	return &User{
		ID:       id,
		Username: username,
		Email:    email,
		Password: password,
	}
}
