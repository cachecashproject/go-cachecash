package main

import (
	"context"
	"database/sql"

	"github.com/cachecashproject/go-cachecash/cache/models"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/testutil"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/boil"
)

func main() {
	l := logrus.New()
	ctx := context.Background()

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
}
