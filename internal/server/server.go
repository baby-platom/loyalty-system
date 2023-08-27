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
	userApiRouter := router.PathPrefix("/api").PathPrefix("/user").Subrouter()
	userApiRouter.Use(logger.Middleware)
	userApiRouter.Use(compress.Middleware)
	userApiRouter.Use(jsonContentTypeMiddleware)
	userApiRouter.Use(middleware.Compress(5))

	noAuthRouter := userApiRouter.Methods(http.MethodPost).Subrouter()
	authRouter := userApiRouter.Methods(http.MethodPost, http.MethodGet).Subrouter()
	authRouter.Use(auth.Middleware)

	noAuthRouter.HandleFunc("/register", UserRegisterAPIHandler)
	noAuthRouter.HandleFunc("/login", UserLoginAPIHandler)

	authRouter.HandleFunc("/orders", UploadOrderAPIHandler).Methods(http.MethodPost)
	authRouter.HandleFunc("/orders", ListUploadedOrdersAPIHandler).Methods(http.MethodGet)

	return http.ListenAndServe(config.Config.Address, router)
}
