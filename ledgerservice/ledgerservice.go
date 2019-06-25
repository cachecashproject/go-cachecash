package ledgerservice

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
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
	s.l.Info("PostTransaction")
	return nil, errors.New("no implementation")
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
