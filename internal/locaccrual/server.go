package locaccrual

import (
	"net/http"

	"github.com/baby-platom/loyalty-system/internal/logger"
	"github.com/gorilla/mux"
)

const LocalAccrualAdress = "localhost:8081"

func Run() error {
	router := mux.NewRouter()
	ordersAPIRouter := router.PathPrefix("/api").PathPrefix("/orders").Subrouter()
	ordersAPIRouter.Use(logger.Middleware)

	ordersAPIRouter.HandleFunc("/{number}", OrderInfoAPIHandler)

	return http.ListenAndServe(LocalAccrualAdress, router)
}
