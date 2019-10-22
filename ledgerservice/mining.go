package ledgerservice

import (
	"context"
	"encoding/hex"
	"math"
	"time"

	"github.com/cachecashproject/go-cachecash/keypair"
	"github.com/cachecashproject/go-cachecash/ledger"
	"github.com/cachecashproject/go-cachecash/ledger/txscript"
	"github.com/cachecashproject/go-cachecash/ledgerservice/models"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/null"
	"github.com/volatiletech/sqlboiler/types"
	"golang.org/x/crypto/ed25519"
)

type LedgerMiner struct {
	l       *logrus.Logger
	storage LedgerStorage
	kp      *keypair.KeyPair

	NewTxChan    chan struct{}
	Interval     time.Duration
	CurrentBlock *models.Block
}

func NewLedgerMiner(l *logrus.Logger, storage LedgerStorage, kp *keypair.KeyPair) (*LedgerMiner, error) {
	newTxChan := make(chan struct{}, 8)

	m := &LedgerMiner{
		l:       l,
		storage: storage,
		kp:      kp,

		NewTxChan: newTxChan,
	}

	return m, nil
}

func (m *LedgerMiner) QueueTX(ctx context.Context, tx *ledger.Transaction) error {
	txBytes, err := tx.Marshal()
	if err != nil {
		return errors.Wrap(err, "failed to marshal tx into bytes")
	}

	txid, err := tx.TXID()
	if err != nil {
		return errors.Wrap(err, "failed to get txid")
	}
	txidBytes := types.BytesArray{0: txid[:]}

	return m.storage.QueueTX(ctx, &models.MempoolTransaction{
		Txid: txidBytes,
		Raw:  txBytes,
	})
}

func (m *LedgerMiner) SetupCurrentBlock(ctx context.Context, totalCoins uint32) error {
	highBlock, err := m.storage.HighestBlock(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to query highest block")
	}
	m.CurrentBlock = highBlock
	if m.CurrentBlock == nil {
		m.l.Info("creating genesis block")
		_, err := m.InitGenesisBlock(ctx, 420000000)
		if err != nil {
			return errors.Wrap(err, "failed to create genesis block")
		}
	} else {
		m.l.WithField("blockid", m.CurrentBlock.BlockID).Infof("starting miner with existing block")
	}
	return nil
}

func (m *LedgerMiner) InitGenesisBlock(ctx context.Context, totalCoins uint32) (*ledger.Block, error) {
	pubKeyHash := txscript.Hash160Sum(m.kp.PublicKey)
	script, err := txscript.MakeP2WPKHInputScript(pubKeyHash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to crate p2wpkh input script")
	}

	scriptBytes, err := script.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal input script")
	}

	tx := ledger.Transaction{
		Version: 1,
		Flags:   0,
		Body: &ledger.GenesisTransaction{
			Outputs: []ledger.TransactionOutput{
				{
					Value:        totalCoins,
					ScriptPubKey: scriptBytes,
				},
			},
		},
	}

	txid, err := tx.TXID()
	if err != nil {
		return nil, err
	}
	m.l.Info("adding genesis transaction: ", txid)

	// the genesis block mints totalCoins to m.kp.PublicKey
	block := &ledger.Block{
		Header: &ledger.BlockHeader{},
		Transactions: &ledger.Transactions{Transactions: []*ledger.Transaction{
			&tx,
		},
		},
	}

	return m.ApplyBlock(ctx, block, []ledger.OutpointKey{})
}

func (m *LedgerMiner) GenerateBlock(ctx context.Context) (*ledger.Block, error) {
	state := ledger.NewSpendingState()

	mempoolTXs, err := m.storage.MempoolTXs(ctx)
	if err != nil {
		return nil, err
	}

	modified := true
	for modified {
		mempoolTXs, modified, err = m.FillBlock(ctx, state, mempoolTXs)
		if err != nil {
			return nil, errors.Wrap(err, "failed to fill block")
		}
	}

	// TODO: start postgres transaction
	txs := state.AcceptedTransactions()
	m.l.Info("accepted txs into this block: ", len(txs))
	if len(txs) == 0 {
		return nil, nil
	}

	previousBlock := ledger.BlockID{}
	copy(previousBlock[:], m.CurrentBlock.BlockID)

	// write new block
	block, err := ledger.NewBlock(m.kp.PrivateKey, previousBlock, txs)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create block")
	}

	return m.ApplyBlock(ctx, block, state.SpentUTXOs())
}

