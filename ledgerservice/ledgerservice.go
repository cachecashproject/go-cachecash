package ledgerservice

import (
	"context"
	"database/sql"
	"math"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/keypair"
	"github.com/cachecashproject/go-cachecash/ledger"
	"github.com/cachecashproject/go-cachecash/ledgerservice/models"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/null"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"github.com/volatiletech/sqlboiler/types"
	"golang.org/x/crypto/ed25519"
)

type LedgerService struct {
	// The LedgerService knows each cache's "inner master key" (aka "master key")?  This is an AES key.
	// For each cache, it also knows an IP address, a port number, and a public key.

	l  *logrus.Logger
	db *sql.DB
	kp *keypair.KeyPair

	CurrentBlock *models.Block
}

func NewLedgerService(l *logrus.Logger, db *sql.DB, kp *keypair.KeyPair) (*LedgerService, error) {
	p := &LedgerService{
		l:  l,
		db: db,
		kp: kp,
	}

	return p, nil
}

func (s *LedgerService) InitGenesisBlock(totalCoins uint32) error {
	// the genesis block mints totalCoins to s.kp.PublicKey
	block := &ledger.Block{
		Header: &ledger.BlockHeader{},
		Transactions: []*ledger.Transaction{
			{
				Version: 1,
				Flags:   0,
				Body: &ledger.GenesisTransaction{
					Outputs: []ledger.TransactionOutput{
						{
							Value:        totalCoins,
							ScriptPubKey: s.kp.PublicKey,
						},
					},
				},
			},
		},
	}

	return s.ApplyBlock(context.TODO(), block, []ledger.OutpointKey{})
}

func (s *LedgerService) BuildBlock(ctx context.Context) error {
	state := ledger.NewSpendingState()

	mempoolTXs, err := models.MempoolTransactions().All(ctx, s.db)
	if err != nil {
		return err
	}

	for _, txRow := range mempoolTXs {
		tx := ledger.Transaction{}
		err := tx.Unmarshal(txRow.Raw)
		if err != nil {
			return errors.New("found invalid transaction in mempool")
		}

		// TODO: verify transaction is correctly signed

		// verify all inputs are unspent
		for _, inpoint := range tx.Inpoints() {
			utxo, err := s.GetUtxo(ctx, inpoint.Key())
			if err != nil {
				return errors.Wrap(err, "failed to get utxo, input probably already spent")
			}
			// input sum and output sum is already verified at this point
			_ = utxo
		}

		// try to add tx into block
		err = state.AddTx(&tx)
		if err != nil {
			s.l.Info("failed to add tx to block, possibly conflicts with other tx")
		}

		// XXX: if the block is almost full we're going to loop through all remaining transactions regardless
	}

	// TODO: start postgres transaction
	txs := state.AcceptedTransactions()

	_, sigKey, err := ed25519.GenerateKey(nil)
	previousBlock := ledger.BlockID{}

	_ = err

	// write new block
	block, err := ledger.NewBlock(sigKey, previousBlock, txs)
	if err != nil {
		return errors.Wrap(err, "Failed to create block")
	}

	return s.ApplyBlock(ctx, block, state.SpentUTXOs())
}

func (s *LedgerService) ApplyBlock(ctx context.Context, block *ledger.Block, spentUtxos []ledger.OutpointKey) error {
	// TODO: use a transaction, if something in here fails, revert
	blockBytes, err := block.Marshal()
	if err != nil {
		return errors.Wrap(err, "Failed to marshal block")
	}

	blockModel := &models.Block{
		BlockID: block.CanonicalDigest(),
		Raw:     blockBytes,
	}

	if s.CurrentBlock == nil {
		blockModel.Height = 0
		blockModel.ParentID = null.BytesFrom([]byte{}) // TODO: this should be 0000..000 instead
	} else {
		blockModel.Height = s.CurrentBlock.Height + 1
		blockModel.ParentID = null.BytesFrom(s.CurrentBlock.BlockID)
	}

	err = blockModel.Insert(ctx, s.db, boil.Infer())
	if err != nil {
		return errors.Wrap(err, "Failed to insert new block to database")
	}

	// mark utxos as spent
	for _, outpoint := range spentUtxos {
		txid := types.BytesArray{0: outpoint[:32]}
		outputIdx := outpoint[32]

		_, err = models.Utxos(qm.Where("txid=? and output_idx=?", txid, outputIdx)).DeleteAll(ctx, s.db)
		if err != nil {
			return errors.Wrap(err, "Failed to remove spent UTXOs")
		}
	}

	for _, tx := range block.Transactions {
		txid, err := tx.TXID()
		if err != nil {
			return errors.Wrap(err, "failed to get txid from transaction")
		}
		dbTxID := types.BytesArray{0: txid[:]}

		// delete transactions from mempool
		_, err = models.MempoolTransactions(qm.Where("txid=?", dbTxID)).DeleteAll(ctx, s.db)
		if err != nil {
			return errors.Wrap(err, "Failed to remove executed TXs")
		}

		// update auditlog
		_, err = models.TransactionAuditlogs(qm.Where("txid=?", dbTxID)).UpdateAll(ctx, s.db, models.M{"status": models.TransactionStatusMined})
		if err != nil {
			return errors.Wrap(err, "Failed to update transaction status in audit log")
		}

		// add new UTXOs
		switch v := tx.Body.(type) {
		case *ledger.TransferTransaction:
			err = s.AddOutputsToDatabase(ctx, txid, v.Outputs)
			if err != nil {
				return err
			}
		case *ledger.GenesisTransaction:
			err = s.AddOutputsToDatabase(ctx, txid, v.Outputs)
			if err != nil {
				return err
			}
		}
	}

	// transaction completed, uptdating our struct
	s.CurrentBlock = blockModel

	return nil
}

