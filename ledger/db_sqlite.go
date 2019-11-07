package ledger

import (
	"context"

	"github.com/gobuffalo/packr"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"

	ledger_models "github.com/cachecashproject/go-cachecash/ledger/models"
	"github.com/cachecashproject/go-cachecash/ledger/sqlite/models"
)

// ChainStorageSqlite implements a block storage backend for sqlite
type ChainStorageSqlite struct{}

var _ ChainSQL = (*ChainStorageSqlite)(nil)

func NewChainStorageSqlite() *ChainStorageSqlite {
	return &ChainStorageSqlite{}
}

func (cdb *ChainStorageSqlite) DialectName() string {
	return "sqlite3"
}

func (cdb *ChainStorageSqlite) MigrationSource() *migrate.PackrMigrationSource {
	return &migrate.PackrMigrationSource{Box: packr.NewBox("sqlite/migrations")}
}

func (cdb *ChainStorageSqlite) OneBlock(ctx context.Context, executor boil.ContextExecutor, query qm.QueryMod) (BlockSQL, error) {
	block, err := models.RawBlocks(query).One(ctx, executor)
	return &BlockSqlite{block}, err
}

func (cbd *ChainStorageSqlite) InsertBlock(ctx context.Context, executor boil.ContextExecutor, blkid []byte, height uint64, bytes []byte) error {
	blockModel := &models.RawBlock{
		Blockid: blkid,
		Height:  int64(height),
		Bytes:   bytes,
	}
	return blockModel.Insert(ctx, executor, boil.Infer())
}

// OneTX retrieves a single tx from the underlying store
func (cdb *ChainStorageSqlite) OneTX(ctx context.Context, executor boil.ContextExecutor, query qm.QueryMod) (TXSQL, error) {
	tx, err := models.RawTxes(query).One(ctx, executor)
	return &TxSqlite{tx}, err
}

// InsertTx inserts a block into the underlying store
func (cdb *ChainStorageSqlite) InsertTx(ctx context.Context, executor boil.ContextExecutor, txid ledger_models.TXID, bytes []byte) error {
	txModel := &models.RawTX{
		Txid:  txid,
		Bytes: bytes,
	}
	return txModel.Insert(ctx, executor, boil.Infer())
}

type BlockSqlite struct {
	*models.RawBlock
}

var _ BlockSQL = (*BlockSqlite)(nil)

func (block *BlockSqlite) BlockID() []byte {
	return block.RawBlock.Blockid
}

func (block *BlockSqlite) Height() uint64 {
	return uint64(block.RawBlock.Height)
}

func (block *BlockSqlite) Bytes() []byte {
	return block.RawBlock.Bytes
}

type TxSqlite struct {
	*models.RawTX
}

var _ TXSQL = (*TxSqlite)(nil)

func (tx *TxSqlite) TxID() ledger_models.TXID {
	return tx.RawTX.Txid
}

func (tx *TxSqlite) Bytes() []byte {
	return tx.RawTX.Bytes
}
