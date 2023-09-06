package accrual

import (
	"context"
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

func getUpdatedOrder(outCh chan UpdatedOrder, cancelCh chan struct{}, order database.Order) {
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

	outCh <- UpdatedOrder{order: orderCopy, changed: changed}
}

func UpdateOrders(ctx context.Context) error {
	var orders []database.Order
	res := database.DB.Conn(ctx).Where(database.Order{Status: database.NEW}).
		Or(database.Order{Status: database.PROCESSING}).
		Find(&orders)

	if res.Error != nil {
		logger.Log.Error("error while getting orders to update", zap.Error(res.Error))
		return res.Error
	}

	outCh := make(chan UpdatedOrder)
	cancelCh := make(chan struct{})
	updatedOrdersCounter := 0
	for _, order := range orders {
		go getUpdatedOrder(outCh, cancelCh, order)
		updatedOrdersCounter += 1
	}

	ordersToUpdate := make([]database.Order, 0)
	ordersWithAccrual := make([]database.Order, 0)
	for i := 0; i < updatedOrdersCounter; i++ {
		updatedOrder := <-outCh
		if updatedOrder.changed {
			ordersToUpdate = append(ordersToUpdate, updatedOrder.order)
			if updatedOrder.order.Accrual != 0 {
				ordersWithAccrual = append(ordersWithAccrual, updatedOrder.order)
			}
		}
	}

	if err := updateOrdersAndBalances(ctx, ordersToUpdate, ordersWithAccrual); err != nil {
		logger.Log.Error("error while updating orders and balances", zap.Error(err))
		return err
	}
	return nil
}

func updateOrdersAndBalances(ctx context.Context, ordersToUpdate, ordersWithAccrual []database.Order) error {
	for _, order := range ordersToUpdate {
		if err := database.DB.Conn(ctx).Save(order).Error; err != nil {
			logger.Log.Error(err)
			return err
		}
	}

	for _, order := range ordersWithAccrual {
		res := database.DB.Conn(ctx).Model(&database.Balance{}).Where(database.Balance{UserID: order.UserID}).
			Update("accumulated", gorm.Expr("accumulated + ?", order.Accrual))

		if res.Error != nil {
			logger.Log.Error(res.Error)
			return res.Error
		}
	}

	return nil
}
