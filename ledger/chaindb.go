package ledger

import (
	"fmt"

	"github.com/pkg/errors"
)

type ChainContext struct {
	BlockID BlockID
	TxIndex uint32
}

// ChainDatabase describes an interface to querying blockchain data.
//
// When considering transaction visiblity, remember that an individual transaction may be included in more than one
// block, and that the block-graph is a tree (since each block has a single parent and cycles are not possible).
//
// TODO: Add functions for walking the block-graph.
type ChainDatabase interface {
	// GetTransaction returns a transaction by ID.  If cc is non-nil, the transaction is returned only if it is visible
	// from that position (that is, if it occurs earlier in the same block or anywhere in an ancestor block).  If cc is
	// nil, the transaction is returned no matter where it is in the block-graph.
	//
	// If the transaction does not exist or is not visible from cc, returns (nil, nil).
	GetTransaction(cc *ChainContext, txid TXID) (*Transaction, error)

	// Unspent returns true iff no transaction spending the output identified by op is visible from the position in the
	// block-graph described by cc.
	Unspent(cc *ChainContext, op Outpoint) (bool, error)
}

// simpleChainDatabase is a simple, in-memory chain database.
//
// Its mutative operations are not atomic; if e.g. AddBlock returns an error, the state of the database may be
// inconsistent.
type simpleChainDatabase struct {
	genesisBlock BlockID
	blocks       map[BlockID]*Block
	txns         map[TXID]*Transaction
}

var _ ChainDatabase = (*simpleChainDatabase)(nil)

func NewSimpleChainDatabase(genesisBlock *Block) (*simpleChainDatabase, error) {
	gbID := genesisBlock.BlockID()
	if !genesisBlock.Header.PreviousBlock.Zero() {
		return nil, errors.New("genesis block must have all-zero parent ID")
	}

	cdb := &simpleChainDatabase{
		genesisBlock: gbID,
		blocks:       map[BlockID]*Block{gbID: genesisBlock},
		txns:         map[TXID]*Transaction{},
	}
	if err := cdb.addTransactions(genesisBlock); err != nil {
		return nil, err
	}

	return cdb, nil
}

func (cdb *simpleChainDatabase) addTransactions(blk *Block) error {
	for _, tx := range blk.Transactions {
		txid, err := tx.TXID()
		if err != nil {
			return errors.Wrap(err, "failed to compute TXID")
		}
		cdb.txns[txid] = tx
	}

	return nil
}

// AddBlock adds a new block to the database.  Its parent must already be in the database; no block with a matching ID
// may already be in the database.
func (cdb *simpleChainDatabase) AddBlock(blk *Block) error {
	if _, ok := cdb.blocks[blk.Header.PreviousBlock]; !ok {
		return errors.New("parent block is not in database")
	}
	blkID := blk.BlockID()
	if _, ok := cdb.blocks[blkID]; ok {
		return errors.New("block ID already present in database")
	}

	cdb.blocks[blkID] = blk
	return cdb.addTransactions(blk)
}

type txVisitor func(tx *Transaction) (bool, error)

func (cdb *simpleChainDatabase) visitTransactions(cc *ChainContext, vis txVisitor) error {
	blkid := cc.BlockID

	for !blkid.Zero() {
		blk, ok := cdb.blocks[blkid]
		if !ok {
			return fmt.Errorf("block not in database: %v", blkid)
		}

		startIdx := len(blk.Transactions) - 1
		if blkid.Equal(cc.BlockID) {
			startIdx = int(cc.TxIndex) - 1
			if startIdx > len(blk.Transactions) {
				return errors.New("out-of-range transaction index")
			}
		}

		for i := startIdx; i >= 0; i-- {
			ok, err := vis(blk.Transactions[i])
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}

		blkid = blk.Header.PreviousBlock
	}

	return nil
}

func (cdb *simpleChainDatabase) GetTransaction(cc *ChainContext, txid TXID) (*Transaction, error) {
	if cc == nil {
		return nil, errors.New("cc==nil case not implemented")
	}

	var resultTx *Transaction
	if err := cdb.visitTransactions(cc, func(vtx *Transaction) (bool, error) {
		vtxid, err := vtx.TXID()
		if err != nil {
			return false, errors.Wrap(err, "failed to compute TXID")
		}
		if txid.Equal(vtxid) {
			resultTx = vtx
			return false, nil
		}
		return true, nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to search for transaction")
	}
	return resultTx, nil
}

func (cdb *simpleChainDatabase) Unspent(cc *ChainContext, op Outpoint) (bool, error) {
	if cc == nil {
		return false, errors.New("cc must not be nil")
	}

	var found, unspent bool
	if err := cdb.visitTransactions(cc, func(vtx *Transaction) (bool, error) {
		vtxid, err := vtx.TXID()
		if err != nil {
			return false, errors.Wrap(err, "failed to compute TXID")
		}

		if vtxid.Equal(op.PreviousTx) {
			// We have found the transaction that created this output without finding evidence that the output has been
			// spent.
			found, unspent = true, true
			return false, nil
		}

		for _, ip := range vtx.Inpoints() {
			if op.Equal(ip) {
				found, unspent = true, false
				return false, nil
			}
		}

		return true, nil
	}); err != nil {
		return false, errors.Wrap(err, "failed to search for outpoint")
	}

	if !found {
		return false, errors.New("failed to locate outpoint")
	}
	return unspent, nil
}
