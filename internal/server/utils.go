package server

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

func jsonContentTypeMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			h.ServeHTTP(w, r)
		},
	)
}

func defaultReactionToInternalServerError(w http.ResponseWriter, logger *zap.SugaredLogger, err error) {
	logger.Error(err)
	http.Error(w, "", http.StatusInternalServerError)
}

func defaultReactionToEncodingResponseError(w http.ResponseWriter, logger *zap.SugaredLogger, err error) {
	logger.Error("Error encoding response", zap.Error(err))
	http.Error(w, "Error encoding response", http.StatusInternalServerError)
}

func writeResponseData(w http.ResponseWriter, logger *zap.SugaredLogger, data any) (ok bool) {
	dataEncoded, err := json.Marshal(data)
	if err != nil {
		defaultReactionToEncodingResponseError(w, logger, err)
		return
	}
	if _, err := w.Write(dataEncoded); err != nil {
		defaultReactionToInternalServerError(w, logger, err)
		return
	}
	return true
}
