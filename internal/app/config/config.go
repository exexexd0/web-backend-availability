package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceHost string
	ServicePort int
	JWT         JWTConfig
	Redis       RedisConfig
}

type JWTConfig struct {
	Token     string
	ExpiresIn time.Duration
	Issuer    string
}

type RedisConfig struct {
	Host        string
	Password    string
	Port        int
	User        string
	DialTimeout time.Duration
	ReadTimeout time.Duration
}

func NewConfig() (*Config, error) {
	configName := "config"
	_ = godotenv.Load()
	if os.Getenv("CONFIG_NAME") != "" {
		configName = os.Getenv("CONFIG_NAME")
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")
	viper.WatchConfig()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}
	if cfg.JWT.Token == "" {
		return nil, fmt.Errorf("jwt token is empty")
	}
	if cfg.JWT.ExpiresIn <= 0 {
		return nil, fmt.Errorf("jwt expires_in must be positive")
	}
	if cfg.JWT.Issuer == "" {
		cfg.JWT.Issuer = "availability"
	}

	if err := applyRedisEnv(cfg); err != nil {
		return nil, err
	}

	log.Info("config parsed")
	return cfg, nil
}

func applyRedisEnv(cfg *Config) error {
	cfg.Redis.Host = envOr("REDIS_HOST", "127.0.0.1")
	cfg.Redis.User = os.Getenv("REDIS_USER")
	cfg.Redis.Password = envOr("REDIS_PASSWORD", "password")

	portText := envOr("REDIS_PORT", "6379")
	port, err := strconv.Atoi(portText)
	if err != nil {
		return fmt.Errorf("redis port must be int value: %w", err)
	}
	cfg.Redis.Port = port

	if cfg.Redis.DialTimeout <= 0 {
		cfg.Redis.DialTimeout = 10 * time.Second
	}
	if cfg.Redis.ReadTimeout <= 0 {
		cfg.Redis.ReadTimeout = 10 * time.Second
	}
	return nil
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
