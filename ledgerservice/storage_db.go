package ledgerservice

import (
	"context"
	"database/sql"

	"github.com/cachecashproject/go-cachecash/dbtx"
	"github.com/cachecashproject/go-cachecash/ledger"
	"github.com/cachecashproject/go-cachecash/ledgerservice/models"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"github.com/volatiletech/sqlboiler/types"
)

type LedgerDatabase struct{}

var _ LedgerStorage = (*LedgerDatabase)(nil)

func NewLedgerDatabase() *LedgerDatabase {
	return &LedgerDatabase{}
}

func (m *LedgerDatabase) MempoolTXs(ctx context.Context) ([]*models.MempoolTransaction, error) {
	return models.MempoolTransactions().All(ctx, dbtx.ExecutorFromContext(ctx))
}

func (m *LedgerDatabase) Utxo(ctx context.Context, outpoint ledger.OutpointKey) (*models.Utxo, error) {
	txid := outpoint.TXID()
	outputIdx := outpoint.Idx()
	return models.Utxos(qm.Where("txid=? and output_idx=?", types.BytesArray{0: txid[:]}, outputIdx)).One(ctx, dbtx.ExecutorFromContext(ctx))
}

func (m *LedgerDatabase) HighestBlock(ctx context.Context) (*models.Block, error) {
	block, err := models.Blocks(qm.OrderBy("height DESC, block_id DESC")).One(ctx, dbtx.ExecutorFromContext(ctx))
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return block, nil
}

func (m *LedgerDatabase) InsertBlock(ctx context.Context, blockModel *models.Block) error {
	return blockModel.Insert(ctx, dbtx.ExecutorFromContext(ctx), boil.Infer())
}

func (m *LedgerDatabase) DeleteMempoolTX(ctx context.Context, txid ledger.TXID) error {
	dbTxID := types.BytesArray{0: txid[:]}
	_, err := models.MempoolTransactions(qm.Where("txid=?", dbTxID)).DeleteAll(ctx, dbtx.ExecutorFromContext(ctx))
	return err
}

func (m *LedgerDatabase) UpdateAuditLog(ctx context.Context, txid ledger.TXID, status string) error {
	dbTxID := types.BytesArray{0: txid[:]}
	_, err := models.TransactionAuditlogs(qm.Where("txid=?", dbTxID)).UpdateAll(ctx, dbtx.ExecutorFromContext(ctx), models.M{"status": status})
	return err
}

func (m *LedgerDatabase) DeleteUtxo(ctx context.Context, outpoint ledger.OutpointKey) error {
	txid := outpoint.TXID()
	outputIdx := outpoint.Idx()
	_, err := models.Utxos(qm.Where("txid=? and output_idx=?", types.BytesArray{0: txid[:]}, outputIdx)).DeleteAll(ctx, dbtx.ExecutorFromContext(ctx))
	return err
}

func (m *LedgerDatabase) InsertUtxo(ctx context.Context, utxo *models.Utxo) error {
	return utxo.Insert(ctx, dbtx.ExecutorFromContext(ctx), boil.Infer())
}

func (m *LedgerDatabase) QueueTX(ctx context.Context, tx *models.MempoolTransaction) error {
	return tx.Insert(ctx, dbtx.ExecutorFromContext(ctx), boil.Infer())
}
