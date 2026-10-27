package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/one-compressive/web-backend-availability/internal/app/config"
	goredis "github.com/redis/go-redis/v9"
)

const servicePrefix = "availability."

type Client struct {
	client *goredis.Client
}

func New(ctx context.Context, cfg config.RedisConfig) (*Client, error) {
	client := goredis.NewClient(&goredis.Options{
		Addr:        cfg.Host + ":" + strconv.Itoa(cfg.Port),
		Username:    cfg.User,
		Password:    cfg.Password,
		DB:          0,
		DialTimeout: cfg.DialTimeout,
		ReadTimeout: cfg.ReadTimeout,
	})
	if _, err := client.Ping(ctx).Result(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("cant ping redis: %w", err)
	}
	return &Client{client: client}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

func jwtKey(token string) string {
	return servicePrefix + "jwt." + token
}

// WriteJWTToBlacklist сохраняет отозванный токен до истечения его срока.
// В значении хранится логин, чтобы в Redis было видно, какой пользователь вышел.
func (c *Client) WriteJWTToBlacklist(ctx context.Context, token, login string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = time.Second
	}
	return c.client.Set(ctx, jwtKey(token), login, ttl).Err()
}

// CheckJWTInBlacklist возвращает redis.Nil, если токена в чёрном списке нет.
func (c *Client) CheckJWTInBlacklist(ctx context.Context, token string) error {
	return c.client.Get(ctx, jwtKey(token)).Err()
}
