package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bagdasarian/checklist-app/db_service/pkg/pb"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserRepository struct {
	db *Postgres
}

func NewUserRepository(db *Postgres) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	query := `
        INSERT INTO users (name, username, password_hash) 
        VALUES ($1, $2, $3) 
        RETURNING id, name, username, created_at
    `

	var user pb.User
	var createdAt time.Time
	err := r.db.Pool.QueryRow(ctx, query,
		req.Name, req.Username, req.Password,
	).Scan(&user.Id, &user.Name, &user.Username, &createdAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	user.CreatedAt = timestamppb.New(createdAt)
	return &user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, userID string) (*pb.User, error) {
	query := `
        SELECT id, name, username, created_at 
        FROM users 
        WHERE id = $1
    `

	var user pb.User
	var createdAt time.Time
	err := r.db.Pool.QueryRow(ctx, query, userID).Scan(
		&user.Id, &user.Name, &user.Username, &createdAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.CreatedAt = timestamppb.New(createdAt)
	return &user, nil
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*pb.User, string, error) {
	query := `
        SELECT id, name, username, password_hash, created_at 
        FROM users 
        WHERE username = $1
    `

	var user struct {
		ID           string
		Name         string
		Username     string
		PasswordHash string
		CreatedAt    interface{}
	}

	err := r.db.Pool.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Name, &user.Username, &user.PasswordHash, &user.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, "", fmt.Errorf("failed to get user by username: %w", err)
	}

	pbUser := &pb.User{
		Id:        user.ID,
		Name:      user.Name,
		Username:  user.Username,
		CreatedAt: convertToTimestamp(user.CreatedAt),
	}

	return pbUser, user.PasswordHash, nil
}
