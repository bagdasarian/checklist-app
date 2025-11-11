package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bagdasarian/checklist-app/kafka_service/config"
	"github.com/bagdasarian/checklist-app/kafka_service/internal/consumer"
	"github.com/bagdasarian/checklist-app/kafka_service/internal/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	eventLogger, err := logger.NewLogger(cfg)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer eventLogger.Close()

	kafkaConsumer, err := consumer.NewConsumer(cfg, eventLogger)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer kafkaConsumer.Close()

	log.Println("Waiting for Kafka to initialize...")
	time.Sleep(10 * time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := kafkaConsumer.Start(ctx); err != nil {
			log.Printf("Consumer error: %v", err)
			cancel()
		}
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := eventLogger.RotateLog(); err != nil {
					log.Printf("Error rotating log: %v", err)
				}
			}
		}
	}()

	log.Println("Kafka service started. Waiting for messages...")

	<-sigChan
	log.Println("Shutting down Kafka service...")
	cancel()
}

