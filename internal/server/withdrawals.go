package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/baby-platom/loyalty-system/internal/auth"
	"github.com/baby-platom/loyalty-system/internal/database"
	"github.com/baby-platom/loyalty-system/internal/logger"
	"github.com/baby-platom/loyalty-system/internal/luhn"
	"github.com/baby-platom/loyalty-system/internal/reflect"
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

	if msg := reflect.CheckIfOneStrcutFieldIsEmpty(withdraw); msg != "" {
		logger.Log.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	if !luhn.CheckLuhn(withdraw.Order) {
		http.Error(w, "incorrect order number", http.StatusUnprocessableEntity)
		return
	}

	userID := auth.GetUserIDFromRequest(r)
	balanceData, err := database.GetBalanceDataByUserID(r.Context(), userID)
	if err != nil {
		defaultReactionToInternalServerError(w, logger.Log, err)
		return
	}

	if balanceData.Current < withdraw.Sum {
		w.WriteHeader(http.StatusPaymentRequired)
		return
	}

	err = database.DB.WithinTransaction(
		r.Context(),
		func(ctx context.Context) error {
			return database.CreateWithdrawAndUpdateBalance(r.Context(), userID, withdraw.Order, withdraw.Sum)
		},
	)

	if err != nil {
		defaultReactionToInternalServerError(w, logger.Log, err)
		return
	}
}

func GetBalanceAPIHandler(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromRequest(r)

	balanceData, err := database.GetBalanceDataByUserID(r.Context(), userID)
	if err != nil {
		defaultReactionToInternalServerError(w, logger.Log, err)
		return
	}

	writeResponseData(w, logger.Log, balanceData)
}

func ListWithdrawalsAPIHandler(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromRequest(r)

	withdrawals, err := database.GetUserWithdrawals(r.Context(), userID)
	if err != nil {
		defaultReactionToInternalServerError(w, logger.Log, err)
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	writeResponseData(w, logger.Log, withdrawals)
}
