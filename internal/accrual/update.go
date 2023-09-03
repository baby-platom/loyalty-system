package accrual

import (
	"github.com/baby-platom/loyalty-system/internal/database"
	"github.com/baby-platom/loyalty-system/internal/logger"
	"go.uber.org/zap"
)

type UpdatedOrder struct {
	order   database.Order
	changed bool
}

func getUpdatedOrder(order database.Order) chan UpdatedOrder {
	out := make(chan UpdatedOrder)

	go func() {
		defer close(out)

		orderCopy := order
		changed := false
		orderData, err := GetInfoAboutOrder(orderCopy.Number)
		if err == nil && orderData != (OrderData{}) {
			orderCopy.Status = OrderStatusByAccrual[orderData.Status]
			if orderCopy.Status == database.PROCESSED {
				orderCopy.Accrual = orderData.Accrual
				changed = true
			}

			if order.Status != orderCopy.Status {
				changed = true
			}
		}

		out <- UpdatedOrder{order: orderCopy, changed: changed}
	}()

	return out
}

func UpdateOrders() {
	var orders []database.Order
	res := database.DB.Where(database.Order{Status: database.NEW}).
		Or(database.Order{Status: database.PROCESSING}).
		Find(&orders)

	if res.Error != nil {
		logger.Log.Error("error while getting orders to update", zap.Error(res.Error))
		return
	}

	outChannels := make([]chan UpdatedOrder, 0)

	for _, order := range orders {
		outChan := getUpdatedOrder(order)
		outChannels = append(outChannels, outChan)
	}

	finalCh := make(chan UpdatedOrder)
	fanIn(finalCh, outChannels...)

	ordersToUpdate := make([]database.Order, 0)
	for updatedOrder := range finalCh {
		if updatedOrder.changed {
			ordersToUpdate = append(ordersToUpdate, updatedOrder.order)
		}
	}

	if err := UpdateOrdersObjects(ordersToUpdate); err != nil {
		logger.Log.Error("error while getting orders to update", zap.Error(err))
	}
}

func UpdateOrdersObjects(orders []database.Order) (err error) {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err = tx.Error; err != nil {
		logger.Log.Error(err)
		return
	}

	for _, order := range orders {
		if err = tx.Save(order).Error; err != nil {
			logger.Log.Error(err)
			tx.Rollback()
			return
		}
	}

	if err = tx.Commit().Error; err != nil {
		logger.Log.Error(err)
		return
	}
	return
}
