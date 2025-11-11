package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/bagdasarian/checklist-app/db_service/pkg/pb"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TaskRepository struct {
	db    *Postgres
	redis *Redis
}

func NewTaskRepository(db *Postgres, redis *Redis) *TaskRepository {
	return &TaskRepository{
		db:    db,
		redis: redis,
	}
}

// CreateTask создает новую задачу
func (r *TaskRepository) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.DbTask, error) {
	query := `
        INSERT INTO tasks (user_id, title, description) 
        VALUES ($1, $2, $3) 
        RETURNING id, user_id, title, description, completed, created_at, completed_at
    `

	var task pb.DbTask
	var createdAt time.Time
	var completedAt *time.Time
	err := r.db.Pool.QueryRow(ctx, query,
		req.UserId, req.Title, req.Description,
	).Scan(&task.Id, &task.UserId, &task.Title, &task.Description,
		&task.Completed, &createdAt, &completedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	task.CreatedAt = timestamppb.New(createdAt)
	if completedAt != nil {
		task.CompletedAt = timestamppb.New(*completedAt)
	}

	if err := r.InvalidateCache(ctx, req.UserId); err == nil {
		fmt.Printf("[CACHE INVALIDATED] UserID: %s | Reason: Task created\n", req.UserId)
	}

	return &task, nil
}

// GetTasks возвращает список задач пользователя
func (r *TaskRepository) GetTasks(ctx context.Context, req *pb.GetTasksRequest) ([]*pb.DbTask, int32, error) {
	cacheKey := r.getCacheKey(req)

	if r.redis != nil && r.redis.Client != nil {
		cachedData, err := r.redis.Client.Get(ctx, cacheKey).Result()
		if err == nil {
			var result struct {
				Tasks      []*pb.DbTask `json:"tasks"`
				TotalCount int32        `json:"total_count"`
			}
			if err := json.Unmarshal([]byte(cachedData), &result); err == nil {
				fmt.Printf("[REDIS CACHE HIT] UserID: %s | Tasks: %d | Total: %d | Key: %s\n",
					req.UserId, len(result.Tasks), result.TotalCount, cacheKey)
				return result.Tasks, result.TotalCount, nil
			}
		} else if err != redis.Nil {
			fmt.Printf("[REDIS ERROR] %v\n", err)
		} else {
			fmt.Printf("[REDIS CACHE MISS] UserID: %s | Fetching from PostgreSQL | Key: %s\n",
				req.UserId, cacheKey)
		}
	}

	countQuery := `
        SELECT COUNT(*) FROM tasks 
        WHERE user_id = $1 AND ($2 OR completed = false)
    `
	var totalCount int32
	err := r.db.Pool.QueryRow(ctx, countQuery, req.UserId, req.IncludeCompleted).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	query := `
        SELECT id, user_id, title, description, completed, created_at, completed_at 
        FROM tasks 
        WHERE user_id = $1 AND ($2 OR completed = false)
        ORDER BY created_at DESC 
        LIMIT $3 OFFSET $4
    `

	rows, err := r.db.Pool.Query(ctx, query,
		req.UserId, req.IncludeCompleted, req.Limit, req.Offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*pb.DbTask
	for rows.Next() {
		var task pb.DbTask
		var createdAt time.Time
		var completedAt *time.Time
		err := rows.Scan(&task.Id, &task.UserId, &task.Title, &task.Description,
			&task.Completed, &createdAt, &completedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan task: %w", err)
		}
		task.CreatedAt = timestamppb.New(createdAt)
		if completedAt != nil {
			task.CompletedAt = timestamppb.New(*completedAt)
		}
		tasks = append(tasks, &task)
	}

	if r.redis != nil && r.redis.Client != nil {
		result := struct {
			Tasks      []*pb.DbTask `json:"tasks"`
			TotalCount int32        `json:"total_count"`
		}{
			Tasks:      tasks,
			TotalCount: totalCount,
		}
		data, err := json.Marshal(result)
		if err == nil {
			if err := r.redis.Client.Set(ctx, cacheKey, data, r.redis.TTL).Err(); err != nil {
				fmt.Printf("[REDIS WRITE ERROR] %v\n", err)
			} else {
				fmt.Printf("[CACHED TO REDIS] UserID: %s | Tasks: %d | Total: %d | TTL: %v | Key: %s\n",
					req.UserId, len(tasks), totalCount, r.redis.TTL, cacheKey)
			}
		}
	}

	fmt.Printf("[POSTGRESQL] userId: %s | Tasks retrieved: %d | Total: %d\n",
		req.UserId, len(tasks), totalCount)

	return tasks, totalCount, nil
}

func (r *TaskRepository) getCacheKey(req *pb.GetTasksRequest) string {
	return fmt.Sprintf("tasks:user:%s:completed:%v:limit:%d:offset:%d",
		req.UserId, req.IncludeCompleted, req.Limit, req.Offset)
}

// DeleteTask удаляет задачу
func (r *TaskRepository) DeleteTask(ctx context.Context, taskID, userID string) (bool, error) {
	query := `
        DELETE FROM tasks 
        WHERE id = $1 AND user_id = $2
    `

	result, err := r.db.Pool.Exec(ctx, query, taskID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to delete task: %w", err)
	}

	deleted := result.RowsAffected() > 0
	if deleted {
		if err := r.InvalidateCache(ctx, userID); err == nil {
			fmt.Printf("[CACHE INVALIDATED] UserID: %s | Reason: Task deleted\n", userID)
		}
	}

	return deleted, nil
}

// CompleteTask отмечает задачу как выполненную
func (r *TaskRepository) CompleteTask(ctx context.Context, taskID, userID string) (*pb.DbTask, error) {
	query := `
        UPDATE tasks 
        SET completed = true, completed_at = NOW()
        WHERE id = $1 AND user_id = $2
        RETURNING id, user_id, title, description, completed, created_at, completed_at
    `

	var task pb.DbTask
	var createdAt time.Time
	var completedAt *time.Time
	err := r.db.Pool.QueryRow(ctx, query, taskID, userID).Scan(
		&task.Id, &task.UserId, &task.Title, &task.Description,
		&task.Completed, &createdAt, &completedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("task not found or access denied")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to complete task: %w", err)
	}

	task.CreatedAt = timestamppb.New(createdAt)
	if completedAt != nil {
		task.CompletedAt = timestamppb.New(*completedAt)
	}

	if err := r.InvalidateCache(ctx, userID); err == nil {
		fmt.Printf("userId: %s | reason: task completed\n", userID)
	}

	return &task, nil
}

// InvalidateCache удаляет кэш для пользователя
func (r *TaskRepository) InvalidateCache(ctx context.Context, userID string) error {
	if r.redis == nil || r.redis.Client == nil {
		return nil
	}

	pattern := fmt.Sprintf("tasks:user:%s:*", userID)
	keys, err := r.redis.Client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get cache keys: %w", err)
	}

	if len(keys) > 0 {
		if err := r.redis.Client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete cache keys: %w", err)
		}
		fmt.Printf("userId: %s, keys deleted: %d\n", userID, len(keys))
	}

	return nil
}
