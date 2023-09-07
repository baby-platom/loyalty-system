package database

import (
	"context"

	"gorm.io/gorm"
)

func CreateWithdrawAndUpdateBalance(ctx context.Context, userID uint, order string, sum float32) (err error) {
	user := User{CustomBaseModel: CustomBaseModel{ID: userID}}
	withdraw := Withdraw{Order: order, Sum: sum}

	user.Withdrawals = append(user.Withdrawals, withdraw)
	if res := DB.Conn(ctx).Save(&user); res.Error != nil {
		return res.Error
	}

	res := DB.Conn(ctx).Model(Balance{}).Where(Balance{UserID: user.ID}).
		Update("withdrawn", gorm.Expr("withdrawn + ?", withdraw.Sum))
	if res.Error != nil {
		return res.Error
	}

	return
}

func GetUserWithdrawals(ctx context.Context, userID uint) (withdrawals []Withdraw, err error) {
	filter := Order{UserID: userID}
	if res := DB.Conn(ctx).Where(&filter).Find(&withdrawals); res.Error != nil {
		return withdrawals, res.Error
	}
	return
}
