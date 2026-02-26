package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ElioNeto/go-graphql-api-boilerplate/internal/models"
	"github.com/jmoiron/sqlx"
)

var ErrNotFound = errors.New("record not found")

type UserRepository interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByIDs(ctx context.Context, ids []string) ([]*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	List(ctx context.Context, limit, offset int) ([]*models.User, error)
}

type postgresUserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	query := `
		INSERT INTO users (name, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, password, created_at, updated_at`

	created := &models.User{}
	err := r.db.QueryRowxContext(ctx, query, user.Name, user.Email, user.Password).StructScan(created)
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}
	return created, nil
}

func (r *postgresUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `SELECT id, name, email, password, created_at, updated_at FROM users WHERE id = $1`
	user := &models.User{}
	err := r.db.QueryRowxContext(ctx, query, id).StructScan(user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("getting user by id: %w", err)
	}
	return user, nil
}

func (r *postgresUserRepository) GetByIDs(ctx context.Context, ids []string) ([]*models.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	
	query, args, err := sqlx.In(`SELECT id, name, email, password, created_at, updated_at FROM users WHERE id IN (?)`, ids)
	if err != nil {
		return nil, err
	}
	
	query = r.db.Rebind(query)
	var users []*models.User
	if err := r.db.SelectContext(ctx, &users, query, args...); err != nil {
		return nil, fmt.Errorf("getting users by ids: %w", err)
	}
	return users, nil
}

func (r *postgresUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, name, email, password, created_at, updated_at FROM users WHERE email = $1`
	user := &models.User{}
	err := r.db.QueryRowxContext(ctx, query, email).StructScan(user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("getting user by email: %w", err)
	}
	return user, nil
}

func (r *postgresUserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	query := `SELECT id, name, email, password, created_at, updated_at FROM users ORDER BY id LIMIT $1 OFFSET $2`
	var users []*models.User
	if err := r.db.SelectContext(ctx, &users, query, limit, offset); err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}
	return users, nil
}
