package server

import (
	"context"
	"fmt"

	"github.com/bagdasarian/checklist-app/db_service/internal/repository/postgres"
	"github.com/bagdasarian/checklist-app/db_service/pkg/pb"
	"golang.org/x/crypto/bcrypt"
)

type TaskService struct {
	pb.UnimplementedDatabaseServiceServer
	userRepo postgres.UserRepositoryInterface
	taskRepo postgres.TaskRepositoryInterface
}

func NewTaskService(userRepo postgres.UserRepositoryInterface, taskRepo postgres.TaskRepositoryInterface) *TaskService {
	return &TaskService{
		userRepo: userRepo,
		taskRepo: taskRepo,
	}
}

// CreateUser создает нового пользователя
func (s *TaskService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	req.Password = string(hashedPassword)
	user, err := s.userRepo.CreateUser(ctx, req)
	if err != nil {
		return nil, err
	}

	return &pb.CreateUserResponse{
		Id:        user.Id,
		Name:      user.Name,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	}, nil
}

// GetUser возвращает информацию о пользователе
func (s *TaskService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err := s.userRepo.GetUserByID(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.GetUserResponse{
		Id:        user.Id,
		Name:      user.Name,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	}, nil
}

// AuthenticateUser проверяет учетные данные пользователя
func (s *TaskService) AuthenticateUser(ctx context.Context, req *pb.AuthenticateUserRequest) (*pb.AuthenticateUserResponse, error) {
	user, storedHash, err := s.userRepo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return &pb.AuthenticateUserResponse{
			Success: false,
			Message: "invalid credentials",
		}, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password))
	if err != nil {
		return &pb.AuthenticateUserResponse{
			Success: false,
			Message: "invalid credentials",
		}, nil
	}

	return &pb.AuthenticateUserResponse{
		Success: true,
		UserId:  user.Id,
		Message: "authenticated successfully",
	}, nil
}

// CreateTask создает новую задачу
func (s *TaskService) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.CreateTaskResponse, error) {
	task, err := s.taskRepo.CreateTask(ctx, req)
	if err != nil {
		return nil, err
	}

	return &pb.CreateTaskResponse{
		Id:          task.Id,
		UserId:      task.UserId,
		Title:       task.Title,
		Description: task.Description,
		Completed:   task.Completed,
		CreatedAt:   task.CreatedAt,
		CompletedAt: task.CompletedAt,
	}, nil
}

// GetTasks возвращает список задач пользователя
func (s *TaskService) GetTasks(ctx context.Context, req *pb.GetTasksRequest) (*pb.GetTasksResponse, error) {
	tasks, totalCount, err := s.taskRepo.GetTasks(ctx, req)
	if err != nil {
		return nil, err
	}

	return &pb.GetTasksResponse{
		Tasks:      tasks,
		TotalCount: totalCount,
	}, nil
}

// DeleteTask удаляет задачу
func (s *TaskService) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*pb.DeleteTaskResponse, error) {
	success, err := s.taskRepo.DeleteTask(ctx, req.Id, req.UserId)
	if err != nil {
		return nil, err
	}

	if !success {
		return &pb.DeleteTaskResponse{
			Success: false,
			Message: "task not found or access denied",
		}, nil
	}

	return &pb.DeleteTaskResponse{
		Success: true,
		Message: "task deleted successfully",
	}, nil
}

// CompleteTask отмечает задачу как выполненную
func (s *TaskService) CompleteTask(ctx context.Context, req *pb.CompleteTaskRequest) (*pb.CompleteTaskResponse, error) {
	task, err := s.taskRepo.CompleteTask(ctx, req.Id, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.CompleteTaskResponse{
		Id:          task.Id,
		Completed:   task.Completed,
		CompletedAt: task.CompletedAt,
	}, nil
}
