package ledgerservice

import (
	"context"
	"database/sql"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/keypair"
	"github.com/cachecashproject/go-cachecash/ledger"
	"github.com/cachecashproject/go-cachecash/ledgerservice/models"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"github.com/volatiletech/sqlboiler/types"
)

type LedgerService struct {
	// The LedgerService knows each cache's "inner master key" (aka "master key")?  This is an AES key.
	// For each cache, it also knows an IP address, a port number, and a public key.

	l  *logrus.Logger
	db *sql.DB
	kp *keypair.KeyPair

	newTxChan *chan struct{}
}

func NewLedgerService(l *logrus.Logger, db *sql.DB, kp *keypair.KeyPair, newTxChan *chan struct{}) (*LedgerService, error) {
	s := &LedgerService{
		l:  l,
		db: db,
		kp: kp,

		newTxChan: newTxChan,
	}

	return s, nil
}

func (s *LedgerService) PostTransaction(ctx context.Context, req *ccmsg.PostTransactionRequest) (*ccmsg.PostTransactionResponse, error) {
	s.l.WithFields(logrus.Fields{"tx": req.Tx}).Info("PostTransaction")

	dbTx, err := s.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin database transaction")
	}
	// defer dbTx.Close() ?

	txBytes, err := req.Tx.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal tx into bytes")
	}

	txid, err := req.Tx.TXID()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get txid")
	}
	txidBytes := types.BytesArray{0: txid[:]}

	mpTx := models.MempoolTransaction{
		Txid: txidBytes,
		Raw:  txBytes,
	}
	if err := mpTx.Insert(ctx, dbTx, boil.Infer()); err != nil {
		return nil, errors.Wrap(err, "failed to insert mempool transaction")
	}

	taTx := models.TransactionAuditlog{
		Txid:   txidBytes,
		Raw:    txBytes,
		Status: models.TransactionStatusPending,
	}
	if err := taTx.Insert(ctx, dbTx, boil.Infer()); err != nil {
		return nil, errors.Wrap(err, "failed to insert transaction into audit log")
	}

	if err := dbTx.Commit(); err != nil {
		return nil, errors.Wrap(err, "failed to commit database transaction")
	}

	if s.newTxChan != nil {
		s.l.Debug("Notifying mining go routine")
		*s.newTxChan <- struct{}{}
	}

	s.l.Info("PostTransaction - success")

	return &ccmsg.PostTransactionResponse{}, nil
}

func (s *LedgerService) GetBlocks(ctx context.Context, req *ccmsg.GetBlocksRequest) (*ccmsg.GetBlocksResponse, error) {
	s.l.Info("GetBlocks")

	if req.Limit > 100 {
		return nil, errors.New("limit is too high")
	}

	dbBlocks, err := models.Blocks(
		qm.Where("height >= ?", req.StartDepth),
		qm.OrderBy("height ASC"),
		qm.Limit(int(req.Limit)),
	).All(ctx, s.db)

	if err != nil {
		return nil, errors.New("failed to get blocks")
	}

	blocks := make([]*ledger.Block, 0, len(dbBlocks))
	for _, dbBlock := range dbBlocks {
		block := &ledger.Block{}
		err = block.Unmarshal(dbBlock.Raw)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse stored block")
		}
		blocks = append(blocks, block)
	}

	s.l.WithFields(logrus.Fields{
		"blocks":     len(blocks),
		"startDepth": req.StartDepth,
		"limit":      req.Limit,
	}).Debug("sending block reply")

	return &ccmsg.GetBlocksResponse{
		Blocks: blocks,
	}, nil
}
