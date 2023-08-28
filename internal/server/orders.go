package server

import (
	"io"
	"net/http"
	"strings"

	"github.com/baby-platom/loyalty-system/internal/auth"
	"github.com/baby-platom/loyalty-system/internal/database"
	"github.com/baby-platom/loyalty-system/internal/logger"
	"gorm.io/gorm"
)

func UploadOrderAPIHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(string(body)) == "" {
		http.Error(w, "body is empty", http.StatusBadRequest)
		return
	}

	if !Luhn(body) {
		http.Error(w, "incorrect order number", http.StatusUnprocessableEntity)
		return
	}
	orderNumber := string(body)

	userLogin := auth.GetUserLogin(r)
	var user database.User
	if err := fillUserByLogin(&user, userLogin); err != nil {
		defaultReactionToInternalServerError(w, logger.Log, err)
		return
	}

	var existingOrder database.Order
	orderFilter := database.Order{Number: orderNumber}
	var res *gorm.DB
	if res = database.DB.Where(&orderFilter).Limit(1).Find(&existingOrder); res.Error != nil {
		defaultReactionToInternalServerError(w, logger.Log, res.Error)
		return
	}

	if res.RowsAffected > 0 {
		switch existingOrder.UserID {
		case user.ID:
			http.Error(w, "order already was uploaded by you", http.StatusOK)
		default:
			http.Error(w, "order already was uploaded by another user", http.StatusConflict)
		}
		return
	}

	newOrder := database.Order{Number: orderNumber}
	user.Orders = append(user.Orders, newOrder)
	if err = database.DB.Save(user).Error; err != nil {
		defaultReactionToInternalServerError(w, logger.Log, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func ListUploadedOrdersAPIHandler(w http.ResponseWriter, r *http.Request) {
	var user database.User
	user, err := fillUserByRequestWithToken(r)
	if err != nil {
		defaultReactionToInternalServerError(w, logger.Log, err)
		return
	}

	orders, status, err := updateOrdersAccrual(r, user.ID, logger.Log)
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
