package ledger

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"

	"github.com/cachecashproject/go-cachecash/dbtx"
	"github.com/cachecashproject/go-cachecash/ledger/models"
)

const migrationsTable = "migrations-chain-pkg"

// Implementation of ChainPersistence written against the following interface

type ChainSQL interface {
	// DialectName returns the name of the SQL dialect
	DialectName() string
	// MigrationSource returns a migrate.PackrMigrationSource reference with the migrations to apply
	MigrationSource() *migrate.PackrMigrationSource
	// OneBlock retrieves a single block from the underlying store
	OneBlock(ctx context.Context, executor boil.ContextExecutor, query qm.QueryMod) (BlockSQL, error)
	// InsertBlock inserts a block into the underlying store
	InsertBlock(ctx context.Context, executor boil.ContextExecutor, blkid []byte, height uint64, bytes []byte) error
	// OneTX retrieves a single tx from the underlying store
	OneTX(ctx context.Context, executor boil.ContextExecutor, query qm.QueryMod) (TXSQL, error)
	// InsertTx inserts a block into the underlying store
	InsertTx(ctx context.Context, executor boil.ContextExecutor, txid models.TXID, bytes []byte) error
}

type BlockSQL interface {
	// BlockID gets the ID of the block
	BlockID() []byte
	// Height gets the height of the block
	Height() uint64
	// Bytes gets the serialised bytes of the block
	Bytes() []byte
}

type TXSQL interface {
	// TxID gets the transaction ID
	TxID() models.TXID
	// Bytes gets the serialised bytes of the transaction
	Bytes() []byte
}

// ChainStorageSQL implements a block storage backend for SQL
type ChainStorageSQL struct {
	l *logrus.Logger
	ChainSQL
}

var _ Persistence = (*ChainStorageSQL)(nil)

func NewChainStorageSQL(l *logrus.Logger, impl ChainSQL) *ChainStorageSQL {
	return &ChainStorageSQL{
		l:        l,
		ChainSQL: impl,
	}
}

func (store *ChainStorageSQL) RunMigrations(db *sql.DB) error {
	store.l.Info("applying chain migrations")
	migrate.SetTable(migrationsTable)
	n, err := migrate.Exec(db, store.DialectName(), store.MigrationSource(), migrate.Up)
	if err != nil {
		return errors.Wrap(err, "failed to apply chain migrations")
	}
	store.l.Infof("applied %d chain migrations", n)
	return nil
}

func (store *ChainStorageSQL) Height(ctx context.Context) (uint64, error) {
	query := qm.OrderBy("height DESC")
	executor := dbtx.ExecutorFromContext(ctx)
	block, err := store.OneBlock(ctx, executor, query)
	if err == sql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return 0, err
	} else {
		return block.Height() + 1, nil
	}
}

func (store *ChainStorageSQL) AddBlock(ctx context.Context, height uint64, blk *Block) error {
	bytes, err := blk.Marshal()
	if err != nil {
		return errors.Wrap(err, "Failed to marshal block")
	}

	blkid := blk.BlockID()
	executor := dbtx.ExecutorFromContext(ctx)
	return store.InsertBlock(ctx, executor, blkid[:], height, bytes)
}

func (store *ChainStorageSQL) GetBlock(ctx context.Context, blkid BlockID) (*Block, uint64, error) {
	query := qm.Where("blockid = ?", blkid[:])
	executor := dbtx.ExecutorFromContext(ctx)
	dbBlock, err := store.OneBlock(ctx, executor, query)
	if err == sql.ErrNoRows {
		return nil, 0, ErrBlockNotFound
	} else if err != nil {
		return nil, 0, err
	}
	block := &Block{}
	if err := block.Unmarshal(dbBlock.Bytes()); err != nil {
		return nil, 0, err
	}

	return block, dbBlock.Height(), nil
}

func (store *ChainStorageSQL) AddTx(ctx context.Context, txid models.TXID, tx *Transaction) error {
	bytes, err := tx.Marshal()
	if err != nil {
		return errors.Wrap(err, "Failed to marshal tx")
	}

	// Block insertion is done within a transaction, so this read-then-write is safe
	if _, err := store.GetTx(ctx, txid); err == nil {
		// transaction already exists, skipping
		return nil
	}
	executor := dbtx.ExecutorFromContext(ctx)
	return store.InsertTx(ctx, executor, txid, bytes)
}

func (store *ChainStorageSQL) GetTx(ctx context.Context, txid models.TXID) (*Transaction, error) {
	query := qm.Where("txid = ?", txid[:])
	executor := dbtx.ExecutorFromContext(ctx)
	dbTX, err := store.OneTX(ctx, executor, query)
	if err != nil {
		return nil, err
	}

	tx := &Transaction{}
	if err := tx.Unmarshal(dbTX.Bytes()); err != nil {
		return nil, err
	}

	return tx, nil
}
