package ledgerclient

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/ledger"
)

// DEFAULT_SYNC_INTERVAL is the number of seconds between RPC calls to retrieve
// new blocks
const DEFAULT_SYNC_INTERVAL = 10

// Replicator replicates the ledger from the central server to a local data store, calling back to subscribers on
// every block.
type Replicator struct {
	l          *logrus.Logger
	storage    *ledger.Database
	conn       *grpc.ClientConn
	GrpcClient ccmsg.LedgerClient
}

// NewReplicator creates a new replicator replicating into `persistence` from `addr`.
func NewReplicator(l *logrus.Logger, storage *ledger.Database, addr string, insecure bool) (*Replicator, error) {
	l.Info("dialing ledger service: ", addr)
	conn, err := common.GRPCDial(addr, insecure)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial ledger service")
	}

	grpcClient := ccmsg.NewLedgerClient(conn)

	return &Replicator{
		storage:    storage,
		conn:       conn,
		GrpcClient: grpcClient,
		l:          l,
	}, nil
}

// SyncChain starts a loop that queries new blocks periodically until ctx is cancelled
func (r *Replicator) SyncChain(ctx context.Context, syncInterval time.Duration) {
	for {
		err := r.FetchBlocks(ctx)
		if err != nil {
			r.l.Error(err)
		}
		select {
		case <-time.After(syncInterval):
		case <-ctx.Done():
			return
		}
	}
}

// FetchBlocks queries new blocks from the ledger
func (r *Replicator) FetchBlocks(ctx context.Context) error {
	height, err := r.storage.Height(ctx)
	if err != nil {
		return err
	}

	r.l.WithFields(logrus.Fields{
		"height": height,
	}).Info("Fetching blocks")
	resp, err := r.GrpcClient.GetBlocks(ctx, &ccmsg.GetBlocksRequest{
		StartDepth: int64(height),
		Limit:      5,
	})
	if err != nil {
		return errors.Wrap(err, "failed to fetch blocks")
	}

	if len(resp.Blocks) == 0 {
		r.l.Info("No new blocks")
	}

	for _, block := range resp.Blocks {
		r.l.WithFields(logrus.Fields{
			"height": height,
		}).Info("Appending block")
		if _, err := r.storage.AddBlock(ctx, block); err != nil {
			return err
		}
		height++
	}

	return nil
}
