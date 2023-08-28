package local_accrual

import (
	"math/rand"
	"net/http"
	"time"
)

type OrderStatus string

const (
	REGISTERED OrderStatus = "REGISTERED"
	PROCESSING OrderStatus = "PROCESSING"
	INVALID    OrderStatus = "INVALID"
	PROCESSED  OrderStatus = "PROCESSED"
)

const (
	minAccrual int = 50
	maxAccrual int = 500
)

var r = rand.New(rand.NewSource(time.Now().Unix()))
var httpStatusOptions = []int{http.StatusNoContent, http.StatusTooManyRequests, http.StatusInternalServerError}
var orderStatusOptions = []OrderStatus{REGISTERED, PROCESSING, INVALID}

func init() {
	for i := 0; i < 7; i++ {
		httpStatusOptions = append(httpStatusOptions, http.StatusOK)
	}
	for i := 0; i < 7; i++ {
		orderStatusOptions = append(orderStatusOptions, PROCESSED)
	}
}

func getRandomHttpStatus() int {
	return httpStatusOptions[r.Intn(len(httpStatusOptions))]
}

type OrderInfo struct {
	Order   string      `json:"order"`
	Status  OrderStatus `json:"status"`
	Accrual float64     `json:"accrual,omitempty"`
}

func getRandomOrderInfo(order string) (orderInfo OrderInfo) {
	orderInfo.Order = order
	orderInfo.Status = orderStatusOptions[r.Intn(len(orderStatusOptions))]
	if orderInfo.Status == PROCESSED {
		orderInfo.Accrual = float64(minAccrual + r.Intn(maxAccrual))
	}
	return orderInfo
}
