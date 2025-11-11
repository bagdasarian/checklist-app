package consumer

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/bagdasarian/checklist-app/kafka_service/config"
	"github.com/bagdasarian/checklist-app/kafka_service/internal/logger"
	"github.com/bagdasarian/checklist-app/kafka_service/pkg/pb"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	logger *logger.Logger
	config *config.Config
}

func NewConsumer(cfg *config.Config, eventLogger *logger.Logger) (*Consumer, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.GetKafkaBrokers(),
		Topic:    cfg.Kafka.Topic,
		GroupID:  cfg.Kafka.GroupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	return &Consumer{
		reader: reader,
		logger: eventLogger,
		config: cfg,
	}, nil
}

// Start запускает обработку сообщений из Kafka
func (c *Consumer) Start(ctx context.Context) error {
	log.Printf("Starting Kafka consumer for topic: %s", c.config.Kafka.Topic)

	var retryCount int
	maxRetryDelay := 30 * time.Second
	baseDelay := 1 * time.Second

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Kafka consumer...")
			return c.reader.Close()
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					return nil
				}

				if isLeaderNotAvailableError(err) {
					retryCount++
					delay := baseDelay * time.Duration(1<<uint(min(retryCount, 5)))
					if delay > maxRetryDelay {
						delay = maxRetryDelay
					}
					log.Printf("Leader not available (attempt %d), retrying in %v: %v", retryCount, delay, err)
					time.Sleep(delay)
					continue
				}

				log.Printf("Error reading message: %v", err)
				retryCount = 0
				time.Sleep(baseDelay)
				continue
			}

			retryCount = 0

			var event pb.TaskEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}

			if err := c.logger.LogEvent(&event); err != nil {
				log.Printf("Error logging event: %v", err)
			}
		}
	}
}

// isLeaderNotAvailableError проверяет, является ли ошибка связанной с недоступностью лидера
func isLeaderNotAvailableError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "leader not available") ||
		strings.Contains(errStr, "leader election") ||
		strings.Contains(errStr, "no leader") ||
		strings.Contains(errStr, "not available for writes")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Close закрывает соединение с Kafka
func (c *Consumer) Close() error {
	return c.reader.Close()
}

