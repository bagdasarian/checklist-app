package client

import (
	"context"

	dbpb "github.com/bagdasarian/checklist-app/db_service/pkg/pb"
)

type DBClientInterface interface {
	CreateUser(ctx context.Context, req *dbpb.CreateUserRequest) (*dbpb.CreateUserResponse, error)
	AuthenticateUser(ctx context.Context, req *dbpb.AuthenticateUserRequest) (*dbpb.AuthenticateUserResponse, error)
	GetUser(ctx context.Context, req *dbpb.GetUserRequest) (*dbpb.GetUserResponse, error)
	CreateTask(ctx context.Context, req *dbpb.CreateTaskRequest) (*dbpb.CreateTaskResponse, error)
	GetTasks(ctx context.Context, req *dbpb.GetTasksRequest) (*dbpb.GetTasksResponse, error)
	DeleteTask(ctx context.Context, req *dbpb.DeleteTaskRequest) (*dbpb.DeleteTaskResponse, error)
	CompleteTask(ctx context.Context, req *dbpb.CompleteTaskRequest) (*dbpb.CompleteTaskResponse, error)
	Close() error
}

