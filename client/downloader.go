package client

import (
	"context"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type DownloadTask struct {
	req    *blockRequest
	notify chan DownloadResult
}

type DownloadResult struct {
	resp  *blockRequest
	cache *Downloader
}

type Downloader struct {
	l       *logrus.Logger
	cache   *cacheConnection
	backlog chan DownloadTask
}

func NewDownloader(l *logrus.Logger, cache *cacheConnection) *Downloader {
	return &Downloader{
		l:       l,
		cache:   cache,
		backlog: make(chan DownloadTask),
	}
}

func (d *Downloader) Close(ctx context.Context) error {
	close(d.backlog)
	if err := d.cache.Close(ctx); err != nil {
		return err
	}
	return nil
}

func (d *Downloader) Run(ctx context.Context) {
	for task := range d.backlog {
		d.l.Debug("got download request")
		blockRequest := task.req
		d.requestBlock(ctx, d.cache, blockRequest)
		d.l.Debug("yielding download result")
		task.notify <- DownloadResult{
			resp:  blockRequest,
			cache: d,
		}
	}
	d.l.Info("downloader successfully terminated")
}

func (d *Downloader) QueueRequest(task DownloadTask) {
	d.backlog <- task
}

func (d *Downloader) ExchangeTicketL2(ctx context.Context, req *ccmsg.ClientCacheRequest) error {
	_, err := d.cache.grpcClient.ExchangeTicketL2(ctx, req)
	return err
}

func (d *Downloader) requestBlock(ctx context.Context, cc *cacheConnection, b *blockRequest) error {
	// Send request ticket to cache; await data.
	reqData, err := b.bundle.BuildClientCacheRequest(b.bundle.TicketRequest[b.idx])
	if err != nil {
		return errors.Wrap(err, "failed to build client-cache request")
	}
	tt := common.StartTelemetryTimer(d.l, "getBlock")
	msgData, err := cc.grpcClient.GetBlock(ctx, reqData)
	if err != nil {
		return errors.Wrap(err, "failed to exchange request ticket with cache")
	}
	tt.Stop()
	d.l.WithFields(logrus.Fields{
		"blockIdx": b.bundle.TicketRequest[b.idx].BlockIdx,
		"len":      len(msgData.Data),
	}).Infof("got data response from cache")

	// Send L1 ticket to cache; await outer decryption key.
	reqL1, err := b.bundle.BuildClientCacheRequest(b.bundle.TicketL1[b.idx])
	if err != nil {
		return errors.Wrap(err, "failed to build client-cache request")
	}
	tt = common.StartTelemetryTimer(d.l, "exchangeTicketL1")
	msgL1, err := cc.grpcClient.ExchangeTicketL1(ctx, reqL1)
	if err != nil {
		return errors.Wrap(err, "failed to exchange request ticket with cache")
	}
	tt.Stop()
	d.l.Infof("got L1 response from cache")

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
