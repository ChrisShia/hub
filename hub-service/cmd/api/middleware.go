package main

import (
	"fmt"
	"net"
	"net/http"

	rl "github.com/ChrisShia/ratelimiter"
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
		if allow, err := allowBasedOnLimiter(r, app.limiter); !allow {
			if err != nil {
				app.serverErrorResponse(w, r, err)
			} else {
				app.rateLimitExceededResponse(w, r)
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

func allowBasedOnLimiter(r *http.Request, limiter *rl.Limiter) (bool, error) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return false, err
	}

	return limiter.Allow(ip)
}
