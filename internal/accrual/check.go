package accrual

import (
	"fmt"
	"net/http"

	"github.com/baby-platom/loyalty-system/internal/config"
	"github.com/baby-platom/loyalty-system/internal/locaccrual"
	"github.com/baby-platom/loyalty-system/internal/logger"
	"github.com/go-resty/resty/v2"
)

type OrderData struct {
	Number  string  `json:"number"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual"`
}

var client = resty.New()
var address string

func PrepareAddress() {
	if config.Config.Local {
		address = fmt.Sprintf("http://%s", locaccrual.LocalAccrualAdress)
		return
	}
	address = config.Config.AccrualSystemAdress
}

func GetInfoAboutOrder(number string) (result OrderData, err error) {
	resp, err := client.
		SetBaseURL(address).
		R().
		SetResult(&result).
		SetPathParams(
			map[string]string{
				"orderNumber": number,
			},
		).
		Get("/api/orders/{orderNumber}")
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
		logger.Log.Info("too many requests to accrual service")
		return
	case http.StatusInternalServerError:
		logger.Log.Infof("internal server error occured in accrual service by address '%s'", address)
		return
	}
	return
}
