package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ElioNeto/go-graphql-api-boilerplate/internal/models"
	"github.com/ElioNeto/go-graphql-api-boilerplate/internal/repositories"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailTaken         = errors.New("email already taken")
	ErrUnauthorized       = errors.New("unauthorized")
)

type UserService interface {
	CreateUser(ctx context.Context, input models.CreateUserInput) (*models.User, error)
	Login(ctx context.Context, input models.LoginInput) (*models.AuthResponse, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	ListUsers(ctx context.Context, limit, offset *int) ([]*models.User, error)
}

type userService struct {
	repo      repositories.UserRepository
	jwtSecret string
}

func NewUserService(repo repositories.UserRepository, jwtSecret string) UserService {
	return &userService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *userService) CreateUser(ctx context.Context, input models.CreateUserInput) (*models.User, error) {
	_, err := s.repo.GetByEmail(ctx, input.Email)
	if err == nil {
		return nil, ErrEmailTaken
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	user := &models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashed),
	}

	return s.repo.Create(ctx, user)
}

func (s *userService) Login(ctx context.Context, input models.LoginInput) (*models.AuthResponse, error) {
	user, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ts, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("signing token: %w", err)
	}

	return &models.AuthResponse{
		Token: ts,
		User:  user,
	}, nil
}

func (s *userService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *userService) ListUsers(ctx context.Context, limit, offset *int) ([]*models.User, error) {
	l := 20
	if limit != nil {
		l = *limit
	}
	o := 0
	if offset != nil {
		o = *offset
	}
	return s.repo.List(ctx, l, o)
}
