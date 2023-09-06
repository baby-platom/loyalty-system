package database

import (
	"context"
	"errors"
	"net/http"
)

func GetUserOrders(ctx context.Context, r *http.Request, userID uint) (orders []Order, status int, err error) {
	filter := Order{UserID: userID}
	if res := DB.Conn(ctx).Where(&filter).Find(&orders); res.Error != nil {
		return orders, http.StatusInternalServerError, res.Error
	}

	if len(orders) == 0 {
		return orders, http.StatusNoContent, nil
	}

	return orders, http.StatusOK, nil
}

var ErrOrderAlreadyUploadedByUser = errors.New("order already was uploaded by user")
var ErrOrderAlreadyUploadedByAnotherUser = errors.New("order already was uploaded by another user")

func CreateOrder(ctx context.Context, orderNumber string, userID uint) error {
	user, err := GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	var existingOrder Order
	orderFilter := Order{Number: orderNumber}
	res := DB.Conn(ctx).Where(&orderFilter).Limit(1).Find(&existingOrder)
	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected > 0 {
		switch existingOrder.UserID {
		case userID:
			return ErrOrderAlreadyUploadedByUser
		default:
			return ErrOrderAlreadyUploadedByAnotherUser
		}
	}

	newOrder := Order{Number: orderNumber}
	user.Orders = append(user.Orders, newOrder)
	if err = DB.Conn(ctx).Save(user).Error; err != nil {
		return err
	}
	return nil
}