func (s *LedgerService) AddOutputsToDatabase(ctx context.Context, txid ledger.TXID, outputs []ledger.TransactionOutput) error {
	for idx, output := range outputs {
		dbTxID := types.BytesArray{0: txid[:]}

		utxo := models.Utxo{
			Txid:         dbTxID,
			OutputIdx:    idx,
			Value:        int(output.Value), // TODO: review and/or fix type
			ScriptPubkey: output.ScriptPubKey,
		}
		err := utxo.Insert(ctx, s.db, boil.Infer())
		if err != nil {
			return errors.Wrap(err, "Failed to insert new utxo to database")
		}
	}

	return nil
}

func (s *LedgerService) GetUtxo(ctx context.Context, outpoint ledger.OutpointKey) (*models.Utxo, error) {
	txid := outpoint[:32]
	output_idx := outpoint[32]

	utxo, err := models.Utxos(qm.Where("txid=? and output_idx=?", txid, output_idx)).One(ctx, s.db)
	if err != nil {
		return nil, err
	}

	return utxo, nil
}

func safeAddU32(base, next uint32) (uint32, error) {
	if base > math.MaxUint32-next {
		return 0, errors.New("transactions would overflow")
	}
	return base + next, nil
}

// XXX: transactions that depend on outputs of transactions we haven't mined are rejected
// TODO: if height is 0, accept genesis blocks with arbitrary outpus, else reject all genesis blocks
func (s *LedgerService) validateInputOutputSums(ctx context.Context, tx ledger.Transaction) error {
	inputSum := uint32(0)
	outputSum := uint32(0)

	switch v := tx.Body.(type) {
	case *ledger.TransferTransaction:
		for _, input := range v.Inputs {
			utxo, err := s.GetUtxo(ctx, input.Key())
			if err != nil {
				return errors.New("could not find utxo")
			}
			inputSum += uint32(utxo.Value)
		}

		for _, output := range v.Outputs {
			var err error
			outputSum, err = safeAddU32(outputSum, output.Value)
			if err != nil {
				return err
			}
		}

		// XXX: there are no fees yet
		if outputSum != inputSum {
			return errors.New("output sum and input sum doesn't match")
		}
	case *ledger.GenesisTransaction:
		if s.CurrentBlock != nil {
			return errors.New("only the first block can be a genesis block")
		}
	}

	return nil
}

func (s *LedgerService) PostTransaction(ctx context.Context, req *ccmsg.PostTransactionRequest) (*ccmsg.PostTransactionResponse, error) {
	s.l.WithFields(logrus.Fields{"tx": req.Tx}).Info("PostTransaction")

	dbTx, err := s.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin database transaction")
	}
	// defer dbTx.Close() ?

	// validate there are enough inputs for the output amount
	err = s.validateInputOutputSums(ctx, req.Tx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to validate input and output sums")
	}

	txBytes, err := req.Tx.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal tx into bytes")
	}

	mpTx := models.MempoolTransaction{Raw: txBytes}
	if err := mpTx.Insert(ctx, dbTx, boil.Infer()); err != nil {
		return nil, errors.Wrap(err, "failed to insert mempool transaction")
	}

	taTx := models.TransactionAuditlog{
		Raw:    txBytes,
		Status: models.TransactionStatusPending,
	}
	if err := taTx.Insert(ctx, dbTx, boil.Infer()); err != nil {
		return nil, errors.Wrap(err, "failed to insert transaction into audit log")
	}

	if err := dbTx.Commit(); err != nil {
		return nil, errors.Wrap(err, "failed to commit database transaction")
	}

	s.l.Info("PostTransaction - success")

	return &ccmsg.PostTransactionResponse{}, nil
}

func (s *LedgerService) GetBlocks(ctx context.Context, req *ccmsg.GetBlocksRequest) (*ccmsg.GetBlocksResponse, error) {
	s.l.Info("GetBlocks")
	return nil, errors.New("no implementation")
}

/*
	db, err := sql.Open("sqlite3", "./cache.db")
	if err != nil {
		l.Fatal(err)
	}

	tx, err := db.Begin()
	if err != nil {
		l.Fatal(err)
	}

	lcms, err := models.LogicalCacheMappings().All(ctx, tx)
	if err != nil {
		l.Fatal(err)
	}
	for i, lcm := range lcms {
		l.Infof("%v: %v", i, lcm)
	}

	txid, err := common.BytesToEscrowID(testutil.RandBytes(common.EscrowIDSize))
	if err != nil {
		panic(err)
	}
	ne := models.LogicalCacheMapping{
		Txid:    txid,
		SlotIdx: uint64(len(lcms)),
	}
	if err := ne.Insert(ctx, tx, boil.Infer()); err != nil {
		l.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		l.Fatal(err)
	}

	l.Info("fin")
	_ = db

*/
