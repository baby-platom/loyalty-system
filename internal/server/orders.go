package server

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/baby-platom/loyalty-system/internal/accrual"
	"github.com/baby-platom/loyalty-system/internal/auth"
	"github.com/baby-platom/loyalty-system/internal/config"
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
	orderNumber := int(binary.BigEndian.Uint64(body))

	userLogin := auth.GetUserLogin(r)
	var user database.User
	fillUserByLogin(w, logger.Log, &user, userLogin)

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

func getOrdersCopyWithUpdatedFields(orders []database.Order) (ordersCopy []database.Order, changed bool) {
	ordersCopy = make([]database.Order, len(orders))
	copy(ordersCopy, orders)

	if config.Config.Local {
		for i, orderCopy := range ordersCopy {
			if orderCopy.Status == database.NEW || orderCopy.Status == database.PROCESSING {
				orderData, err := accrual.GetInfoAboutOrder(orderCopy.Number)
				if err == nil && orderData != (accrual.OrderData{}) {
					orderCopy.Status = database.OrderStatus(orderData.Status)
					if orderCopy.Status == database.PROCESSED {
						orderCopy.Accrual = orderData.Accrual
						changed = true
					}

					if orders[i].Status != orderCopy.Status {
						changed = true
					}
				}
			}
		}
	}
	return
}

func updateOrdersObjects(orders []database.Order) (err error) {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err = tx.Error; err != nil {
		return
	}

	for _, order := range orders {
		if err = tx.Model(&database.Order{}).Updates(order).Error; err != nil {
			tx.Rollback()
			return
		}
	}

	if err = tx.Commit().Error; err != nil {
		return
	}
	return
}

func ListUploadedOrdersAPIHandler(w http.ResponseWriter, r *http.Request) {
	userLogin := auth.GetUserLogin(r)
	var user database.User
	fillUserByLogin(w, logger.Log, &user, userLogin)

	filter := database.Order{UserID: user.ID}
	var orders []database.Order
	if res := database.DB.Where(&filter).Find(&orders); res.Error != nil {
		defaultReactionToInternalServerError(w, logger.Log, res.Error)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	ordersCopy, changed := getOrdersCopyWithUpdatedFields(orders)
	if changed {
		if err := updateOrdersObjects(ordersCopy); err != nil {
			defaultReactionToInternalServerError(w, logger.Log, err)
			return
		}
	}

	data, err := json.Marshal(ordersCopy)
	if err != nil {
		defaultReactionToEncodingResponseError(w, logger.Log, err)
		return
	}
	if _, err := w.Write(data); err != nil {
		defaultReactionToInternalServerError(w, logger.Log, err)
		return
	}
}
