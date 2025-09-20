package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) Routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)

	return router
}
