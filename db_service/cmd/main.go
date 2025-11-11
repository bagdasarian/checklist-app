package main

import (
	"context"
	"log"
	"net"

	"github.com/bagdasarian/checklist-app/db_service/config"
	"github.com/bagdasarian/checklist-app/db_service/internal/repository/postgres"
	"github.com/bagdasarian/checklist-app/db_service/internal/server"
	"github.com/bagdasarian/checklist-app/db_service/pkg/pb"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()
	db, err := postgres.New(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	redisClient, err := postgres.NewRedis(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	userRepo := postgres.NewUserRepository(db)
	taskRepo := postgres.NewTaskRepository(db, redisClient)

	grpcServer := grpc.NewServer()
	taskService := server.NewTaskService(userRepo, taskRepo)
	pb.RegisterDatabaseServiceServer(grpcServer, taskService)

	lis, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Starting DB service on port %s", cfg.GRPC.Port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
