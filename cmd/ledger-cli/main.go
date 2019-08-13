package main

import (
	"context"
	"crypto/rand"
	"flag"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/ledger"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	ledgerAddr = flag.String("ledgerAddr", "localhost:7778", "Address of ledgerd instance")
)

func main() {
	common.Main(mainC)
}

func mainC() error {
	l := logrus.New()
	ctx := context.Background()

	txBytes := make([]byte, 8)
	_, _ = rand.Read(txBytes)

	txdata := ledger.Transaction{}
	if err := txdata.Unmarshal(txBytes); err != nil {
		return errors.Wrap(err, "failed to unmarshal transaction")
	}

	l.WithFields(logrus.Fields{"tx": txdata}).Info("generated new faux-transaction")

	conn, err := common.GRPCDial(*ledgerAddr)
	if err != nil {
		return errors.Wrap(err, "failed to dial ledger service")
	}

	grpcClient := ccmsg.NewLedgerClient(conn)

	resp, err := grpcClient.PostTransaction(ctx, &ccmsg.PostTransactionRequest{Tx: txdata})
	if err != nil {
		return errors.Wrap(err, "failed to post transaction")
	}

	_ = resp

	l.Info("fin")
	return nil
}
