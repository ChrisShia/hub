package main

import (
	"authentication-service/internal/data"
	"os"

	"github.com/ChrisShia/jsonlog"
)

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
