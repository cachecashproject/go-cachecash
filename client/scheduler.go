package client

import (
	"context"

	cachecash "github.com/cachecashproject/go-cachecash"
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/pkg/errors"
)

type fetchGroup struct {
	bundle *ccmsg.TicketBundle
	notify []chan DownloadResult
}

func (cl *client) schedule(ctx context.Context, path string) {
	var blockSize uint64
	var rangeBegin uint64

	for {
		bundle, err := cl.requestBundle(ctx, path, rangeBegin*blockSize)
		if err != nil {
			cl.l.Error("failed to fetch block-group at offset ", rangeBegin, ": ", err)
			break
		}

		chunks := uint64(len(bundle.TicketRequest))
		cl.l.Infof("pushing bundle with %d chunks to downloader", chunks)

		///

		cacheQty := len(bundle.CacheInfo)
		// For each block in TicketBundle, dispatch a request to the appropriate cache.
		chunkResults := make([]*blockRequest, cacheQty)
		cacheConns := make([]*Downloader, cacheQty)

		fetchGroup := &fetchGroup{
			bundle: bundle,
			notify: []chan DownloadResult{},
		}

		for i := 0; i < cacheQty; i++ {
			b := &blockRequest{
				bundle: bundle,
				idx:    i,
			}
			chunkResults[i] = b

			cid := (cacheID)(bundle.CacheInfo[i].Addr.ConnectionString())
			d, ok := cl.cacheConns[cid]
			if !ok {
				// XXX: It's problematic to pass ctx here, because canceling the context will destroy the cache connections!
				// (It should only cancel this particular block-group request.)
				cc, err := newCacheConnection(ctx, cl.l, bundle.CacheInfo[i].Addr.ConnectionString())
				if err != nil {
					cc.l.WithError(err).Error("failed to connect to cache")
					b.err = err
				}
				d = NewDownloader(cl.l, cc)
				cl.cacheConns[cid] = d
				go d.Run(ctx)
			}
			cacheConns = append(cacheConns, d)

			notify := make(chan DownloadResult)
			fetchGroup.notify = append(fetchGroup.notify, notify)
			d.QueueRequest(DownloadTask{
				req:    b,
				notify: notify,
			})
		}

		cl.queue <- fetchGroup
		rangeBegin += chunks

		if rangeBegin >= bundle.Metadata.BlockCount() {
			cl.l.Info("got all bundles")
			break
		}
		blockSize = bundle.Metadata.BlockSize
	}
	cl.l.Info("scheduler successfully terminated")
	close(cl.queue)
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
