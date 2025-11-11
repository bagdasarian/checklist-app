package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	JWT struct {
		SecretKey     string `yaml:"secret_key" env:"JWT_SECRET_KEY" env-default:"your-secret-key-change-in-production"`
		TokenDuration int    `yaml:"token_duration" env:"JWT_TOKEN_DURATION" env-default:"3600"`
	} `yaml:"jwt"`

	DBService struct {
		Host string `yaml:"host" env:"DB_SERVICE_HOST" env-default:"localhost"`
		Port string `yaml:"port" env:"DB_SERVICE_PORT" env-default:"50051"`
	} `yaml:"db_service"`

	GRPC struct {
		Port string `yaml:"port" env:"GRPC_PORT" env-default:"50052"`
	} `yaml:"grpc"`

	HTTP struct {
		Port string `yaml:"port" env:"HTTP_PORT" env-default:"8080"`
	} `yaml:"http"`

	Kafka struct {
		Brokers []string `yaml:"brokers" env:"KAFKA_BROKERS" env-default:"localhost:9092"`
		Topic   string   `yaml:"topic" env:"KAFKA_TOPIC" env-default:"task-events"`
		Enabled bool     `yaml:"enabled" env:"KAFKA_ENABLED" env-default:"true"`
	} `yaml:"kafka"`
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

func (c *Config) GetDBServiceAddr() string {
	return fmt.Sprintf("%s:%s", c.DBService.Host, c.DBService.Port)
}

func (c *Config) GetTokenDuration() time.Duration {
	duration := time.Duration(c.JWT.TokenDuration) * time.Second
	if duration == 0 {
		return time.Hour
	}
	return duration
}

func (c *Config) GetKafkaBrokers() []string {
	if len(c.Kafka.Brokers) == 0 {
		return []string{"localhost:9092"}
	}
	return c.Kafka.Brokers
}

