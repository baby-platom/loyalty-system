package local_accrual

import (
	"net/http"

	"github.com/baby-platom/loyalty-system/internal/logger"
	"github.com/gorilla/mux"
)

const LocalAccrualAdress = "localhost:8081"

func Run() error {
	router := mux.NewRouter()
	ordersApiRouter := router.PathPrefix("/api").PathPrefix("/orders").Subrouter()
	ordersApiRouter.Use(logger.Middleware)

	ordersApiRouter.HandleFunc("/{number}", OrderInfoAPIHandler)

	return http.ListenAndServe(LocalAccrualAdress, router)
}
