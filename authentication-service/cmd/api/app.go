package main

import (
	"authentication-service/internal/data"
	"context"
	"database/sql"
	"time"

	"github.com/ChrisShia/jsonlog"
	rl "github.com/ChrisShia/ratelimiter"
	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
)

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	redis struct {
		addr string
	}
	test bool
}

type application struct {
	config  config
	logger  *jsonlog.Logger
	models  data.Models
	limiter *rl.Limiter
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func redisClientLimiter(cfg config) (*rl.Limiter, func(), error) {
	counts := 0
	for {
		client, err := establishRedisConnAndPing(cfg)
		if err != nil {
			counts++
		} else {
			return rl.NewRedisLimiter(client, 4, time.Second), func() { client.Close() }, nil
		}

		if counts > 5 {
			return nil, nil, err
		}
	}
}

func establishRedisConnAndPing(cfg config) (*redis.Client, error) {
	client, err := redisClient(cfg)
	if err != nil {
		return nil, err
	}

	if err = client.Ping().Err(); err != nil {
		return nil, err
	}
	return client, nil
}

func redisClient(cfg config) (*redis.Client, error) {
	opt, err := redis.ParseURL(cfg.redis.addr)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)
	return client, nil
}
