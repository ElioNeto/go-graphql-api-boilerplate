package graph_test

import (
	"context"
	"testing"

	"github.com/ElioNeto/go-graphql-api-boilerplate/graph"
	"github.com/ElioNeto/go-graphql-api-boilerplate/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, input models.CreateUserInput) (*models.User, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) Login(ctx context.Context, input models.LoginInput) (*models.AuthResponse, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AuthResponse), args.Error(1)
}

func (m *MockUserService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) ListUsers(ctx context.Context, limit, offset *int) ([]*models.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.User), args.Error(1)
}

func TestQueryResolver_User(t *testing.T) {
	mockSvc := new(MockUserService)
	resolver := &graph.Resolver{UserService: mockSvc}
	q := resolver.Query()

	expectedUser := &models.User{ID: "1", Name: "Alice"}
	mockSvc.On("GetUserByID", mock.Anything, "1").Return(expectedUser, nil)

	user, err := q.User(context.Background(), "1")

	assert.NoError(t, err)
	assert.Equal(t, "Alice", user.Name)
	mockSvc.AssertExpectations(t)
}
