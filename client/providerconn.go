package client

// XXX: This is a terrible duplicate of cacheconn.go.

import (
	"context"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// - Assigns sequence numbers to outbound messages.
// - Routes replies by matching sequence numbers.
// - How do we handle the consumer of a reply exiting/terminating/canceling?
type providerConnection struct {
	l *logrus.Logger

	nextSequenceNo uint64

	conn       *grpc.ClientConn
	grpcClient ccmsg.ClientProviderClient
}

func newProviderConnection(ctx context.Context, l *logrus.Logger, addr string) (*providerConnection, error) {
	// XXX: No transport security!
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial")
	}
	grpcClient := ccmsg.NewClientProviderClient(conn)

	return &providerConnection{
		l: l,

		nextSequenceNo: 3000, // XXX: Make this easier to pick out of logs.

		conn:       conn,
		grpcClient: grpcClient,
	}, nil
}

func (cc *providerConnection) Close(ctx context.Context) error {
	cc.l.Info("providerConnection.Close() - enter")

	if err := cc.conn.Close(); err != nil {
		return errors.Wrap(err, "failed to close connection")
	}
	return nil

}
