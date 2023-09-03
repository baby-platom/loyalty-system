package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/baby-platom/loyalty-system/internal/accrual"
	"github.com/baby-platom/loyalty-system/internal/auth"
	"github.com/baby-platom/loyalty-system/internal/compress"
	"github.com/baby-platom/loyalty-system/internal/config"
	"github.com/baby-platom/loyalty-system/internal/logger"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/mux"
)

func prepareServer() http.Server {
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

	return http.Server{
		Addr:    config.Config.Address,
		Handler: router,
	}
}

func Run(shutdownFuncs ...func()) {
	server := prepareServer()
	go accrual.UpdateOrdersInBackground()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatalf("listen: %s\n", err)
		}
	}()
	logger.Log.Info("Server Started")

	<-done
	logger.Log.Info("Server Is Stopping")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		for _, f := range shutdownFuncs {
			f()
		}
		cancel()
	}()

	if err := server.Shutdown(ctx); err != nil {
		logger.Log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	logger.Log.Info("Server Exited Properly")
}
