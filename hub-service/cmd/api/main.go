package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
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
		addr     string
		disabled bool
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
		logger.PrintFatal(err, nil)
	}
	defer redisCloser()

	app := &application{
		config:  cfg,
		logger:  logger,
		models:  data.NewModels(db),
		limiter: limiter,
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorLog:     log.New(logger, "", 0),
	}

	logger.PrintInfo("starting server", map[string]string{
		"env":   cfg.env,
		"addr":  srv.Addr,
		"redis": cfg.redis.addr,
	})

	err = srv.ListenAndServe()
	logger.PrintFatal(err, nil)
}

func redisClientLimiter(cfg config) (*rate.L, func(), error) {
	client, err := establishRedisClient(cfg)
	if err != nil {
		return nil, nil, err
	}
	return rate.NewRateLimiter(client, 2, time.Second), func() { client.Close() }, err
}

func establishRedisClient(cfg config) (*redis.Client, error) {
	//TODO: maybe add repeated tries for establishing connection
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
