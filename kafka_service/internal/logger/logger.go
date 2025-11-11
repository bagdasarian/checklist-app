package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/bagdasarian/checklist-app/kafka_service/config"
	"github.com/bagdasarian/checklist-app/kafka_service/pkg/pb"
)

type Logger struct {
	file   *os.File
	mu     sync.Mutex
	config *config.Config
}

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	UserID    string    `json:"user_id"`
	TaskID    string    `json:"task_id,omitempty"`
	Details   string    `json:"details,omitempty"`
}

func NewLogger(cfg *config.Config) (*Logger, error) {
	// Создаем директорию для логов, если её нет
	logDir := filepath.Dir(cfg.Logging.FilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Открываем файл для записи (append mode)
	file, err := os.OpenFile(cfg.Logging.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &Logger{
		file:   file,
		config: cfg,
	}, nil
}

func (l *Logger) LogEvent(event *pb.TaskEvent) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Конвертируем proto timestamp в time.Time
	var timestamp time.Time
	if event.Timestamp != nil {
		timestamp = event.Timestamp.AsTime()
	} else {
		timestamp = time.Now()
	}

	// Конвертируем ActionType в строку
	actionStr := event.Action.String()

	// Создаем запись лога
	logEntry := LogEntry{
		Timestamp: timestamp,
		Action:    actionStr,
		UserID:    event.UserId,
		TaskID:    event.TaskId,
		Details:   event.Details,
	}

	// Сериализуем в JSON
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// Записываем в файл с новой строкой
	if _, err := l.file.WriteString(string(jsonData) + "\n"); err != nil {
		return fmt.Errorf("failed to write to log file: %w", err)
	}

	// Синхронизируем файл
	if err := l.file.Sync(); err != nil {
		return fmt.Errorf("failed to sync log file: %w", err)
	}

	return nil
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Close()
}

// RotateLog проверяет размер файла и выполняет ротацию при необходимости
func (l *Logger) RotateLog() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	info, err := l.file.Stat()
	if err != nil {
		return err
	}

	maxSize := int64(l.config.Logging.MaxSize * 1024 * 1024) // Конвертируем MB в байты
	if info.Size() < maxSize {
		return nil // Файл еще не достиг максимального размера
	}

	// Закрываем текущий файл
	if err := l.file.Close(); err != nil {
		return err
	}

	// Переименовываем текущий файл с timestamp
	timestamp := time.Now().Format("20060102-150405")
	oldPath := l.config.Logging.FilePath
	newPath := fmt.Sprintf("%s.%s", oldPath, timestamp)
	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}

	// Открываем новый файл
	file, err := os.OpenFile(oldPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	l.file = file

	// Удаляем старые файлы, если их больше чем max_files
	l.cleanupOldLogs()

	return nil
}

func (l *Logger) cleanupOldLogs() {
	logDir := filepath.Dir(l.config.Logging.FilePath)
	logBase := filepath.Base(l.config.Logging.FilePath)

	files, err := filepath.Glob(filepath.Join(logDir, logBase+".*"))
	if err != nil {
		return
	}

	// Сортируем файлы по времени модификации (новые первыми)
	// И удаляем лишние
	if len(files) > l.config.Logging.MaxFiles {
		// Простая реализация: удаляем самые старые файлы
		// В production лучше использовать более сложную логику
		for i := l.config.Logging.MaxFiles; i < len(files); i++ {
			os.Remove(files[i])
		}
	}
}

