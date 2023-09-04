package accrual

import (
	"errors"

	"github.com/baby-platom/loyalty-system/internal/database"
	"github.com/baby-platom/loyalty-system/internal/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UpdatedOrder struct {
	order   database.Order
	changed bool
}

func getUpdatedOrder(cancelCh chan struct{}, order database.Order) chan UpdatedOrder {
	out := make(chan UpdatedOrder)

	go func() {
		defer close(out)
		orderCopy := order
		changed := false

		select {
		case <-cancelCh:
			break
		default:
			orderData, err := GetInfoAboutOrder(orderCopy.Number)
			if errors.Is(err, errManyRequestsError) {
				close(cancelCh)
				break
			}

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

	cancelCh := make(chan struct{})
	for _, order := range orders {
		outChan := getUpdatedOrder(cancelCh, order)
		outChannels = append(outChannels, outChan)
	}

	finalCh := make(chan UpdatedOrder)
	fanIn(finalCh, outChannels...)

	ordersToUpdate := make([]database.Order, 0)
	ordersWithAccrual := make([]database.Order, 0)
	for updatedOrder := range finalCh {
		if updatedOrder.changed {
			ordersToUpdate = append(ordersToUpdate, updatedOrder.order)
			if updatedOrder.order.Accrual != 0 {
				ordersWithAccrual = append(ordersWithAccrual, updatedOrder.order)
			}
		}
	}

	if err := updateOrdersAndBalances(ordersToUpdate, ordersWithAccrual); err != nil {
		logger.Log.Error("error while updating orders and balances", zap.Error(err))
	}
}

func updateOrdersAndBalances(ordersToUpdate, ordersWithAccrual []database.Order) (err error) {
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

	if err = updateOrders(tx, ordersToUpdate); err != nil {
		tx.Rollback()
		return err
	}

	if err = updateBalancecFromOrders(tx, ordersWithAccrual); err != nil {
		tx.Rollback()
		return err
	}

	if err = database.CommitTransaction(tx); err != nil {
		return err
	}
	return nil
}

func updateOrders(tx *gorm.DB, ordersToUpdate []database.Order) error {
	for _, order := range ordersToUpdate {
		if err := tx.Save(order).Error; err != nil {
			logger.Log.Error(err)
			return err
		}
	}
	return nil
}

func updateBalancecFromOrders(tx *gorm.DB, ordersWithAccrual []database.Order) error {
	for _, order := range ordersWithAccrual {
		if err := updateBalanceFromOrder(tx, order); err != nil {
			logger.Log.Error(err)
			return err
		}
	}
	return nil
}
