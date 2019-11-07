package ledger

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/cachecashproject/go-cachecash/ledger/models"
)

// ChainStorageMemory implmenets an in-memory block storage backend
type ChainStorageMemory struct {
	blocks map[BlockID]*Block
	height uint64
	txns   map[models.TXID]*Transaction
}

var _ Persistence = (*ChainStorageMemory)(nil)

func NewChainStorageMemory(genesisBlock *Block) *ChainStorageMemory {
	gbID := genesisBlock.BlockID()
	return &ChainStorageMemory{
		blocks: map[BlockID]*Block{gbID: genesisBlock},
		height: 1,
		txns:   map[models.TXID]*Transaction{},
	}
}

func (cdb *ChainStorageMemory) Height(ctx context.Context) (uint64, error) {
	return cdb.height, nil
}

func (cdb *ChainStorageMemory) GetBlock(ctx context.Context, blkid BlockID) (*Block, uint64, error) {
	blk, ok := cdb.blocks[blkid]
	if !ok {
		return nil, 0, fmt.Errorf("block not in database: %v", blkid)
	}
	return blk, 0, nil
}

func (cdb *ChainStorageMemory) AddBlock(ctx context.Context, height uint64, blk *Block) error {
	blkID := blk.BlockID()
	if _, ok := cdb.blocks[blkID]; ok {
		return errors.New("block ID already present in database")
	}

	cdb.blocks[blkID] = blk
	if height == cdb.height {
		cdb.height = height + 1
	}
	return nil
}

func (cdb *ChainStorageMemory) AddTx(ctx context.Context, txid models.TXID, tx *Transaction) error {
	cdb.txns[txid] = tx
	return nil
}

func (cdb *ChainStorageMemory) GetTx(ctx context.Context, txid models.TXID) (*Transaction, error) {
	tx := cdb.txns[txid]
	return tx, nil
}
