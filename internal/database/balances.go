package database

import "context"

type BalanceData struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

func GetBalanceDataByUserID(ctx context.Context, userID uint) (balance BalanceData, err error) {
	filter := Balance{UserID: userID}
	var balanceObject Balance

	if res := DB.Conn(ctx).Where(&filter).First(&balanceObject); res.Error != nil {
		return balance, res.Error
	}

	balance.Withdrawn = balanceObject.Withdrawn
	balance.Current = balanceObject.Accumulated - balanceObject.Withdrawn
	return
}

func createBalance(ctx context.Context, userID uint) (Balance, error) {
	var balance = Balance{UserID: userID}
	return balance, DB.Conn(ctx).Create(&balance).Error
}
