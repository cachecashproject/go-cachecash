package ledgerservice

import (
	"context"
	"database/sql"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/ledgerservice/models"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/boil"
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

	mpTx := models.MempoolTransaction{Raw: txBytes}
	if err := mpTx.Insert(ctx, dbTx, boil.Infer()); err != nil {
		return nil, errors.Wrap(err, "failed to insert mempool transaction")
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
