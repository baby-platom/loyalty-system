package database

import (
	"context"

	"github.com/baby-platom/loyalty-system/internal/logger"
	"gorm.io/gorm"
)

type Database struct {
	session *gorm.DB
}

func (db *Database) Conn(ctx context.Context) *gorm.DB {
	tx := extractTx(ctx)
	if tx != nil {
		return tx
	}
	return db.session
}

func (db *Database) beginTransaction() (*gorm.DB, error) {
	tx := db.session.Begin()

	if err := tx.Error; err != nil {
		logger.Log.Error(err)
		return nil, err
	}
	return tx, nil
}

func (db *Database) commitTransaction(tx *gorm.DB) error {
	if err := tx.Commit().Error; err != nil {
		logger.Log.Error(err)
		return err
	}
	return nil
}
