package database

import (
	"github.com/baby-platom/loyalty-system/internal/config"
	"github.com/baby-platom/loyalty-system/internal/logger"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"moul.io/zapgorm2"
)

var (
	DB  *gorm.DB
	err error
)

func Prepare() {
	DB, err = gorm.Open(
		postgres.Open(config.Config.DatabaseURI),
		&gorm.Config{
			TranslateError: true,
			Logger:         zapgorm2.New(zap.L()),
			PrepareStmt:    true,
		},
	)
	if err != nil {
		logger.Log.Error("cannot open db", zap.Error(err))
	}

	err = DB.AutoMigrate(User{}, Order{}, Withdraw{}, Balance{})
	if err != nil {
		logger.Log.Error("cannot perform migration", zap.Error(err))
	}
}
