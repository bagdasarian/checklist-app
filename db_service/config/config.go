package config

import (
    "fmt"
    "github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
    DB struct {
        Host     string `yaml:"host" env:"DB_HOST" env-default:"postgres"`
        Port     string `yaml:"port" env:"DB_PORT" env-default:"5432"`
        User     string `yaml:"user" env:"DB_USER" env-default:"docker"`
        Password string `yaml:"password" env:"DB_PASSWORD" env-default:"docker"`
        Name     string `yaml:"name" env:"DB_NAME" env-default:"test_db"`
    } `yaml:"db"`
    
    Redis struct {
        Host     string `yaml:"host" env:"REDIS_HOST" env-default:"redis"`
        Port     string `yaml:"port" env:"REDIS_PORT" env-default:"6379"`
        Password string `yaml:"password" env:"REDIS_PASSWORD" env-default:""`
        DB       int    `yaml:"db" env:"REDIS_DB" env-default:"0"`
        TTL      int    `yaml:"ttl" env:"REDIS_TTL" env-default:"300"`
    } `yaml:"redis"`
    
    GRPC struct {
        Port string `yaml:"port" env:"GRPC_PORT" env-default:"50051"`
    } `yaml:"grpc"`
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

func (c *Config) GetDBURL() string {
    return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
        c.DB.User, c.DB.Password, c.DB.Host, c.DB.Port, c.DB.Name)
}

func (c *Config) GetRedisAddr() string {
    return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}