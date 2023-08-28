package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/baby-platom/loyalty-system/internal/database"
	"github.com/baby-platom/loyalty-system/internal/logger"
)

type withdrawDataStruct struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}

func RequestWithdrawAPIHandler(w http.ResponseWriter, r *http.Request) {
	var withdraw withdrawDataStruct

	if err := json.NewDecoder(r.Body).Decode(&withdraw); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if msg := checkIfOneStrcutFieldIsEmpty(withdraw); msg != "" {
		logger.Log.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	if !Luhn([]byte(withdraw.Order)) {
		http.Error(w, "incorrect order number", http.StatusUnprocessableEntity)
		return
	}

	var user database.User
	user, err := fillUserByRequestWithToken(r)
	if err != nil {
		defaultReactionToInternalServerError(w, logger.Log, err)
		return
	}

	orders, status, err := updateOrdersAccrual(r, user.ID, logger.Log)
	switch status {
	case http.StatusInternalServerError:
		defaultReactionToInternalServerError(w, logger.Log, err)
		return
	case http.StatusNoContent:
		w.WriteHeader(http.StatusPaymentRequired)
		return
	}

	balance, err := calculateBalance(user.ID, orders)
	if err != nil {
		defaultReactionToInternalServerError(w, logger.Log, err)
		return
	}

	if available := (balance.Accumulated - balance.Withdrawn); available < withdraw.Sum {
		w.WriteHeader(http.StatusPaymentRequired)
		return
	}

	orderInt, _ := strconv.Atoi(withdraw.Order)
	newWithdraw := database.Withdraw{Order: orderInt, Sum: withdraw.Sum}
	user.Withdrawals = append(user.Withdrawals, newWithdraw)
	if err = database.DB.Save(user).Error; err != nil {
		defaultReactionToInternalServerError(w, logger.Log, err)
		return
	}
}

func GetBalanceAPIHandler(w http.ResponseWriter, r *http.Request) {
	var user database.User
	user, err := fillUserByRequestWithToken(r)
	if err != nil {
		defaultReactionToInternalServerError(w, logger.Log, err)
		return
	}

	orders, status, err := updateOrdersAccrual(r, user.ID, logger.Log)
	if status == http.StatusInternalServerError {
		defaultReactionToInternalServerError(w, logger.Log, err)
		return
	}

	balance, err := calculateBalance(user.ID, orders)
	if err != nil {
		defaultReactionToInternalServerError(w, logger.Log, err)
		return
	}

	writeResponseData(w, logger.Log, balance)
}

func ListWithdrawalsAPIHandler(w http.ResponseWriter, r *http.Request) {
	var user database.User
	user, err := fillUserByRequestWithToken(r)
	if err != nil {
		defaultReactionToInternalServerError(w, logger.Log, err)
		return
	}

	var withdrawals []database.Withdraw
	filter := database.Order{UserID: user.ID}
	if res := database.DB.Where(&filter).Find(&withdrawals); res.Error != nil {
		defaultReactionToInternalServerError(w, logger.Log, err)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	writeResponseData(w, logger.Log, withdrawals)
}
