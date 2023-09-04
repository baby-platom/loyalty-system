package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"strings"

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

func checkIfOneStrcutFieldIsEmpty(s interface{}) string {
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

func fillUserByID(user *database.User, id uint) (err error) {
	userFilter := database.User{CustomBaseModel: database.CustomBaseModel{ID: id}}
	res := database.DB.Where(&userFilter).First(&user)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return res.Error
	}
	return err
}

// CheckLuhn validates the provided data using the Luhn algorithm.
func CheckLuhn(number string) bool {
	digits := strings.Split(strings.ReplaceAll(number, " ", ""), "")
	lengthOfString := len(digits)

	if lengthOfString < 2 {
		return false
	}

	sum := 0
	flag := false

	for i := lengthOfString - 1; i > -1; i-- {
		digit, _ := strconv.Atoi(digits[i])

		if flag {
			digit *= 2

			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		flag = !flag
	}

	return math.Mod(float64(sum), 10) == 0
}

func fillUserByRequestWithToken(r *http.Request) (user database.User, err error) {
	id := auth.GetUserIDFromRequest(r)
	if err = fillUserByID(&user, id); err != nil {
		return
	}
	return
}

func getUserOrders(r *http.Request, userID uint, logger *zap.SugaredLogger) (orders []database.Order, status int, err error) {
	filter := database.Order{UserID: userID}
	if res := database.DB.Where(&filter).Find(&orders); res.Error != nil {
		return orders, http.StatusInternalServerError, res.Error
	}

	if len(orders) == 0 {
		return orders, http.StatusNoContent, nil
	}

	return orders, http.StatusOK, nil
}

type Balance struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

func getBalance(userID uint, orders []database.Order) (balance Balance, err error) {
	filter := database.Balance{UserID: userID}
	var balanceObject database.Balance

	if res := database.DB.Where(&filter).First(&balanceObject); res.Error != nil {
		return balance, res.Error
	}

	balance.Withdrawn = balanceObject.Withdrawn
	balance.Current = balanceObject.Accumulated - balanceObject.Withdrawn
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

func createWithdrawAndUpdateBalance(user *database.User, newWithdraw database.Withdraw) (err error) {
	tx, err := database.GetTransaction()
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			err = errors.New(r.(string))
			tx.Rollback()
		}
	}()

	user.Withdrawals = append(user.Withdrawals, newWithdraw)
	if res := tx.Save(user); res.Error != nil {
		tx.Rollback()
		return res.Error
	}

	res := tx.Model(&database.Balance{}).Where(database.Balance{UserID: user.ID}).
		Update("withdrawn", gorm.Expr("withdrawn + ?", newWithdraw.Sum))
	if res.Error != nil {
		tx.Rollback()
		return res.Error
	}

	if err = database.CommitTransaction(tx); err != nil {
		return err
	}
	return nil
}
