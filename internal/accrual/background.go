package accrual

import (
	"time"

	"github.com/baby-platom/loyalty-system/internal/logger"
)

func UpdateOrdersInBackground() {
	ticker := time.NewTicker(10 * time.Second)

	for {
		<-ticker.C
		logger.Log.Info("Updating orders")
		UpdateOrders()
		logger.Log.Info("Updated orders")
	}
}
