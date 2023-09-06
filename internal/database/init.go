package database

import (
	"github.com/baby-platom/loyalty-system/internal/config"
	"github.com/baby-platom/loyalty-system/internal/logger"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"moul.io/zapgorm2"
)

var DB Database

func Prepare() {
	session, err := gorm.Open(
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

	err = session.AutoMigrate(User{}, Order{}, Withdraw{}, Balance{})
	if err != nil {
		logger.Log.Error("cannot perform migration", zap.Error(err))
	}

	DB.session = session
}
