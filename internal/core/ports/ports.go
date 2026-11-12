package ports

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/ZinollinIlyas/tarkov_oauth/internal/core/model"
)

var (
	ErrNicknameTaken      = errors.New("nickname already taken")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid nickname or password")
	ErrUserDisabled       = errors.New("user is disabled")
	ErrInvalidToken       = errors.New("invalid token")
)

type SignUpInput struct {
	Nickname    string
	DisplayName string
	Password    string
}

type SignInInput struct {
	Nickname string
	Password string
}

type Token struct {
	AccessToken string
	ExpiresAt   time.Time
}

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByNickname(ctx context.Context, nickname string) (*model.User, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID, at time.Time) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]model.User, error)
}

type AuthService interface {
	SignUp(ctx context.Context, in SignUpInput) (*model.User, error)
	SignIn(ctx context.Context, in SignInInput) (*Token, error)
	// Authenticate проверяет access-токен и возвращает его владельца.
	// ErrInvalidToken — токен невалиден/просрочен или пользователя нет,
	// ErrUserDisabled — пользователь заблокирован.
	Authenticate(ctx context.Context, accessToken string) (*model.User, error)
	UsersByIDs(ctx context.Context, ids []uuid.UUID) ([]model.User, error)
}
