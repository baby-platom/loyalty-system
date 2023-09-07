package server

import (
	"io"
	"net/http"
	"strings"

	"github.com/baby-platom/loyalty-system/internal/auth"
	"github.com/baby-platom/loyalty-system/internal/database"
	"github.com/baby-platom/loyalty-system/internal/logger"
	"github.com/baby-platom/loyalty-system/internal/luhn"
)

func UploadOrderAPIHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orderNumber := string(body)
	if strings.TrimSpace(orderNumber) == "" {
		http.Error(w, "body is empty", http.StatusBadRequest)
		return
	}

	if !luhn.CheckLuhn(orderNumber) {
		http.Error(w, "incorrect order number", http.StatusUnprocessableEntity)
		return
	}

	userID := auth.GetUserIDFromRequest(r)
	if err = database.CreateOrder(r.Context(), orderNumber, userID); err != nil {
		switch err {
		case database.ErrOrderAlreadyUploadedByUser:
			http.Error(w, "order already was uploaded by you", http.StatusOK)
		case database.ErrOrderAlreadyUploadedByAnotherUser:
			http.Error(w, "order already was uploaded by another user", http.StatusConflict)
		default:
			defaultReactionToInternalServerError(w, logger.Log, err)
		}
	}

	w.WriteHeader(http.StatusAccepted)
}

func ListUploadedOrdersAPIHandler(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromRequest(r)

	orders, status, err := database.GetUserOrders(r.Context(), r, userID)
	switch status {
	case http.StatusNoContent:
		w.WriteHeader(status)
		return
	case http.StatusInternalServerError:
		defaultReactionToInternalServerError(w, logger.Log, err)
		return
	}

	writeResponseData(w, logger.Log, orders)
}
