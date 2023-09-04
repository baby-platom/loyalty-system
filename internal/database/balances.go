package database

type BalanceData struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

func GetBalance(userID uint, orders []Order) (balance BalanceData, err error) {
	filter := Balance{UserID: userID}
	var balanceObject Balance

	if res := DB.Where(&filter).First(&balanceObject); res.Error != nil {
		return balance, res.Error
	}

	balance.Withdrawn = balanceObject.Withdrawn
	balance.Current = balanceObject.Accumulated - balanceObject.Withdrawn
	return
}