func (m *LedgerMiner) FillBlock(ctx context.Context, state *ledger.SpendingState, mempoolTXs []*models.MempoolTransaction) ([]*models.MempoolTransaction, bool, error) {
	if len(mempoolTXs) == 0 {
		// no pending transactions
		m.l.Debug("no pending transactions in mempool")
		return nil, false, nil
	}
	m.l.Debug("found transactions in mempool: ", len(mempoolTXs))

	unaccepted := []*models.MempoolTransaction{}
	modified := false
	for _, txRow := range mempoolTXs {
		tx := ledger.Transaction{}
		err := tx.Unmarshal(txRow.Raw)
		if err != nil {
			return nil, false, errors.New("found invalid transaction in mempool")
		}

		err = m.ValidateTX(ctx, tx, state)
		if err != nil {
			m.l.Info("failed to validate tx, skipping: ", err)
			unaccepted = append(unaccepted, txRow)
			continue
		}

		// try to add tx into block
		err = state.AddTx(&tx)
		if err != nil {
			m.l.Warn("failed to add tx to block, possibly conflicts with other tx: ", err)
			unaccepted = append(unaccepted, txRow)
			continue
		}

		m.l.Info("added tx to block", tx)
		modified = true
		// XXX: if the block is almost full we're going to loop through all remaining transactions regardless
	}

	return unaccepted, modified, nil
}

func (m *LedgerMiner) ValidateTX(ctx context.Context, tx ledger.Transaction, state *ledger.SpendingState) error {
	// validate there are enough inputs for the output amount
	prevOuts, err := m.validateInputOutputSums(ctx, tx, state)
	if err != nil {
		return errors.Wrap(err, "failed to validate input and output sums")
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

		if err := txscript.ExecuteVerify(inScr, outScr, witnesses[i].Data, &tx, i, int64(prevOuts[i].Value)); err != nil {
			return errors.Wrap(err, "failed to execute and verify script pair")
		}
	}

	return nil
}

func (m *LedgerMiner) GetUtxo(ctx context.Context, outpoint ledger.OutpointKey, state *ledger.SpendingState) (*ledger.TransactionOutput, error) {
	// check spending state first
	utxo := state.IsNewUnspent(outpoint)
	if utxo != nil {
		return utxo, nil
	}

	// fall back to database afterwards
	m.l.WithFields(logrus.Fields{
		"txid":      outpoint.TXID(),
		"outputIdx": outpoint.Idx(),
	}).Debugf("fetching utxo from database")
	model, err := m.storage.Utxo(ctx, outpoint)
	if err != nil {
		return nil, err
	}

	return &ledger.TransactionOutput{
		Value:        uint32(model.Value),
		ScriptPubKey: model.ScriptPubkey,
	}, nil
}

func (m *LedgerMiner) ApplyBlock(ctx context.Context, block *ledger.Block, spentUtxos []ledger.OutpointKey) (*ledger.Block, error) {
	var err error

	block.Header.MerkleRoot, err = block.MerkleRoot()
	if err != nil {
		return nil, errors.Wrap(err, "failed to calculate MerkleRoot")
	}

	// TODO: sign actual data
	block.Header.Signature = ed25519.Sign(m.kp.PrivateKey, []byte{})

	// TODO: use a transaction, if something in here fails, revert
	blockBytes, err := block.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to marshal block")
	}

	blockModel := &models.Block{
		BlockID: block.CanonicalDigest(),
		Raw:     blockBytes,
	}

	if m.CurrentBlock == nil {
		blockModel.Height = 0
		blockModel.ParentID = null.BytesFrom([]byte{}) // TODO: this should be 0000..000 instead
	} else {
		blockModel.Height = m.CurrentBlock.Height + 1
		blockModel.ParentID = null.BytesFrom(m.CurrentBlock.BlockID)
	}

	m.l.WithFields(logrus.Fields{
		"blockID": hex.EncodeToString(blockModel.BlockID),
		"height":  blockModel.Height,
	}).Info("adding block to database")
	err = m.storage.InsertBlock(ctx, blockModel)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to insert new block to database")
	}

	for _, tx := range block.Transactions.Transactions {
		txid, err := tx.TXID()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get txid from transaction")
		}

		// delete transactions from mempool
		err = m.storage.DeleteMempoolTX(ctx, txid)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to remove executed TXs")
		}

		// update auditlog
		err = m.storage.UpdateAuditLog(ctx, txid, models.TransactionStatusMined)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to update transaction status in audit log")
		}

		// add new UTXOs
		switch v := tx.Body.(type) {
		case *ledger.TransferTransaction:
			err = m.AddOutputsToDatabase(ctx, txid, v.Outputs)
			if err != nil {
				return nil, err
			}
		case *ledger.GenesisTransaction:
			err = m.AddOutputsToDatabase(ctx, txid, v.Outputs)
			if err != nil {
				return nil, err
			}
		}
	}

	// mark utxos as spent
	m.l.Info("marking utxos as spent")
	for _, outpoint := range spentUtxos {
		err = m.storage.DeleteUtxo(ctx, outpoint)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to remove spent UTXOs")
		}
	}

	// transaction completed, updating our struct
	m.CurrentBlock = blockModel

	return block, nil
}

