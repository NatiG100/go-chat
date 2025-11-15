package usecases

import (
	"context"
	"time"

	"example.com/go-chat/internal/core"
	"example.com/go-chat/internal/drivers"
	"github.com/google/uuid"
)

type AuthUsecase struct {
	repos core.Repositories
	jwt   *drivers.JWTManager
}

func NewAuthUsecase(r core.Repositories, j *drivers.JWTManager) *AuthUsecase { return &AuthUsecase{repos: r, jwt: j} }

func (a *AuthUsecase) SignUp(ctx context.Context, username, email, password string) (*core.User, error) {
	u := &core.User{Username: username, Email: email}
	if err := a.repos.UserRepo().CreateUser(ctx, u, password); err != nil {
		return nil, err
	}
	return u, nil
}

func (a *AuthUsecase) Login(ctx context.Context, email, password string) (string, *core.User, error) {
	u, err := a.repos.UserRepo().VerifyPassword(ctx, email, password)
	if err != nil {
		return "", nil, err
	}
	token, err := a.jwt.Generate(u.ID, time.Hour*24)
	if err != nil {
		return "", nil, err
	}
	return token, u, nil
}

func (a *AuthUsecase) Me(ctx context.Context, id uuid.UUID) (*core.User, error) {
	return a.repos.UserRepo().GetUserByID(ctx, id)
}
