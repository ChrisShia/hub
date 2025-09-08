package main

import (
	"fmt"
	"net"
	"net/http"

	"github.com/ChrisShia/moviehub/internal/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if allow, err := requestLimit(r, app.limiter); !allow {
			if err != nil {
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

func requestLimit(r *http.Request, limiter *rate.L) (bool, error) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return false, err
	}

	return limiter.Allow(ip)
}
