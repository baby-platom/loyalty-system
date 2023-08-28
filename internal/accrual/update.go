package accrual

import (
	"sync"

	"github.com/baby-platom/loyalty-system/internal/database"
)

func updateOrderCopy(orderCopy database.Order, order database.Order, ordersCopy *[]database.Order, changed *bool, wg *sync.WaitGroup) {
	orderData, err := GetInfoAboutOrder(orderCopy.Number)
	if err == nil && orderData != (OrderData{}) {
		orderCopy.Status = OrderStatusByAccrual[orderData.Status]
		if orderCopy.Status == database.PROCESSED {
			orderCopy.Accrual = orderData.Accrual
			*changed = true
		}

		if order.Status != orderCopy.Status {
			*changed = true
		}
	}

	*ordersCopy = append(*ordersCopy, orderCopy)
	wg.Done()
}

func GetOrdersCopyWithUpdatedFields(orders []database.Order) (ordersCopy []database.Order, changed bool) {
	ordersCopy = make([]database.Order, len(orders))
	copy(ordersCopy, orders)

	var wg sync.WaitGroup
	for i, orderCopy := range ordersCopy {
		if orderCopy.Status == database.NEW || orderCopy.Status == database.PROCESSING {
			wg.Add(1)
			go updateOrderCopy(orderCopy, orders[i], &ordersCopy, &changed, &wg)
		}
	}
	wg.Wait()
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
