package database

import (
	"context"
	"errors"

	"github.com/baby-platom/loyalty-system/internal/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type key string

const txKey key = "txKey"

// injectTx injects transaction to context
func injectTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

// extractTx extracts transaction from context
func extractTx(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return tx
	}
	return nil
}

func (db *Database) WithinTransaction(ctx context.Context, tFunc func(ctx context.Context) error) (err error) {
	tx, err := DB.beginTransaction()
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			err = errors.New(r.(string))
			tx.Rollback()
		}
	}()

	err = tFunc(injectTx(ctx, tx))
	if err != nil {
		logger.Log.Errorf("rollbacking transaction", zap.Error(err))
		tx.Rollback()
		return err
	}

	if err = DB.commitTransaction(tx); err != nil {
		logger.Log.Errorf("while commiting transaction error occured", zap.Error(err))
		return err
	}
	return nil
}
