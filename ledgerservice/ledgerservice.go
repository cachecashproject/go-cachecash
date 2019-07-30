package ledgerservice

import (
	"context"
	"database/sql"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/ledger"
	"github.com/cachecashproject/go-cachecash/ledgerservice/models"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"golang.org/x/crypto/ed25519"
)

type LedgerService struct {
	// The LedgerService knows each cache's "inner master key" (aka "master key")?  This is an AES key.
	// For each cache, it also knows an IP address, a port number, and a public key.

	l  *logrus.Logger
	db *sql.DB
}

func NewLedgerService(l *logrus.Logger, db *sql.DB) (*LedgerService, error) {
	p := &LedgerService{
		l:  l,
		db: db,
	}

	return p, nil
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
				return errors.Wrap(err, "failed to get utxo")
			}
			_ = utxo
			// inputSum += utxo.Amount
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
	blockBytes, err := block.Marshal()
	if err != nil {
		return errors.Wrap(err, "Failed to marshal block")
	}

	blockModel := models.Block{
		/*
			Height   int
			BlockID  []byte
			ParentID null.Bytes
			Raw      []byte
		*/
		Raw: blockBytes,
	}
	err = blockModel.Insert(ctx, s.db, boil.Infer())
	if err != nil {
		return errors.Wrap(err, "Failed to insert new block to database")
	}

	// mark utxos as spent
	_, err = models.Utxos(qm.WhereIn("txid in ?", state.SpentUtxos())).DeleteAll(ctx, s.db)
	if err != nil {
		return errors.Wrap(err, "Failed to remove spent UTXOs")
	}

	// delete transactions from mempool
	_, err = models.MempoolTransactions(qm.WhereIn("txid in ?", txs)).DeleteAll(ctx, s.db)
	if err != nil {
		return errors.Wrap(err, "Failed to remove executed TXs")
	}

	// update auditlog
	_, err = models.TransactionAuditlogs(qm.WhereIn("txid in ?", txs)).UpdateAll(ctx, s.db, models.M{"status": models.TransactionStatusMined})
	if err != nil {
		return errors.Wrap(err, "Failed to update transaction status in audit log")
	}

	// add new UTXOs
	for _, tx := range txs {
		txid, err := tx.TXID()
		if err != nil {
			return errors.Wrap(err, "failed to get txid from transaction")
		}

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

	return nil
}

func (s *LedgerService) AddOutputsToDatabase(ctx context.Context, txid ledger.TXID, outputs []ledger.TransactionOutput) error {
	for idx, output := range outputs {
		utxo := models.Utxo{
			Txid:         txid[:],
			OutputIdx:    idx,
			Value:        int(output.Value), // TODO: review and/or fix type
			ScriptPubkey: output.ScriptPubKey,
			/*
				Txid         []byte `boil:"txid" json:"txid" toml:"txid" yaml:"txid"`
				OutputIdx    int    `boil:"output_idx" json:"output_idx" toml:"output_idx" yaml:"output_idx"`
				Value        int    `boil:"value" json:"value" toml:"value" yaml:"value"`
				ScriptPubkey []byte `boil:"script_pubkey" json:"script_pubkey" toml:"script_pubkey" yaml:"script_pubkey"`
			*/
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

// XXX: transactions that depend on outputs of transactions we haven't mined are rejected
// TODO: if height is 0, accept genesis blocks with arbitrary outpus, else reject all genesis blocks
func (s *LedgerService) validateInputOutputSums(ctx context.Context, tx ledger.Transaction) error {
	inputSum := 0
	outputSum := 0

	switch v := tx.Body.(type) {
	case *ledger.TransferTransaction:
		// TODO: make sure our addition is overflow safe

		for _, input := range v.Inputs {
			utxo, err := s.GetUtxo(ctx, input.Key())
			if err != nil {
				return errors.New("could not find utxo")
			}
			inputSum += utxo.Value
		}

		for _, output := range v.Outputs {
			outputSum += int(output.Value)
		}

		// XXX: there are no fees yet
		if outputSum != inputSum {
			return errors.New("output sum and input sum doesn't match")
		}
	case *ledger.GenesisTransaction:
		return errors.New("only the first block can be a genesis block")
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
