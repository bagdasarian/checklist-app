package postgres

import (
	"context"

	"github.com/bagdasarian/checklist-app/db_service/pkg/pb"
)

type UserRepositoryInterface interface {
	CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error)
	GetUserByID(ctx context.Context, userID string) (*pb.User, error)
	GetUserByUsername(ctx context.Context, username string) (*pb.User, string, error)
}

type TaskRepositoryInterface interface {
	CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.DbTask, error)
	GetTasks(ctx context.Context, req *pb.GetTasksRequest) ([]*pb.DbTask, int32, error)
	DeleteTask(ctx context.Context, taskID, userID string) (bool, error)
	CompleteTask(ctx context.Context, taskID, userID string) (*pb.DbTask, error)
	InvalidateCache(ctx context.Context, userID string) error
}

