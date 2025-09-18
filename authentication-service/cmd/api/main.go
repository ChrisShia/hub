package main

import (
	"authentication-service/internal/data"
	"context"
	"database/sql"
	"flag"
	"os"
	"time"

	"github.com/ChrisShia/jsonlog"
	rl "github.com/ChrisShia/ratelimiter"
	"github.com/go-redis/redis"
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

func (cfg *config) setFlags() {
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL DSN, postgres://user:password@host/db?sslmode=disable")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	flag.StringVar(&cfg.redis.addr, "redis", "", "Redis address, redis://<user>:<pass>@host:port/<db>")
	flag.Parse()
}

type application struct {
	config  config
	logger  *jsonlog.Logger
	models  data.Models
	limiter *rl.Limiter
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
