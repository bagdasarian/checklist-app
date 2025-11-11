package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	kafkapb "github.com/bagdasarian/checklist-app/api_service/pkg/pb/kafka"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Producer struct {
	writer *kafka.Writer
	topic  string
}

func NewProducer(brokers []string, topic string) *Producer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
		Async:    true,
	}

	return &Producer{
		writer: writer,
		topic:  topic,
	}
}

// SendEvent отправляет событие в Kafka
func (p *Producer) SendEvent(ctx context.Context, action kafkapb.ActionType, userID, taskID, details string) error {
	event := &kafkapb.TaskEvent{
		Timestamp: timestamppb.Now(),
		Action:    action,
		UserId:    userID,
		TaskId:    taskID,
		Details:   details,
	}

	jsonData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(userID),
		Value: jsonData,
		Time:  time.Now(),
	}

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		err := p.writer.WriteMessages(ctx, msg)
		if err == nil {
			return nil
		}

		log.Printf("Failed to send event to Kafka (attempt %d/%d): %v", i+1, maxRetries, err)

		if i < maxRetries-1 {
			time.Sleep(time.Second * time.Duration(i+1))
		}
	}

	log.Printf("Failed to send event to Kafka after %d attempts, continuing without event", maxRetries)
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
