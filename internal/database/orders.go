package database

import (
	"net/http"

	"go.uber.org/zap"
)

func GetUserOrders(r *http.Request, userID uint, logger *zap.SugaredLogger) (orders []Order, status int, err error) {
	filter := Order{UserID: userID}
	if res := DB.Where(&filter).Find(&orders); res.Error != nil {
		return orders, http.StatusInternalServerError, res.Error
	}

	if len(orders) == 0 {
		return orders, http.StatusNoContent, nil
	}

	return orders, http.StatusOK, nil
}
