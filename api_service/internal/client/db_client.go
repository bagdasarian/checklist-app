package client

import (
	"context"
	"fmt"
	"log"

	dbpb "github.com/bagdasarian/checklist-app/db_service/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type DBClient struct {
	conn   *grpc.ClientConn
	client dbpb.DatabaseServiceClient
}

func NewDBClient(addr string) (*DBClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db_service: %w", err)
	}

	log.Printf("Successfully connected to db_service at %s", addr)

	return &DBClient{
		conn:   conn,
		client: dbpb.NewDatabaseServiceClient(conn),
	}, nil
}

func (c *DBClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *DBClient) CreateUser(ctx context.Context, req *dbpb.CreateUserRequest) (*dbpb.CreateUserResponse, error) {
	return c.client.CreateUser(ctx, req)
}

func (c *DBClient) AuthenticateUser(ctx context.Context, req *dbpb.AuthenticateUserRequest) (*dbpb.AuthenticateUserResponse, error) {
	return c.client.AuthenticateUser(ctx, req)
}

func (c *DBClient) GetUser(ctx context.Context, req *dbpb.GetUserRequest) (*dbpb.GetUserResponse, error) {
	return c.client.GetUser(ctx, req)
}

func (c *DBClient) CreateTask(ctx context.Context, req *dbpb.CreateTaskRequest) (*dbpb.CreateTaskResponse, error) {
	return c.client.CreateTask(ctx, req)
}

func (c *DBClient) GetTasks(ctx context.Context, req *dbpb.GetTasksRequest) (*dbpb.GetTasksResponse, error) {
	return c.client.GetTasks(ctx, req)
}

func (c *DBClient) DeleteTask(ctx context.Context, req *dbpb.DeleteTaskRequest) (*dbpb.DeleteTaskResponse, error) {
	return c.client.DeleteTask(ctx, req)
}

func (c *DBClient) CompleteTask(ctx context.Context, req *dbpb.CompleteTaskRequest) (*dbpb.CompleteTaskResponse, error) {
	return c.client.CompleteTask(ctx, req)
}

