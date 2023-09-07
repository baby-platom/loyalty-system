package auth

import (
	"context"
	"net/http"

	"github.com/baby-platom/loyalty-system/internal/logger"
)

type key string

const UserID key = "userID"

// Middleware for authentification
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

		id, err := GetUserIDFromToken(authCookie.Value)
		if err != nil {
			logger.Log.Warnw(
				"Error occured while parsing auth cookie",
				"error", err,
			)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserID, id)
		h.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

func GetUserIDFromRequest(r *http.Request) uint {
	return r.Context().Value(UserID).(uint)
}
