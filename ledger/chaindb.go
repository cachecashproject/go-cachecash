package ledger

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"

	"github.com/cachecashproject/go-cachecash/dbtx"
	"github.com/cachecashproject/go-cachecash/ledger/models"
	"github.com/cachecashproject/go-cachecash/ledger/txscript"
)

// Position is the location of a transaction: the Nth transaction in a specific block.
type Position struct {
	BlockID BlockID
	TxIndex uint32
}

// Persistence descibes the storage backend that is used to store blocks and transactions
type Persistence interface {
	Height(ctx context.Context) (uint64, error)
	AddBlock(ctx context.Context, height uint64, blk *Block) error
	GetBlock(ctx context.Context, blkid BlockID) (*Block, uint64, error)
	AddTx(ctx context.Context, txid models.TXID, tx *Transaction) error
	GetTx(ctx context.Context, txid models.TXID) (*Transaction, error)
}

// NewBlockSubscriber describes the in-transaction callback made by
// Database for new blocks.
type NewBlockSubscriber interface {
	NewBlock(ctx context.Context, height uint64, block *Block) error
}

// Database describes an interface to query blockchain data.
//
// When considering transaction visiblity, remember that an individual transaction may be included in more than one
// block, and that the block-graph is a tree (since each block has a single parent and cycles are not possible).
//
// TODO: Add functions for walking the block-graph.
type Database struct {
	storage     Persistence
	subscribers []NewBlockSubscriber
}

func NewDatabase(storage Persistence) *Database {
	return &Database{
		storage: storage,
	}
}

func (chain *Database) Height(ctx context.Context) (uint64, error) {
	height, err := chain.storage.Height(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get block height")
	}
	return height, nil
}

func (chain *Database) addTransactions(ctx context.Context, blk *Block) error {
	for _, tx := range blk.Transactions.Transactions {
		txid, err := tx.TXID()
		if err != nil {
			return errors.Wrap(err, "failed to compute TXID")
		}
		if err := chain.storage.AddTx(ctx, txid, tx); err != nil {
			return errors.Wrap(err, "failed to store TX")
		}
	}

	return nil
}

// AddBlock adds a new block to the database.  Its parent must already be in the database; no block with a matching ID
// may already be in the database.
//
// Every subscriber is called and has the opportunity to error if this block would be invalid/inconsistent with their
// logic. Note that a new block may have an equal or lower height than the current highest block when a fork is being
// incorporated (to be a fork there must at least two blocks of the same height).
//
// It is not yet clear whether all data derived from the chain should be calculated transactionally in subscriber
// callbacks, or nontransactionally in some sort of cache layer built above this; but for clarity, this transactional
// callback feature exists to provide loose coupling within the core components, not to tie together caching or other
// layers.
//
// As a special case, if there are no subscribers, no transaction is created. This exists to support the in-memory DB
// which doesn't support transactional semantics, and for which we thus cannot support the callback mechanism.
func (chain *Database) AddBlock(ctx context.Context, blk *Block) (height uint64, err error) {
	var tx *sql.Tx
	if len(chain.subscribers) != 0 {
		ctx, tx, err = dbtx.BeginTx(ctx)
	}
	if err != nil {
		return 0, errors.Wrap(err, "failed to begin transaction")
	}
	defer func() {
		if err != nil && tx != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = errors.Wrapf(err, "failed to rollback transaction with error (%q)", rollbackErr)
			}
		}
	}()

	height = uint64(0)
	parentBlockID := blk.Header.PreviousBlock
	if parentBlockID.Zero() {
		height, err := chain.storage.Height(ctx)
		if err != nil {
			return 0, err
		}
		if height > 0 {
			return 0, errors.New("genesis blocks must be height #0")
		}
	} else {
		_, parentHeight, err := chain.storage.GetBlock(ctx, blk.Header.PreviousBlock)
		if err != nil {
			return 0, errors.Wrap(err, "failed to find parent block")
		}
		height = parentHeight + 1
	}

	if err := chain.storage.AddBlock(ctx, height, blk); err != nil {
		return 0, errors.Wrap(err, "failed to store block")
	}
	if err := chain.addTransactions(ctx, blk); err != nil {
		return 0, errors.Wrap(err, "failed to store transactions")
	}

	for _, sub := range chain.subscribers {
		if err := sub.NewBlock(ctx, height, blk); err != nil {
			return 0, errors.Wrap(err, "failed to update subscriber")
		}
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
			return 0, errors.Wrap(err, "failed to commit transaction")
		}
	}

	return height, nil
}

type txVisitor func(tx *Transaction) (bool, error)

func (chain *Database) visitTransactions(ctx context.Context, cc *Position, vis txVisitor) error {
	blkid := cc.BlockID

	for !blkid.Zero() {
		blk, _, err := chain.storage.GetBlock(ctx, blkid)
		if err != nil {
			return err
		}

		startIdx := len(blk.Transactions.Transactions) - 1
		if blkid.Equal(cc.BlockID) {
			startIdx = int(cc.TxIndex) - 1
			if startIdx > len(blk.Transactions.Transactions) {
				return errors.New("out-of-range transaction index")
			}
		}

		for i := startIdx; i >= 0; i-- {
			ok, err := vis(blk.Transactions.Transactions[i])
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

// GetTransaction returns a transaction by ID.  If cc is non-nil, the transaction is returned only if it is visible
// from that position (that is, if it occurs earlier in the same block or anywhere in an ancestor block).  If cc is
// nil, the transaction is returned no matter where it is in the block-graph.
//
// If the transaction does not exist or is not visible from cc, returns (nil, nil).
func (chain *Database) GetTransaction(ctx context.Context, cc *Position, txid models.TXID) (*Transaction, error) {
	if cc == nil {
		return nil, errors.New("cc==nil case not implemented")
	}

	var resultTx *Transaction
	if err := chain.visitTransactions(ctx, cc, func(vtx *Transaction) (bool, error) {
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

// Subscribe adds a new NewBlockSubscriber. The subscriber will be called back on new blocks added to the chain, within
// the transaction that is adding the block.
func (chain *Database) Subscribe(subscriber NewBlockSubscriber) {
	chain.subscribers = append(chain.subscribers, subscriber)
}

// Unspent returns true iff no transaction spending the output identified by op is visible from the position in the
// block-graph described by cc.
func (chain *Database) Unspent(ctx context.Context, cc *Position, op Outpoint) (bool, error) {
	if cc == nil {
		return false, errors.New("cc must not be nil")
	}

	var found, unspent bool
	if err := chain.visitTransactions(ctx, cc, func(vtx *Transaction) (bool, error) {
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
				// we found the transaction but the input has been spent
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

// TransactionValid determines whether a given transaction is valid at a particular place in the block-graph.
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
func (cdb *Database) TransactionValid(ctx context.Context, cc *Position, tx *Transaction) error {
	switch tx.Body.TxType() {
	case TxTypeGenesis:
	case TxTypeTransfer:
		break
	default:
		// TODO: We need to think carefully about validation as we add new transaction types.  This is intended to
		// prevent us from unintentionally forgetting to do so.
		return errors.New("unexpected transaction type")
	}

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
		ok, err := cdb.Unspent(ctx, cc, ip)
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
		prevTx, err := cdb.GetTransaction(ctx, cc, ip.PreviousTx)
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
