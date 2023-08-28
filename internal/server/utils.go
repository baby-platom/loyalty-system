package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/baby-platom/loyalty-system/internal/accrual"
	"github.com/baby-platom/loyalty-system/internal/auth"
	"github.com/baby-platom/loyalty-system/internal/database"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func jsonContentTypeMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			h.ServeHTTP(w, r)
		},
	)
}

func checkIfOneStrcutFieldIsEmpty(s any) string {
	v := reflect.ValueOf(s)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field, typeOfField := v.Field(i), t.Field(i)
		if field.IsZero() {
			return fmt.Sprintf("field %s is empty", typeOfField.Name)
		}
	}
	return ""
}

func defaultReactionToInternalServerError(w http.ResponseWriter, logger *zap.SugaredLogger, err error) {
	logger.Error(err)
	http.Error(w, "", http.StatusInternalServerError)
}

func defaultReactionToEncodingResponseError(w http.ResponseWriter, logger *zap.SugaredLogger, err error) {
	logger.Error("Error encoding response", zap.Error(err))
	http.Error(w, "Error encoding response", http.StatusInternalServerError)
}

func fillUserByLogin(user *database.User, userLogin string) (err error) {
	userFilter := database.User{Login: userLogin}
	res := database.DB.Where(&userFilter).First(&user)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return res.Error
	}
	return err
}

// Luhn validates the provided data using the Luhn algorithm.
func Luhn(s []byte) bool {
	n := len(s)
	number := 0
	result := 0
	for i := 0; i < n; i++ {
		number = int(s[i]) - '0'
		if i%2 != 0 {
			result += number
			continue
		}
		number *= 2
		if number > 9 {
			number -= 9
		}
		result += number
	}
	return result%10 == 0
}

func fillUserByRequestWithToken(r *http.Request) (user database.User, err error) {
	userLogin := auth.GetUserLogin(r)
	if err = fillUserByLogin(&user, userLogin); err != nil {
		return
	}
	return
}

func updateOrdersAccrual(r *http.Request, userID uint, logger *zap.SugaredLogger) (ordersCopy []database.Order, status int, err error) {
	filter := database.Order{UserID: userID}
	var orders []database.Order
	if res := database.DB.Where(&filter).Find(&orders); res.Error != nil {
		return ordersCopy, http.StatusInternalServerError, res.Error
	}

	if len(orders) == 0 {
		return ordersCopy, http.StatusNoContent, nil
	}

	ordersCopy, changed := accrual.GetOrdersCopyWithUpdatedFields(orders)
	if changed {
		if err = accrual.UpdateOrdersObjects(ordersCopy); err != nil {
			return ordersCopy, http.StatusInternalServerError, err
		}
	}
	return ordersCopy, http.StatusOK, nil
}

type Balance struct {
	Accumulated float64 `json:"current"`
	Withdrawn   float64 `json:"withdrawn"`
}

func calculateBalance(userID uint, orders []database.Order) (balance Balance, err error) {
	filter := database.Withdraw{UserID: userID}
	var withdrawals []database.Withdraw
	if res := database.DB.Where(&filter).Find(&withdrawals); res.Error != nil {
		return balance, res.Error
	}

	var withdrawn float64
	for _, withdraw := range withdrawals {
		withdrawn += withdraw.Sum
	}

	var accumulated float64
	for _, order := range orders {
		accumulated += order.Accrual
	}

	balance.Accumulated, balance.Withdrawn = accumulated, withdrawn
	return
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
