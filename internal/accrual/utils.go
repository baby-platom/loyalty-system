package accrual

import (
	"sync"

	"github.com/baby-platom/loyalty-system/internal/database"
)

var OrderStatusByAccrual = map[string]database.OrderStatus{
	"REGISTERED": database.NEW,
	"PROCESSING": database.PROCESSING,
	"INVALID":    database.INVALID,
	"PROCESSED":  database.PROCESSED,
}

func fanIn(finalCh chan UpdatedOrder, resultChs ...chan UpdatedOrder) {
	var wg sync.WaitGroup

	for _, ch := range resultChs {
		chClosure := ch
		wg.Add(1)

		go func() {
			defer wg.Done()

			for data := range chClosure {
				finalCh <- data
			}
		}()
	}

	go func() {
		wg.Wait()
		close(finalCh)
	}()
}
