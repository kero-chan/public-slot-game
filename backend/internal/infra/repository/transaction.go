package repository

import (
	"context"

	"gorm.io/gorm"
)

// txKey is the context key for database transactions
type txKey struct{}

// TxManager handles database transactions
type TxManager struct {
	db *gorm.DB
}

// NewTxManager creates a new transaction manager
func NewTxManager(db *gorm.DB) *TxManager {
	return &TxManager{db: db}
}

// WithTransaction executes fn within a database transaction
// If fn returns an error, the transaction is rolled back
// If fn succeeds, the transaction is committed
func (m *TxManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Inject transaction into context
		txCtx := context.WithValue(ctx, txKey{}, tx)
		return fn(txCtx)
	})
}

// DB returns the underlying database connection
func (m *TxManager) DB() *gorm.DB {
	return m.db
}

// GetTxFromContext retrieves the transaction from context
// Returns the transaction if present, otherwise returns nil
func GetTxFromContext(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(txKey{}).(*gorm.DB)
	if !ok {
		return nil
	}
	return tx
}

// GetDBOrTx returns the transaction from context if present, otherwise returns the provided db
func GetDBOrTx(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx := GetTxFromContext(ctx); tx != nil {
		return tx.WithContext(ctx)
	}
	return db.WithContext(ctx)
}
