package rate

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

func NewRateLimiter(client *redis.Client, limit int, window time.Duration) *L {
	return &L{
		client: client,
		limit:  limit,
		window: window,
	}
}

type L struct {
	client   *redis.Client
	limit    int
	window   time.Duration
	disabled bool
}

func (rl *L) Allow(ip string) (bool, error) {
	if rl.disabled {
		return true, nil
	}

	key := fmt.Sprintf("ratelimit:%s", ip)
	now := time.Now().Unix()

	res, err := rl.client.TxPipelined(func(pipe redis.Pipeliner) error {
		pipe.HIncrBy(key, "count", 1)
		pipe.HSetNX(key, "timestamp", now)
		pipe.Expire(key, rl.window)
		pipe.HGet(key, "timestamp")
		return nil
	})
	if err != nil {
		return false, err
	}

	fmt.Println("res:   ", res)

	count, err := res[0].(*redis.IntCmd).Result()
	if err != nil {
		return false, err
	}

	timestamp, err := res[3].(*redis.StringCmd).Int64()
	if err != nil {
		return false, err
	}

	if now-int64(rl.window.Seconds()) > timestamp {
		rl.client.HSet(key, "timestamp", now)
		rl.client.HSet(key, "count", 1)
		return true, nil
	}

	// check if the request count exceeds the limit
	if count > int64(rl.limit) {
		return false, nil
	}

	return true, nil
}
