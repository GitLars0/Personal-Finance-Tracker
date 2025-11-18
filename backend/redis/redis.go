// backend/redis/redis.go
package redis

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis() error {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		log.Println("⚠️ REDIS_URL not set, skipping Redis init")
		return nil
	}

	opt, err := redis.ParseURL(url)
	if err != nil {
		return err
	}

	RDB = redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := RDB.Ping(ctx).Err(); err != nil {
		return err
	}

	log.Println("✅ Connected to Redis!")
	return nil
}
