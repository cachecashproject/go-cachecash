package ledgerservice

import (
	"context"
	"errors"

	"github.com/cachecashproject/go-cachecash/ledger"
	ledger_models "github.com/cachecashproject/go-cachecash/ledger/models"
	"github.com/cachecashproject/go-cachecash/ledgerservice/models"
)

type LedgerMemory struct {
	mempool map[ledger_models.TXID]*models.MempoolTransaction
	utxos   map[ledger.OutpointKey]*models.Utxo
}

var _ LedgerStorage = (*LedgerMemory)(nil)

func NewLedgerMemory() *LedgerMemory {
	return &LedgerMemory{
		mempool: map[ledger_models.TXID]*models.MempoolTransaction{},
		utxos:   map[ledger.OutpointKey]*models.Utxo{},
	}
}

func (m *LedgerMemory) MempoolTXs(ctx context.Context) ([]*models.MempoolTransaction, error) {
	mempool := make([]*models.MempoolTransaction, 0, len(m.mempool))
	for _, tx := range m.mempool {
		mempool = append(mempool, tx)
	}
	return mempool, nil
}

func (m *LedgerMemory) Utxo(ctx context.Context, outpoint ledger.OutpointKey) (*models.Utxo, error) {
	utxo, ok := m.utxos[outpoint]
	if ok {
		return utxo, nil
	} else {
		return nil, errors.New("utxo not found")
	}
}

func (m *LedgerMemory) HighestBlock(ctx context.Context) (*models.Block, error) {
	// panic("todo")
	return nil, nil
}

func (m *LedgerMemory) InsertBlock(ctx context.Context, blockModel *models.Block) error {
	// panic("todo")
	return nil
}

func (m *LedgerMemory) DeleteMempoolTX(ctx context.Context, txid ledger_models.TXID) error {
	delete(m.mempool, txid)
	return nil
}

func (m *LedgerMemory) UpdateAuditLog(ctx context.Context, txid ledger_models.TXID, status string) error {
	return nil
}

func (m *LedgerMemory) DeleteUtxo(ctx context.Context, outpoint ledger.OutpointKey) error {
	delete(m.utxos, outpoint)
	return nil
}

func (m *LedgerMemory) InsertUtxo(ctx context.Context, utxo *models.Utxo) error {
	outpoint, err := ledger.NewOutpointKey(utxo.Txid[0], byte(utxo.OutputIdx))
	if err != nil {
		return err
	}
	m.utxos[*outpoint] = utxo
	return nil
}

func (m *LedgerMemory) QueueTX(ctx context.Context, tx *models.MempoolTransaction) error {
	txid := ledger_models.TXID{}
	copy(txid[:], tx.Txid[0])
	m.mempool[txid] = tx
	return nil
}
