package app

import (
	"context"
	"fmt"
	"time"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/config"
	goredis "github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg config.Redis) *goredis.Client {
	client := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		panic(fmt.Sprintf("failed to connect to Redis at %s: %v", cfg.Address, err))
	}
	return client
}
