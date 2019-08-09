package ledger

import (
	"fmt"

	"github.com/cachecashproject/go-cachecash/ledger/txscript"
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

// TransactionValid determines whether a gievn transaction is valid at a particular place in the block-graph.
//
// A transfer transaction is valid iff
// - it is well-formed (has no obvious syntactic/validity issues);
// - its input and output scripts are standard;
// - it spends only previously-unspent inputs;
// - all of its input scripts execute successfully when paired with the corresponding output scripts (which is where
//   input signatures are checked); and
// - the sum of input values is *exactly* equal to the sum of output values (since we do not currently support fees).
//
// Of these requirements, the first two can be checked in isolation.
//
// TODO: We probably want to be able to distinguish between errors that mean "this is not a valid transaction but the
// mechanism worked correctly" and "something went wrong in the chain database (or somewhere else)".
//
// XXX: We clearly need to do some refactoring.  We wind up searching the chain database multiple times, parsing each
// script more than once, etc.
//
func TransactionValid(cdb ChainDatabase, cc *ChainContext, tx *Transaction) error {
	// Check if the transaction is well-formed.
	if err := tx.WellFormed(); err != nil {
		return errors.Wrap(err, "transaction is not well-formed")
	}

	// Check if all input and output scripts are standard.
	if err := tx.Standard(); err != nil {
		return errors.Wrap(err, "script(s) are not standard")
	}

	// Check that all inputs are previously unspent.
	for _, ip := range tx.Inpoints() {
		ok, err := cdb.Unspent(cc, ip)
		if err != nil {
			return errors.Wrap(err, "failed to check inpoint")
		}
		if !ok {
			return errors.New("transaction has previously-spent input")
		}
	}

	// Get matching output for each input.
	var prevOuts []TransactionOutput
	for _, ip := range tx.Inpoints() {
		prevTx, err := cdb.GetTransaction(cc, ip.PreviousTx)
		if err != nil {
			return errors.Wrap(err, "failed to get previous transaction")
		}
		if prevTx == nil {
			return errors.New("failed to locate previous transaction")
		}

		// XXX: Let's make sure we wind up with tests covering situations like this.
		prevTxOuts := prevTx.Outputs()
		if int(ip.Index) >= len(prevTxOuts) {
			return errors.New("input index is out-of-range for previous transaction")
		}

		prevOuts = append(prevOuts, prevTxOuts[int(ip.Index)])
	}

	// Check that all script pairs execute correctly.
	witnesses := tx.Witnesses()
	for i, ti := range tx.Inputs() {
		inScr, err := txscript.ParseScript(ti.ScriptSig)
		if err != nil {
			return errors.Wrap(err, "failed to parse input script")
		}

		outScr, err := txscript.ParseScript(prevOuts[i].ScriptPubKey)
		if err != nil {
			return errors.Wrap(err, "failed to parse output script")
		}

		// TODO: this should be int32 instead of int64?
		if err := txscript.ExecuteVerify(inScr, outScr, witnesses[i].Data, tx, i, int64(prevOuts[i].Value)); err != nil {
			return errors.Wrap(err, "failed to execute and verify script pair")
		}
	}

	// Check that the sum of input and output values matches.
	// N.B.: We use a uint64 to avoid overflow issues with adding multiple uint32s.
	var totalIn, totalOut uint64
	for _, pto := range prevOuts {
		totalIn += uint64(pto.Value)
	}
	for _, to := range tx.Outputs() {
		totalOut += uint64(to.Value)
	}
	if totalIn != totalOut {
		return errors.New("value of inputs does not equal value of outputs")
	}

	return nil
}
