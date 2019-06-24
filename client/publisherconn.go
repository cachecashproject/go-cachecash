package client

// XXX: This is a terrible duplicate of cacheconn.go.

import (
	"context"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ed25519"
	"google.golang.org/grpc"
)

// - Assigns sequence numbers to outbound messages.
// - Routes replies by matching sequence numbers.
// - How do we handle the consumer of a reply exiting/terminating/canceling?
type publisherGrpc struct {
	l *logrus.Logger

	nextSequenceNo uint64

	conn       *grpc.ClientConn
	grpcClient ccmsg.ClientPublisherClient
}

type publisherConnection interface {
	newCacheConnection(context.Context, *logrus.Logger, string, ed25519.PublicKey) (cacheConnection, error)
	GetContent(context.Context, *ccmsg.ContentRequest) (*ccmsg.ContentResponse, error)
	Close(context.Context) error
}

var _ publisherConnection = (*publisherGrpc)(nil)

func newPublisherConnection(ctx context.Context, l *logrus.Logger, addr string) (*publisherGrpc, error) {
	// XXX: No transport security!
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial")
	}
	grpcClient := ccmsg.NewClientPublisherClient(conn)

	return &publisherGrpc{
		l: l,

		nextSequenceNo: 3000, // XXX: Make this easier to pick out of logs.

		conn:       conn,
		grpcClient: grpcClient,
	}, nil
}

func (pc *publisherGrpc) newCacheConnection(ctx context.Context, l *logrus.Logger, addr string, pubkey ed25519.PublicKey) (cacheConnection, error) {
	return newCacheConnection(ctx, l, addr, pubkey)
}

func (pc *publisherGrpc) GetContent(ctx context.Context, req *ccmsg.ContentRequest) (*ccmsg.ContentResponse, error) {
	return pc.grpcClient.GetContent(ctx, req)
}

func (pc *publisherGrpc) Close(ctx context.Context) error {
	pc.l.Info("publisherConnection.Close() - enter")

	if err := pc.conn.Close(); err != nil {
		return errors.Wrap(err, "failed to close connection")
	}
	return nil

}
