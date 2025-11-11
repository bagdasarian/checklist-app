package middleware

import (
	"context"
	"strings"

	"github.com/bagdasarian/checklist-app/api_service/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	authorizationHeader = "authorization"
	bearerPrefix        = "Bearer "
)

type AuthInterceptor struct {
	jwtManager *service.JWTManager
}

func NewAuthInterceptor(jwtManager *service.JWTManager) *AuthInterceptor {
	return &AuthInterceptor{
		jwtManager: jwtManager,
	}
}

// Unary возвращает интерсептор для аутентификации
func (interceptor *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}

		authHeaders := md.Get(authorizationHeader)
		if len(authHeaders) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}

		accessToken := authHeaders[0]
		if !strings.HasPrefix(accessToken, bearerPrefix) {
			return nil, status.Errorf(codes.Unauthenticated, "invalid authorization header format")
		}

		accessToken = strings.TrimPrefix(accessToken, bearerPrefix)

		claims, err := interceptor.jwtManager.Verify(accessToken)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
		}

		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)

		return handler(ctx, req)
	}
}

// isPublicMethod проверяет, является ли метод публичным
func isPublicMethod(method string) bool {
	publicMethods := []string{
		"/checklist.api.TaskService/RegisterUser",
		"/checklist.api.TaskService/LoginUser",
		"/checklist.TaskService/RegisterUser",
		"/checklist.TaskService/LoginUser",
	}
	for _, publicMethod := range publicMethods {
		if method == publicMethod {
			return true
		}
	}
	return false
}

// GetUserIDFromContext извлекает user_id из контекста
func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return "", status.Errorf(codes.Internal, "user_id not found in context")
	}
	return userID, nil
}
