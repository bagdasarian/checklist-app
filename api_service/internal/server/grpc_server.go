package server

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bagdasarian/checklist-app/api_service/internal/client"
	"github.com/bagdasarian/checklist-app/api_service/internal/middleware"
	"github.com/bagdasarian/checklist-app/api_service/internal/producer"
	"github.com/bagdasarian/checklist-app/api_service/internal/service"
	"github.com/bagdasarian/checklist-app/api_service/pkg/pb"
	kafkapb "github.com/bagdasarian/checklist-app/api_service/pkg/pb/kafka"
	dbpb "github.com/bagdasarian/checklist-app/db_service/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TaskService struct {
	pb.UnimplementedTaskServiceServer
	dbClient      client.DBClientInterface
	jwtManager    *service.JWTManager
	kafkaProducer *producer.Producer
}

func NewTaskService(dbClient client.DBClientInterface, jwtManager *service.JWTManager, kafkaProducer *producer.Producer) *TaskService {
	return &TaskService{
		dbClient:      dbClient,
		jwtManager:    jwtManager,
		kafkaProducer: kafkaProducer,
	}
}

// RegisterUser регистрирует нового пользователя
func (s *TaskService) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}
	if strings.TrimSpace(req.Username) == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username is required")
	}
	if strings.TrimSpace(req.Password) == "" {
		return nil, status.Errorf(codes.InvalidArgument, "password is required")
	}
	if len(req.Password) < 6 {
		return nil, status.Errorf(codes.InvalidArgument, "password must be at least 6 characters long")
	}

	createUserReq := &dbpb.CreateUserRequest{
		Name:     strings.TrimSpace(req.Name),
		Username: strings.TrimSpace(req.Username),
		Password: req.Password,
	}

	createUserResp, err := s.dbClient.CreateUser(ctx, createUserReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &pb.RegisterUserResponse{
		Id:        createUserResp.Id,
		Name:      createUserResp.Name,
		Username:  createUserResp.Username,
		CreatedAt: createUserResp.CreatedAt,
	}, nil
}

