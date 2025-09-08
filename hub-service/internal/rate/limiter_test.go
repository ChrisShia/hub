package rate

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis"
)

func TestRateLimit(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		//Addr: os.Getenv("REDIS_URL"),
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer client.Close()

	limiter := NewRateLimiter(client, 3, time.Second)

	allowed, _ := limiter.Allow("127.0.0.1")
	allowed, _ = limiter.Allow("127.0.0.1")

	if allowed {
		fmt.Println("OK")
	} else {
		fmt.Println("Not allowed")
	}
}
