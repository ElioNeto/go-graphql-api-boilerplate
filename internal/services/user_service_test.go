package services_test

import (
	"context"
	"testing"

	"github.com/ElioNeto/go-graphql-api-boilerplate/internal/models"
	"github.com/ElioNeto/go-graphql-api-boilerplate/internal/repositories"
	"github.com/ElioNeto/go-graphql-api-boilerplate/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByIDs(ctx context.Context, ids []string) ([]*models.User, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.User), args.Error(1)
}

func TestUserService_CreateUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := services.NewUserService(mockRepo, "secret")

	input := models.CreateUserInput{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "password",
	}

	mockRepo.On("GetByEmail", mock.Anything, input.Email).Return(nil, repositories.ErrNotFound)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(&models.User{ID: "1", Name: "Alice", Email: "alice@example.com"}, nil)

	user, err := svc.CreateUser(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "Alice", user.Name)
	mockRepo.AssertExpectations(t)
}