// LoginUser аутентифицирует пользователя и возвращает JWT токен
func (s *TaskService) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	if strings.TrimSpace(req.Username) == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username is required")
	}
	if strings.TrimSpace(req.Password) == "" {
		return nil, status.Errorf(codes.InvalidArgument, "password is required")
	}

	authReq := &dbpb.AuthenticateUserRequest{
		Username: strings.TrimSpace(req.Username),
		Password: req.Password,
	}

	authResp, err := s.dbClient.AuthenticateUser(ctx, authReq)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	if !authResp.Success {
		return nil, status.Errorf(codes.Unauthenticated, "%s", authResp.Message)
	}

	getUserReq := &dbpb.GetUserRequest{
		UserId: authResp.UserId,
	}
	userResp, err := s.dbClient.GetUser(ctx, getUserReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	token, err := s.jwtManager.Generate(authResp.UserId, userResp.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &pb.LoginUserResponse{
		Success: true,
		UserId:  authResp.UserId,
		Message: "login successful",
		Token:   token,
	}, nil
}

// CreateTask создает новую задачу
func (s *TaskService) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.CreateTaskResponse, error) {
	if strings.TrimSpace(req.Title) == "" {
		return nil, status.Errorf(codes.InvalidArgument, "title is required")
	}

	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	createTaskReq := &dbpb.CreateTaskRequest{
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		UserId:      userID,
	}

	createTaskResp, err := s.dbClient.CreateTask(ctx, createTaskReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	if s.kafkaProducer != nil {
		go func() {
			kafkaCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.kafkaProducer.SendEvent(kafkaCtx, kafkapb.ActionType_ACTION_CREATE_TASK, userID, createTaskResp.Id, fmt.Sprintf("Created task: %s", createTaskResp.Title)); err != nil {
				fmt.Printf("Failed to send Kafka event: %v\n", err)
			}
		}()
	}

	return &pb.CreateTaskResponse{
		Id:          createTaskResp.Id,
		UserId:      createTaskResp.UserId,
		Title:       createTaskResp.Title,
		Description: createTaskResp.Description,
		Completed:   createTaskResp.Completed,
		CreatedAt:   createTaskResp.CreatedAt,
		CompletedAt: createTaskResp.CompletedAt,
	}, nil
}

// GetTasks получает список задач пользователя
func (s *TaskService) GetTasks(ctx context.Context, req *pb.GetTasksRequest) (*pb.GetTasksResponse, error) {
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		return nil, status.Errorf(codes.InvalidArgument, "limit cannot exceed 100")
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	getTasksReq := &dbpb.GetTasksRequest{
		UserId:           userID,
		IncludeCompleted: req.IncludeCompleted,
		Limit:            limit,
		Offset:           offset,
	}

	getTasksResp, err := s.dbClient.GetTasks(ctx, getTasksReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	if s.kafkaProducer != nil {
		go func() {
			kafkaCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.kafkaProducer.SendEvent(kafkaCtx, kafkapb.ActionType_ACTION_GET_TASKS, userID, "", fmt.Sprintf("Retrieved %d tasks", getTasksResp.TotalCount)); err != nil {
				fmt.Printf("Failed to send Kafka event: %v\n", err)
			}
		}()
	}

	tasks := make([]*pb.Task, len(getTasksResp.Tasks))
	for i, task := range getTasksResp.Tasks {
		tasks[i] = &pb.Task{
			Id:          task.Id,
			UserId:      task.UserId,
			Title:       task.Title,
			Description: task.Description,
			Completed:   task.Completed,
			CreatedAt:   task.CreatedAt,
			CompletedAt: task.CompletedAt,
		}
	}

	return &pb.GetTasksResponse{
		Tasks:      tasks,
		TotalCount: getTasksResp.TotalCount,
	}, nil
}

// DeleteTask удаляет задачу
func (s *TaskService) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*pb.DeleteTaskResponse, error) {
	if strings.TrimSpace(req.Id) == "" {
		return nil, status.Errorf(codes.InvalidArgument, "task id is required")
	}

	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	deleteTaskReq := &dbpb.DeleteTaskRequest{
		Id:     req.Id,
		UserId: userID,
	}

	deleteTaskResp, err := s.dbClient.DeleteTask(ctx, deleteTaskReq)
	if err != nil {
		return nil, fmt.Errorf("failed to delete task: %w", err)
	}

	if !deleteTaskResp.Success {
		return nil, status.Errorf(codes.NotFound, "task not found")
	}

	if s.kafkaProducer != nil {
		go func() {
			kafkaCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.kafkaProducer.SendEvent(kafkaCtx, kafkapb.ActionType_ACTION_DELETE_TASK, userID, req.Id, "Task deleted"); err != nil {
				fmt.Printf("Failed to send Kafka event: %v\n", err)
			}
		}()
	}

	return &pb.DeleteTaskResponse{
		Success: true,
		Message: "task deleted successfully",
	}, nil
}

// CompleteTask отмечает задачу как выполненную
func (s *TaskService) CompleteTask(ctx context.Context, req *pb.CompleteTaskRequest) (*pb.CompleteTaskResponse, error) {
	if strings.TrimSpace(req.Id) == "" {
		return nil, status.Errorf(codes.InvalidArgument, "task id is required")
	}

	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	completeTaskReq := &dbpb.CompleteTaskRequest{
		Id:     req.Id,
		UserId: userID,
	}

	completeTaskResp, err := s.dbClient.CompleteTask(ctx, completeTaskReq)
	if err != nil {
		if strings.Contains(err.Error(), "task not found") {
			return nil, status.Errorf(codes.NotFound, "task not found")
		}
		return nil, fmt.Errorf("failed to complete task: %w", err)
	}

	if s.kafkaProducer != nil {
		go func() {
			kafkaCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.kafkaProducer.SendEvent(kafkaCtx, kafkapb.ActionType_ACTION_COMPLETE_TASK, userID, req.Id, "Task completed"); err != nil {
				fmt.Printf("Failed to send Kafka event: %v\n", err)
			}
		}()
	}

	return &pb.CompleteTaskResponse{
		Id:          completeTaskResp.Id,
		Completed:   completeTaskResp.Completed,
		CompletedAt: completeTaskResp.CompletedAt,
	}, nil
}
