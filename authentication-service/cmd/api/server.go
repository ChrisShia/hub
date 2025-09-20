package main

import (
	"fmt"

	srv "github.com/ChrisShia/serve"
)

func (app *application) serve() error {
	return srv.ListenAndServe(app, app.config.port)
}

func (app *application) LogStartUp() {
	app.logger.PrintInfo("starting application server", map[string]string{
		"port":  fmt.Sprintf(":%d", app.config.port),
		"env":   app.config.env,
		"redis": app.config.redis.addr,
	})
}

func (app *application) LogShutdown() {
	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": fmt.Sprintf(":%d", app.config.port),
	})
}

func (app *application) PrintInfo(msg string, properties map[string]string) {
	app.logger.PrintInfo(msg, properties)
}

func (app *application) Write(p []byte) (n int, err error) {
	return app.logger.Write(p)
}
