package database

import (
	"github.com/baby-platom/loyalty-system/internal/logger"
	"gorm.io/gorm"
)

func GetTransaction() (*gorm.DB, error) {
	tx := DB.Begin()

	if err := tx.Error; err != nil {
		logger.Log.Error(err)
		return nil, err
	}
	return tx, nil
}

func CommitTransaction(tx *gorm.DB) error {
	if err := tx.Commit().Error; err != nil {
		logger.Log.Error(err)
		return err
	}
	return nil
}
