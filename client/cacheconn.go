package client

import (
	"context"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// - Assigns sequence numbers to outbound messages.
// - Routes replies by matching sequence numbers.
// - How do we handle the consumer of a reply exiting/terminating/canceling?
type cacheConnection struct {
	l    *logrus.Logger
	addr string

	nextSequenceNo uint64

	conn       *grpc.ClientConn
	grpcClient ccmsg.ClientCacheClient
	backlog    chan DownloadTask
}

type DownloadTask struct {
	req    *blockRequest
	notify chan DownloadResult
}

type DownloadResult struct {
	resp  *blockRequest
	cache *cacheConnection
}

func newCacheConnection(ctx context.Context, l *logrus.Logger, addr string) (*cacheConnection, error) {
	// XXX: No transport security!
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial")
	}

	grpcClient := ccmsg.NewClientCacheClient(conn)

	return &cacheConnection{
		l:    l,
		addr: addr,

		nextSequenceNo: 4000, // XXX: Make this easier to pick out of logs.

		conn:       conn,
		grpcClient: grpcClient,
		backlog:    make(chan DownloadTask, 128),
	}, nil
}

func (cc *cacheConnection) Close(ctx context.Context) error {
	cc.l.WithField("cache", cc.addr).Info("cacheConnection.Close() - enter")
	close(cc.backlog)
	if err := cc.conn.Close(); err != nil {
		return errors.Wrap(err, "failed to close connection")
	}
	return nil
}

func (cc *cacheConnection) Run(ctx context.Context) {
	l := cc.l.WithFields(logrus.Fields{
		"cache": cc.addr,
	})
	for task := range cc.backlog {
		l.Debug("got download request")
		blockRequest := task.req
		err := cc.requestBlock(ctx, blockRequest)
		blockRequest.err = err
		l.Debug("yielding download result")
		task.notify <- DownloadResult{
			resp:  blockRequest,
			cache: cc,
		}
	}
	l.Info("downloader successfully terminated")
}

func (cc *cacheConnection) QueueRequest(task DownloadTask) {
	cc.backlog <- task
}

func (cc *cacheConnection) ExchangeTicketL2(ctx context.Context, req *ccmsg.ClientCacheRequest) error {
	_, err := cc.grpcClient.ExchangeTicketL2(ctx, req)
	return err
}

func (cc *cacheConnection) requestBlock(ctx context.Context, b *blockRequest) error {
	// Send request ticket to cache; await data.
	reqData, err := b.bundle.BuildClientCacheRequest(b.bundle.TicketRequest[b.idx])
	if err != nil {
		return errors.Wrap(err, "failed to build client-cache request")
	}
	tt := common.StartTelemetryTimer(cc.l, "getBlock")
	msgData, err := cc.grpcClient.GetBlock(ctx, reqData)
	if err != nil {
		return errors.Wrap(err, "failed to exchange request ticket with cache")
	}
	tt.Stop()
	cc.l.WithFields(logrus.Fields{
		"cache":    cc.addr,
		"blockIdx": b.bundle.TicketRequest[b.idx].BlockIdx,
		"len":      len(msgData.Data),
	}).Infof("got data response from cache")

	// Send L1 ticket to cache; await outer decryption key.
	reqL1, err := b.bundle.BuildClientCacheRequest(b.bundle.TicketL1[b.idx])
	if err != nil {
		return errors.Wrap(err, "failed to build client-cache request")
	}
	tt = common.StartTelemetryTimer(cc.l, "exchangeTicketL1")
	msgL1, err := cc.grpcClient.ExchangeTicketL1(ctx, reqL1)
	if err != nil {
		return errors.Wrap(err, "failed to exchange request ticket with cache")
	}
	tt.Stop()
	cc.l.WithField("cache", cc.addr).Infof("got L1 response from cache")

	// Decrypt data.
	encData, err := util.EncryptDataBlock(
		b.bundle.TicketRequest[b.idx].BlockIdx,
		b.bundle.Remainder.RequestSequenceNo,
		msgL1.OuterKey.Key,
		msgData.Data)
	if err != nil {
		return errors.Wrap(err, "failed to decrypt data")
	}
	b.encData = encData

	// Done!
	return nil
}
