package auth

import (
	"context"
	"net/http"

	"github.com/baby-platom/loyalty-system/internal/logger"
)

type key int

const UserLogin key = 0

// Middleware for compression and decompression
func Middleware(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		authCookie, err := r.Cookie("auth")
		if err != nil {
			if err == http.ErrNoCookie {
				logger.Log.Info("No auth cookie passed")
			} else {
				logger.Log.Errorw(
					"Unexpected error occured while getting auth cookie",
					"error", err,
				)
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		login, err := GetLogin(authCookie.Value)
		if err != nil {
			logger.Log.Warnw(
				"Error occured while parsing auth cookie",
				"error", err,
			)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserLogin, login)
		h.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

func GetUserLogin(r *http.Request) string {
	return r.Context().Value(UserLogin).(string)
}
