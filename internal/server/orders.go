package server

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/baby-platom/loyalty-system/internal/auth"
	"github.com/baby-platom/loyalty-system/internal/database"
	"github.com/baby-platom/loyalty-system/internal/logger"
	"github.com/baby-platom/loyalty-system/internal/luhn"
	"gorm.io/gorm"
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
	var user database.User
	if err := database.FillUserByID(&user, userID); err != nil {
		defaultReactionToInternalServerError(w, logger.Log, err)
		return
	}

	tx, err := database.GetTransaction()
	if err != nil {
		defaultReactionToInternalServerError(w, logger.Log, err)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			err = errors.New(r.(string))
			tx.Rollback()
			defaultReactionToInternalServerError(w, logger.Log, err)
		}
	}()

	var existingOrder database.Order
	orderFilter := database.Order{Number: orderNumber}
	var res *gorm.DB
	if res = tx.Where(&orderFilter).Limit(1).Find(&existingOrder); res.Error != nil {
		tx.Rollback()
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
	if err = tx.Save(user).Error; err != nil {
		tx.Rollback()
		defaultReactionToInternalServerError(w, logger.Log, err)
		return
	}

	if err = database.CommitTransaction(tx); err != nil {
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

	orders, status, err := database.GetUserOrders(r, user.ID, logger.Log)
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
