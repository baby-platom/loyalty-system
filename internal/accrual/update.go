package accrual

import (
	"github.com/baby-platom/loyalty-system/internal/config"
	"github.com/baby-platom/loyalty-system/internal/database"
)

func GetOrdersCopyWithUpdatedFields(orders []database.Order) (ordersCopy []database.Order, changed bool) {
	ordersCopy = make([]database.Order, len(orders))
	copy(ordersCopy, orders)

	if config.Config.Local {
		for i, orderCopy := range ordersCopy {
			if orderCopy.Status == database.NEW || orderCopy.Status == database.PROCESSING {
				orderData, err := GetInfoAboutOrder(orderCopy.Number)
				if err == nil && orderData != (OrderData{}) {
					orderCopy.Status = OrderStatusByAccrual[orderData.Status]
					if orderCopy.Status == database.PROCESSED {
						orderCopy.Accrual = orderData.Accrual
						changed = true
					}

					if orders[i].Status != orderCopy.Status {
						changed = true
					}
				}
			}
			ordersCopy[i] = orderCopy
		}
	}
	return
}

func UpdateOrdersObjects(orders []database.Order) (err error) {
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
		if err = tx.Save(order).Error; err != nil {
			tx.Rollback()
			return
		}
	}

	if err = tx.Commit().Error; err != nil {
		return
	}
	return
}
