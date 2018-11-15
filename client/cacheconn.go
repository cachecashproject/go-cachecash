package client

import (
	"context"

	"github.com/kelleyk/go-cachecash/ccmsg"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// - Assigns sequence numbers to outbound messages.
// - Routes replies by matching sequence numbers.
// - How do we handle the consumer of a reply exiting/terminating/canceling?
type cacheConnection struct {
	l *logrus.Logger

	nextSequenceNo uint64

	conn       *grpc.ClientConn
	grpcClient ccmsg.ClientCacheClient
}

func newCacheConnection(ctx context.Context, l *logrus.Logger, addr string) (*cacheConnection, error) {
	// XXX: No transport security!
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial")
	}

	grpcClient := ccmsg.NewClientCacheClient(conn)

	return &cacheConnection{
		l: l,

		nextSequenceNo: 4000, // XXX: Make this easier to pick out of logs.

		conn:       conn,
		grpcClient: grpcClient,
	}, nil
}

func (cc *cacheConnection) Close(ctx context.Context) error {
	cc.l.Info("cacheConnection.Close() - enter")
	if err := cc.conn.Close(); err != nil {
		return errors.Wrap(err, "failed to close connection")
	}
	return nil
}
