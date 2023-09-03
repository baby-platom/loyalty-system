package main

import (
	"github.com/baby-platom/loyalty-system/internal/accrual"
	"github.com/baby-platom/loyalty-system/internal/config"
	"github.com/baby-platom/loyalty-system/internal/database"
	"github.com/baby-platom/loyalty-system/internal/logger"
	"github.com/baby-platom/loyalty-system/internal/server"
)

func main() {
	config.ParseFlags()
	if err := logger.Initialize(config.Config.LogLevel); err != nil {
		panic(err)
	}
	database.Prepare()
	accrual.Prepare()

	server.Run()
}
