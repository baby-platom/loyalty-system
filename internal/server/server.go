package server

import (
	"net/http"

	"github.com/baby-platom/loyalty-system/internal/auth"
	"github.com/baby-platom/loyalty-system/internal/compress"
	"github.com/baby-platom/loyalty-system/internal/config"
	"github.com/baby-platom/loyalty-system/internal/logger"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/mux"
)

func Run() error {
	router := mux.NewRouter()
	userAPIRouter := router.PathPrefix("/api").PathPrefix("/user").Subrouter()
	userAPIRouter.Use(logger.Middleware)
	userAPIRouter.Use(compress.Middleware)
	userAPIRouter.Use(jsonContentTypeMiddleware)
	userAPIRouter.Use(middleware.Compress(5))

	noAuthRouter := userAPIRouter.Methods(http.MethodPost).Subrouter()
	authRouter := userAPIRouter.Methods(http.MethodPost, http.MethodGet).Subrouter()
	authRouter.Use(auth.Middleware)

	noAuthRouter.HandleFunc("/register", UserRegisterAPIHandler)
	noAuthRouter.HandleFunc("/login", UserLoginAPIHandler)

	authRouter.HandleFunc("/orders", UploadOrderAPIHandler).Methods(http.MethodPost)
	authRouter.HandleFunc("/orders", ListUploadedOrdersAPIHandler).Methods(http.MethodGet)

	authRouter.HandleFunc("/balance", GetBalanceAPIHandler).Methods(http.MethodGet)
	authRouter.HandleFunc("/balance/withdraw", RequestWithdrawAPIHandler).Methods(http.MethodPost)
	authRouter.HandleFunc("/withdrawals", ListWithdrawalsAPIHandler).Methods(http.MethodGet)

	return http.ListenAndServe(config.Config.Address, router)
}
