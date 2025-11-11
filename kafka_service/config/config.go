package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Kafka struct {
		Brokers []string `yaml:"brokers" env:"KAFKA_BROKERS" env-default:"localhost:9092"`
		Topic   string   `yaml:"topic" env:"KAFKA_TOPIC" env-default:"task-events"`
		GroupID string   `yaml:"group_id" env:"KAFKA_GROUP_ID" env-default:"kafka-service"`
	} `yaml:"kafka"`

	Logging struct {
		FilePath string `yaml:"file_path" env:"LOG_FILE_PATH" env-default:"/var/log/kafka-service/events.log"`
		MaxSize  int    `yaml:"max_size" env:"LOG_MAX_SIZE" env-default:"100"` // MB
		MaxFiles int    `yaml:"max_files" env:"LOG_MAX_FILES" env-default:"10"`
	} `yaml:"logging"`
}

func Load() (*Config, error) {
	var cfg Config

	err := cleanenv.ReadConfig("config.yml", &cfg)
	if err != nil {
		err = cleanenv.ReadEnv(&cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	return &cfg, nil
}

func (c *Config) GetKafkaBrokers() []string {
	if len(c.Kafka.Brokers) == 0 {
		return []string{"localhost:9092"}
	}
	return c.Kafka.Brokers
}

