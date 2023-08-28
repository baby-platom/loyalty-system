package main

import (
	"github.com/baby-platom/loyalty-system/internal/accrual"
	"github.com/baby-platom/loyalty-system/internal/config"
	"github.com/baby-platom/loyalty-system/internal/database"
	"github.com/baby-platom/loyalty-system/internal/local_accrual"
	"github.com/baby-platom/loyalty-system/internal/logger"
	"github.com/baby-platom/loyalty-system/internal/server"
)

func main() {
	config.ParseFlags()
	if err := logger.Initialize(config.Config.LogLevel); err != nil {
		panic(err)
	}
	database.Prepare()
	accrual.PrepareAddress()

	if config.Config.Local {
		logger.Log.Info("local accrual address in use")
		go func() {
			if err := local_accrual.Run(); err != nil {
				panic(err)
			}
		}()
	}

	if err := server.Run(); err != nil {
		panic(err)
	}
}
