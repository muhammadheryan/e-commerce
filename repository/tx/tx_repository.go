package tx

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type TxRepository interface {
	BeginTx(ctx context.Context) (*sqlx.Tx, error)
	CommitTx(tx *sqlx.Tx) error
	RollbackTx(tx *sqlx.Tx) error
}

type txRepo struct {
	db *sqlx.DB
}

func NewTxRepository(db *sqlx.DB) TxRepository {
	return &txRepo{db: db}
}

func (r *txRepo) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}

func (r *txRepo) CommitTx(tx *sqlx.Tx) error {
	return tx.Commit()
}

func (r *txRepo) RollbackTx(tx *sqlx.Tx) error {
	return tx.Rollback()
}
