package ledgerservice

import (
	"context"

	"github.com/cachecashproject/go-cachecash/ledger"
	ledger_models "github.com/cachecashproject/go-cachecash/ledger/models"
	"github.com/cachecashproject/go-cachecash/ledgerservice/models"
)

type LedgerStorage interface {
	MempoolTXs(ctx context.Context) ([]*models.MempoolTransaction, error)
	Utxo(ctx context.Context, outpoint ledger.OutpointKey) (*models.Utxo, error)
	HighestBlock(ctx context.Context) (*models.Block, error)
	InsertBlock(ctx context.Context, blockModel *models.Block) error
	DeleteMempoolTX(ctx context.Context, txid ledger_models.TXID) error
	UpdateAuditLog(ctx context.Context, txid ledger_models.TXID, status string) error
	DeleteUtxo(ctx context.Context, outpoint ledger.OutpointKey) error
	InsertUtxo(ctx context.Context, utxo *models.Utxo) error
	QueueTX(ctx context.Context, tx *models.MempoolTransaction) error
}
