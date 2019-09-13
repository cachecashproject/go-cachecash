package blockexplorer

import (
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// LedgerClient provides access to the ledger over GRPC
type LedgerClient struct {
	l          *logrus.Logger
	conn       *grpc.ClientConn
	grpcClient ccmsg.LedgerClient
}

// NewLedgerClient creates a new LedgerClient
func NewLedgerClient(l *logrus.Logger, addr string, insecure bool) (*LedgerClient, error) {
	l.Info("dialing ledger service: ", addr)
	conn, err := common.GRPCDial(addr, insecure)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial ledger service")
	}

	grpcClient := ccmsg.NewLedgerClient(conn)

	return &LedgerClient{
		l:          l,
		grpcClient: grpcClient,
		conn:       conn,
	}, nil
}
