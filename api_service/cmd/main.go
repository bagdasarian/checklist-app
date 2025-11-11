package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/bagdasarian/checklist-app/api_service/config"
	"github.com/bagdasarian/checklist-app/api_service/internal/client"
	"github.com/bagdasarian/checklist-app/api_service/internal/middleware"
	"github.com/bagdasarian/checklist-app/api_service/internal/producer"
	"github.com/bagdasarian/checklist-app/api_service/internal/server"
	"github.com/bagdasarian/checklist-app/api_service/internal/service"
	"github.com/bagdasarian/checklist-app/api_service/pkg/pb"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// customHTTPErrorHandler преобразует gRPC ошибку в http method not allowed 405, а не internal server error 501
func customHTTPErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	const fallback = `{"code": 13, "message": "Internal server error", "details": []}`

	w.Header().Set("Content-Type", marshaler.ContentType(nil))

	s := status.Convert(err)
	httpStatus := runtime.HTTPStatusFromCode(s.Code())

	if s.Code() == codes.Unimplemented {
		if strings.Contains(strings.ToLower(err.Error()), "method") ||
			strings.Contains(strings.ToLower(s.Message()), "method") {
			httpStatus = http.StatusMethodNotAllowed
		}
	}

	w.WriteHeader(httpStatus)

	body := map[string]interface{}{
		"code":    int(s.Code()),
		"message": s.Message(),
		"details": s.Details(),
	}

	buf, merr := json.Marshal(body)
	if merr != nil {
		w.Write([]byte(fallback))
		return
	}

	w.Write(buf)
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()

	dbClient, err := client.NewDBClient(cfg.GetDBServiceAddr())
	if err != nil {
		log.Fatalf("Failed to connect to db_service: %v", err)
	}
	defer dbClient.Close()

	jwtManager := service.NewJWTManager(cfg.JWT.SecretKey, cfg.GetTokenDuration())

	var kafkaProducer *producer.Producer
	if cfg.Kafka.Enabled {
		kafkaProducer = producer.NewProducer(cfg.GetKafkaBrokers(), cfg.Kafka.Topic)
		defer kafkaProducer.Close()
		log.Printf("Kafka producer initialized for topic: %s", cfg.Kafka.Topic)
	} else {
		log.Println("Kafka producer disabled")
	}

	authInterceptor := middleware.NewAuthInterceptor(jwtManager)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
	)

	taskService := server.NewTaskService(dbClient, jwtManager, kafkaProducer)
	pb.RegisterTaskServiceServer(grpcServer, taskService)

	grpcListener, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port: %v", err)
	}

	go func() {
		log.Printf("Starting gRPC server on port %s", cfg.GRPC.Port)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	mux := runtime.NewServeMux(
		runtime.WithErrorHandler(customHTTPErrorHandler),
	)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err = pb.RegisterTaskServiceHandlerFromEndpoint(ctx, mux, ":"+cfg.GRPC.Port, opts)
	if err != nil {
		log.Fatalf("Failed to register gateway: %v", err)
	}

	log.Printf("Starting HTTP Gateway server on port %s", cfg.HTTP.Port)
	if err := http.ListenAndServe(":"+cfg.HTTP.Port, mux); err != nil {
		log.Fatalf("Failed to serve HTTP: %v", err)
	}
}
