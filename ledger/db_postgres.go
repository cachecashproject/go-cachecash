package ledger

import (
	"context"

	"github.com/gobuffalo/packr"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"

	ledger_models "github.com/cachecashproject/go-cachecash/ledger/models"
	"github.com/cachecashproject/go-cachecash/ledger/postgres/models"
)

// ChainStoragePostgres implements a block storage backend for postgres
type ChainStoragePostgres struct{}

var _ ChainSQL = (*ChainStoragePostgres)(nil)

func NewChainStoragePostgres() *ChainStoragePostgres {
	return &ChainStoragePostgres{}
}

func (cdb *ChainStoragePostgres) DialectName() string {
	return "postgres"
}

func (cdb *ChainStoragePostgres) MigrationSource() *migrate.PackrMigrationSource {
	return &migrate.PackrMigrationSource{Box: packr.NewBox("postgres/migrations")}
}

func (cdb *ChainStoragePostgres) OneBlock(ctx context.Context, executor boil.ContextExecutor, query qm.QueryMod) (BlockSQL, error) {
	block, err := models.RawBlocks(query).One(ctx, executor)
	return &BlockPostgresql{block}, err
}

func (cbd *ChainStoragePostgres) InsertBlock(ctx context.Context, executor boil.ContextExecutor, blkid []byte, height uint64, bytes []byte) error {
	blockModel := &models.RawBlock{
		Blockid: blkid,
		Height:  int64(height),
		Bytes:   bytes,
	}
	return blockModel.Insert(ctx, executor, boil.Infer())
}

// OneTX retrieves a single tx from the underlying store
func (cdb *ChainStoragePostgres) OneTX(ctx context.Context, executor boil.ContextExecutor, query qm.QueryMod) (TXSQL, error) {
	tx, err := models.RawTxes(query).One(ctx, executor)
	return &TxPostgresql{tx}, err
}

// InsertTx inserts a block into the underlying store
func (cdb *ChainStoragePostgres) InsertTx(ctx context.Context, executor boil.ContextExecutor, txid ledger_models.TXID, bytes []byte) error {
	txModel := &models.RawTX{
		Txid:  txid,
		Bytes: bytes,
	}
	return txModel.Insert(ctx, executor, boil.Infer())
}

type BlockPostgresql struct {
	*models.RawBlock
}

var _ BlockSQL = (*BlockPostgresql)(nil)

func (block *BlockPostgresql) BlockID() []byte {
	return block.RawBlock.Blockid
}

func (block *BlockPostgresql) Height() uint64 {
	return uint64(block.RawBlock.Height)
}

func (block *BlockPostgresql) Bytes() []byte {
	return block.RawBlock.Bytes
}

type TxPostgresql struct {
	*models.RawTX
}

var _ TXSQL = (*TxPostgresql)(nil)

func (tx *TxPostgresql) TxID() ledger_models.TXID {
	return tx.RawTX.Txid
}

func (tx *TxPostgresql) Bytes() []byte {
	return tx.RawTX.Bytes
}
