package ledgerservice

import (
	"database/sql"

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
