package main

import (
	"context"
	"database/sql"
	"os"
	"time"

	"github.com/ChrisShia/moviehub/internal/data"
	"github.com/ChrisShia/moviehub/internal/jsonlog"
	"github.com/ChrisShia/moviehub/internal/rate"
	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
)

const version = "1.0.0"

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
	limiter *rate.L
}

func main() {
	var cfg config

	cfg.setFlags()

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer db.Close()
	logger.PrintInfo("database connection pool established", nil)

	limiter, redisCloser, err := redisClientLimiter(cfg)
	if err != nil {
		logger.PrintFatal(err, map[string]string{
			"env":   cfg.env,
			"redis": cfg.redis.addr,
		})
	}
	defer redisCloser()

	app := &application{
		config:  cfg,
		logger:  logger,
		models:  data.NewModels(db),
		limiter: limiter,
	}

	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}

func redisClientLimiter(cfg config) (*rate.L, func(), error) {
	counts := 0
	for {
		client, err := establishRedisConnAndPing(cfg)
		if err != nil {
			counts++
		} else {
			return rate.NewRateLimiter(client, 4, time.Second), func() { client.Close() }, nil
		}

		if counts > 5 {
			return nil, nil, err
		}
	}
}

func establishRedisConnAndPing(cfg config) (*redis.Client, error) {
	client, err := establishRedisClient(cfg)
	if err != nil {
		return nil, err
	}

	if err = client.Ping().Err(); err != nil {
		return nil, err
	}
	return client, nil
}

func establishRedisClient(cfg config) (*redis.Client, error) {
	opt, err := redis.ParseURL(cfg.redis.addr)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)
	return client, nil
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
