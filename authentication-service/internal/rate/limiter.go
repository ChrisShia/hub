package rate

import (
	"time"

	rl "github.com/ChrisShia/ratelimiter"
	"github.com/go-redis/redis"
)

func RedisLimiter(client *redis.Client, limit int, window time.Duration) *rl.Limiter {
	return rl.NewRedisLimiter(client, limit, window)
}
