package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/ZinollinIlyas/tarkov_oauth/internal/core/model"
	"github.com/ZinollinIlyas/tarkov_oauth/internal/core/ports"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users     ports.UserRepository
	jwtSecret []byte
	jwtTTL    time.Duration
}

var _ ports.AuthService = (*AuthService)(nil)

func NewAuthService(users ports.UserRepository, jwtSecret string, jwtTTL time.Duration) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: []byte(jwtSecret),
		jwtTTL:    jwtTTL,
	}
}

func (s *AuthService) SignUp(ctx context.Context, in ports.SignUpInput) (*model.User, error) {
	nickname := strings.TrimSpace(in.Nickname)
	displayName := strings.TrimSpace(in.DisplayName)
	if displayName == "" {
		displayName = nickname
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		Nickname:    nickname,
		DisplayName: displayName,
		Password:    string(hash),
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) SignIn(ctx context.Context, in ports.SignInInput) (*ports.Token, error) {
	user, err := s.users.GetByNickname(ctx, strings.TrimSpace(in.Nickname))
	if err != nil {
		if err == ports.ErrUserNotFound {
			return nil, ports.ErrInvalidCredentials
		}
		return nil, err
	}

	if user.IsDisabled() {
		return nil, ports.ErrUserDisabled
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)) != nil {
		return nil, ports.ErrInvalidCredentials
	}

	now := time.Now()
	expiresAt := now.Add(s.jwtTTL)

	claims := jwt.RegisteredClaims{
		Subject:   user.ID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("sign token: %w", err)
	}

	if err := s.users.UpdateLastLogin(ctx, user.ID, now); err != nil {
		return nil, fmt.Errorf("update last login: %w", err)
	}

	return &ports.Token{AccessToken: token, ExpiresAt: expiresAt}, nil
}

func (s *AuthService) Authenticate(ctx context.Context, accessToken string) (*model.User, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(accessToken, claims,
		func(*jwt.Token) (any, error) { return s.jwtSecret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, ports.ErrInvalidToken
	}

	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, ports.ErrInvalidToken
	}

	user, err := s.users.GetByID(ctx, id)
	if errors.Is(err, ports.ErrUserNotFound) {
		return nil, ports.ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}
	if user.IsDisabled() {
		return nil, ports.ErrUserDisabled
	}
	return user, nil
}

func (s *AuthService) UsersByIDs(ctx context.Context, ids []uuid.UUID) ([]model.User, error) {
	return s.users.GetByIDs(ctx, ids)
}
