package accrual

import (
	"net/http"
	"strconv"

	"github.com/baby-platom/loyalty-system/internal/config"
	"github.com/baby-platom/loyalty-system/internal/logger"
	"github.com/go-resty/resty/v2"
)

type OrderData struct {
	Number  int     `json:"number"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

var client = resty.New()

func GetInfoAboutOrder(number int) (result OrderData, err error) {
	resp, err := client.R().
		SetResult(&result).
		SetPathParams(
			map[string]string{
				"accrualSystemAdress": config.Config.AccrualSystemAdress,
				"orderNumber":         strconv.Itoa(number),
			},
		).
		Get("{accrualSystemAdress}/api/orders/{orderNumber}")
	if err != nil {
		logger.Log.Infof("Cannot make a GET request to '%s'", resp.Request.URL)
		return result, err
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return result, nil
	case http.StatusNoContent:
		return
	case http.StatusTooManyRequests:
		logger.Log.Warn("too many requests to accrual service")
		return
	case http.StatusInternalServerError:
		logger.Log.Error("internal server error occured in accrual service by address '%s'", config.Config.AccrualSystemAdress)
		return
	}
	return
}
