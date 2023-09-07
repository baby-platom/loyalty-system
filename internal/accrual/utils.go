package accrual

import (
	"github.com/baby-platom/loyalty-system/internal/database"
)

var OrderStatusByAccrual = map[string]database.OrderStatus{
	"REGISTERED": database.NEW,
	"PROCESSING": database.PROCESSING,
	"INVALID":    database.INVALID,
	"PROCESSED":  database.PROCESSED,
}