func (m *LedgerMiner) AddOutputsToDatabase(ctx context.Context, txid ledger.TXID, outputs []ledger.TransactionOutput) error {
	for idx, output := range outputs {
		dbTxID := types.BytesArray{0: txid[:]}

		utxo := &models.Utxo{
			Txid:         dbTxID,
			OutputIdx:    idx,
			Value:        int(output.Value), // TODO: review and/or fix type
			ScriptPubkey: output.ScriptPubKey,
		}
		err := m.storage.InsertUtxo(ctx, utxo)
		if err != nil {
			return errors.Wrap(err, "Failed to insert new utxo to database")
		}
	}

	return nil
}

func safeAddU64(base, next uint64) (uint64, error) {
	if base > math.MaxUint64-next {
		return 0, errors.New("transactions would overflow")
	}
	return base + next, nil
}

// XXX: transactions that depend on outputs of transactions we haven't mined are rejected
// TODO: if height is 0, accept genesis blocks with arbitrary outpus, else reject all genesis blocks
func (m *LedgerMiner) validateInputOutputSums(ctx context.Context, tx ledger.Transaction, state *ledger.SpendingState) ([]*ledger.TransactionOutput, error) {
	var prevOuts []*ledger.TransactionOutput
	inputSum := uint64(0)
	outputSum := uint64(0)

	switch v := tx.Body.(type) {
	case *ledger.TransferTransaction:
		for _, input := range v.Inputs {
			m.l.Info("verifying utxo: ", input.Key())
			utxo, err := m.GetUtxo(ctx, input.Key(), state)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get utxo, input probably already spent")
			}
			inputSum += uint64(utxo.Value)
			prevOuts = append(prevOuts, utxo)
		}

		for _, output := range v.Outputs {
			var err error
			outputSum, err = safeAddU64(outputSum, uint64(output.Value))
			if err != nil {
				return nil, err
			}
		}

		// XXX: there are no fees yet
		if outputSum != inputSum {
			return nil, errors.New("output sum and input sum doesn't match")
		}
	case *ledger.GenesisTransaction:
		if m.CurrentBlock != nil {
			return nil, errors.New("only the first block can be a genesis block")
		}
	}

	return prevOuts, nil
}

func (m *LedgerMiner) Run(ctx context.Context) {
	for {
		m.l.Debug("checking for new transactions in mempool")
		block, err := m.GenerateBlock(ctx)
		if err != nil {
			m.l.Error("failed to create block: ", err)
		} else if block != nil {
			m.l.WithFields(logrus.Fields{
				"blockID": hex.EncodeToString(block.CanonicalDigest()),
				"txs":     len(block.Transactions.Transactions),
			}).Info("block has been created")
		} else {
			m.l.Debug("mempool is empty")
		}

		if !m.BlockScheduler(ctx) {
			m.l.Info("shutting down ledger mining go routine")
			break
		}
	}
}

func (m *LedgerMiner) BlockScheduler(ctx context.Context) bool {
	if m.Interval != 0 {
		return m.BlockSchedulerTimer(ctx)
	} else {
		return m.BlockSchedulerImmediate(ctx)
	}
}

func (m *LedgerMiner) BlockSchedulerTimer(ctx context.Context) bool {
	remaining := m.Interval * time.Second
	resume := time.Now().Add(remaining)

	for remaining > 0 {
		select {
		case <-ctx.Done():
			// exit the go routine
			return false
		case <-time.After(remaining):
		case <-m.NewTxChan:
		}

		now := time.Now()
		remaining = resume.Sub(now)
	}

	// continue the go routine
	return true
}

func (m *LedgerMiner) BlockSchedulerImmediate(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		// exit the go routine
		return false
	case <-time.After(m.Interval * time.Second):
	case <-m.NewTxChan:
	}

	// continue the go routine
	return true
}
