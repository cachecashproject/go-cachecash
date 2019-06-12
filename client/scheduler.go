package client

import (
	"context"
	"time"

	cachecash "github.com/cachecashproject/go-cachecash"
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type fetchGroup struct {
	bundle *ccmsg.TicketBundle
	notify []chan DownloadResult
}

func (cl *client) schedule(ctx context.Context, path string, queue chan *fetchGroup) {
	defer close(queue)

	var blockSize uint64
	var rangeBegin uint64

	minimumBacklogDepth := 0
	bundleRequestInterval := 0
	schedulerNotify := make(chan bool, 64)

	for {
		cl.l.Info("enumerating backlog length")
		for _, cc := range cl.cacheConns {
			cl.l.WithFields(logrus.Fields{
				"cache": cc.pubkey,
			}).Info("backlog length: ", len(cc.backlog))
		}

		cl.l.Info("requesting bundle")
		bundle, err := cl.requestBundle(ctx, path, rangeBegin*blockSize)
		if err != nil {
			cl.l.Error("failed to fetch block-group at offset ", rangeBegin, ": ", err)
			break
		}

		if bundle != nil {
			chunks := len(bundle.TicketRequest)
			cl.l.WithFields(logrus.Fields{
				"len(chunks)": chunks,
			}).Info("pushing bundle to downloader")

			// For each chunk in TicketBundle, dispatch a request to the appropriate cache.
			chunkResults := make([]*blockRequest, chunks)

			fetchGroup := &fetchGroup{
				bundle: bundle,
				notify: []chan DownloadResult{},
			}

			for i := 0; i < chunks; i++ {
				b := &blockRequest{
					bundle: bundle,
					idx:    i,
				}
				chunkResults[i] = b

				cid := (cacheID)(bundle.CacheInfo[i].Addr.ConnectionString())
				cc, ok := cl.cacheConns[cid]
				if !ok {
					var err error
					ci := bundle.CacheInfo[i]

					// XXX: It's problematic to pass ctx here, because canceling the context will destroy the cache connections!
					// (It should only cancel this particular block-group request.)
					cc, err = newCacheConnection(ctx, cl.l, ci.Addr.ConnectionString(), ci.Pubkey.GetPublicKey())
					if err != nil {
						cc.l.WithError(err).Error("failed to connect to cache")
						b.err = err
					}

					cl.cacheConns[cid] = cc
					go cc.Run(ctx)
				}

				clientNotify := make(chan DownloadResult, 128)
				fetchGroup.notify = append(fetchGroup.notify, clientNotify)
				cc.QueueRequest(DownloadTask{
					req:             b,
					clientNotify:    clientNotify,
					schedulerNotify: schedulerNotify,
				})
			}

			queue <- fetchGroup
			rangeBegin += uint64(chunks)

			if rangeBegin >= bundle.Metadata.BlockCount() {
				cl.l.Info("got all bundles")
				break
			}
			blockSize = bundle.Metadata.BlockSize

			minimumBacklogDepth = int(bundle.MinimumBacklogDepth)
			bundleRequestInterval = int(bundle.BundleRequestInterval)
		}

		cl.waitUntilNextRequest(schedulerNotify, minimumBacklogDepth, bundleRequestInterval)
	}
	cl.l.Info("scheduler successfully terminated")
}

func (cl *client) waitUntilNextRequest(schedulerNotify chan bool, minimumBacklogDepth int, bundleRequestInterval int) {
	for {
		select {
		case <-schedulerNotify:
			cl.l.WithFields(logrus.Fields{
				"minimumBacklogDepth": minimumBacklogDepth,
			}).Debug("checking cache backlog depth")
			if cl.checkBacklogDepth(minimumBacklogDepth) {
				cl.l.Info("cache backlog is running low, requesting new bundle")
				return
			}
		case <-time.After(time.Duration(bundleRequestInterval) * time.Second):
			cl.l.WithFields(logrus.Fields{
				"interval": bundleRequestInterval,
			}).Info("interval reached, requesting new bundles")
			return
		}
	}
}

func (cl *client) checkBacklogDepth(n int) bool {
	for _, c := range cl.cacheConns {
		if len(c.backlog) <= n {
			return true
		}
	}
	return false
}

type blockRequest struct {
	bundle *ccmsg.TicketBundle
	idx    int

	encData []byte // Singly-encrypted data.
	err     error
}

func (cl *client) requestBundle(ctx context.Context, path string, rangeBegin uint64) (*ccmsg.TicketBundle, error) {
	req := &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            path,
		RangeBegin:      rangeBegin,
		RangeEnd:        0, // "continue to the end of the object"
	}
	cl.l.Infof("sending content request to publisher: %v", req)

	// Send request to publisher; get TicketBundle in response.
	resp, err := cl.publisherConn.grpcClient.GetContent(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to request bundle from publisher")
	}
	bundle := resp.Bundle
	cl.l.Info("got ticket bundle from publisher for escrow: ", bundle.GetRemainder().GetEscrowId())
	// cl.l.Debugf("got ticket bundle from publisher: %v", proto.MarshalTextString(bundle))

	return bundle, nil
}
