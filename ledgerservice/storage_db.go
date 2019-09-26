package ledgerservice

import (
	"context"
	"database/sql"

	"github.com/cachecashproject/go-cachecash/ledger"
	"github.com/cachecashproject/go-cachecash/ledgerservice/models"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"github.com/volatiletech/sqlboiler/types"
)

type LedgerDatabase struct {
	db *sql.DB
}

var _ LedgerStorage = (*LedgerDatabase)(nil)

func NewLedgerDatabase(db *sql.DB) *LedgerDatabase {
	return &LedgerDatabase{
		db: db,
	}
}

func (m *LedgerDatabase) MempoolTXs(ctx context.Context) ([]*models.MempoolTransaction, error) {
	return models.MempoolTransactions().All(ctx, m.db)
}

func (m *LedgerDatabase) Utxo(ctx context.Context, outpoint ledger.OutpointKey) (*models.Utxo, error) {
	txid := outpoint.TXID()
	outputIdx := outpoint.Idx()
	return models.Utxos(qm.Where("txid=? and output_idx=?", types.BytesArray{0: txid}, outputIdx)).One(ctx, m.db)
}

func (m *LedgerDatabase) HighestBlock(ctx context.Context) (*models.Block, error) {
	return models.Blocks(qm.OrderBy("height DESC, block_id DESC")).One(ctx, m.db)
}

func (m *LedgerDatabase) InsertBlock(ctx context.Context, blockModel *models.Block) error {
	return blockModel.Insert(ctx, m.db, boil.Infer())
}

func (m *LedgerDatabase) DeleteMempoolTX(ctx context.Context, txid ledger.TXID) error {
	dbTxID := types.BytesArray{0: txid[:]}
	_, err := models.MempoolTransactions(qm.Where("txid=?", dbTxID)).DeleteAll(ctx, m.db)
	return err
}

func (m *LedgerDatabase) UpdateAuditLog(ctx context.Context, txid ledger.TXID, status string) error {
	dbTxID := types.BytesArray{0: txid[:]}
	_, err := models.TransactionAuditlogs(qm.Where("txid=?", dbTxID)).UpdateAll(ctx, m.db, models.M{"status": status})
	return err
}

func (m *LedgerDatabase) DeleteUtxo(ctx context.Context, outpoint ledger.OutpointKey) error {
	txid := outpoint.TXID()
	outputIdx := outpoint.Idx()
	_, err := models.Utxos(qm.Where("txid=? and output_idx=?", types.BytesArray{0: txid}, outputIdx)).DeleteAll(ctx, m.db)
	return err
}

func (m *LedgerDatabase) InsertUtxo(ctx context.Context, utxo *models.Utxo) error {
	return utxo.Insert(ctx, m.db, boil.Infer())
}

func (m *LedgerDatabase) QueueTX(ctx context.Context, tx *models.MempoolTransaction) error {
	return tx.Insert(ctx, m.db, boil.Infer())
}
