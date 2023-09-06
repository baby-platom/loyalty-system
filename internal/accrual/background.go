package accrual

import (
	"context"
	"time"

	"github.com/baby-platom/loyalty-system/internal/database"
	"github.com/baby-platom/loyalty-system/internal/logger"
	"go.uber.org/zap"
)

func UpdateOrdersInBackground() {
	ctx := context.Background()
	ticker := time.NewTicker(10 * time.Second)

	for {
		<-ticker.C
		logger.Log.Debug("Updating orders")
		if err := database.DB.WithinTransaction(ctx, UpdateOrders); err != nil {
			logger.Log.Errorf("error occured while updating orders", zap.Error(err))
		}
		logger.Log.Debug("Updated orders")
	}
}
