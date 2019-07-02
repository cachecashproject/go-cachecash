package client

import (
	"context"
	"time"

	cachecash "github.com/cachecashproject/go-cachecash"
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/trace"
)

type fetchGroup struct {
	bundle *ccmsg.TicketBundle
	err    error
	notify []chan DownloadResult
}

// Outcome is used to signal what the outcome of the client handling of a bundle was.
type Outcome int

const (
	// Completed indicates the bundle has been handled by the client
	Completed Outcome = iota
	// Deferred indicates the bundle cannot be handled yet (e.g. the client has requested a retry
	// and is waiting for the retried bundle to arrive)
	Deferred
	// Retry indicates that the bundle was corrupt in some fashion and should be retried (this covers
	// all manner of faults - bad cache connections, bad data from a cache, bad
	// tickets etc).
	Retry
)

// BundleOutcome is used to track successful decryption and handling of bundles.
// It is ok to fail to send a failed outcome if the context is cancelled; it is
// not ok to fail to sent an Ok outcome - read ahead and completion tracking
// is based around successful outcome notifications.
type BundleOutcome struct {
	Outcome     Outcome
	ChunkOffset uint64
	Chunks      uint64
	// Bundle carries a deferred bundle which the client is not ready to process yet
	Bundle *fetchGroup
}

// schedule is responsible for requesting bundles from publishers and chunk data from caches.
//
//
//
// it tries to balance:
// - reading ahead to mitigate RTT, publisher, and cache latencies
// - don't saturate the network with data the client isn't ready for yet
// - don't bloat the process with data the client isn't ready for yet
//
// retries:
// - when a bundle fails hard for any reason it *will be* [not implemented] retried from the publisher as a narrow
//   byte-range specified bundle request
// - the code tracks
//   successfully processed high water mark completedChunks based on notification from
//   the client must only notify in-order currently.
//   ... coming soon
func (cl *client) schedule(ctx context.Context, path string, queue chan<- *fetchGroup, bundleOutcomes <-chan BundleOutcome) {
	defer close(queue)

	var chunkRangeBegin uint64
	var byteRangeBegin uint64
	// NB: for a zero-length file this being zero by default is trivially fine.
	//     Scheduling is finished when completed is strictly = to chunk count, not chunk *index*.
	//     That is, completing chunk 0 sets this to 1, 1 to 2 and so on.
	var completedChunks uint64

	minimumBacklogDepth := uint64(0)
	bundleRequestInterval := 0
	schedulerNotify := make(chan bool, 64)

	for {
		var bundles []*ccmsg.TicketBundle
		var err error

		finishedOutcomes := false
		for !finishedOutcomes {
			select {
			case bundleOutcome := <-bundleOutcomes:
				switch bundleOutcome.Outcome {
				case Completed:
					completedChunks = bundleOutcome.ChunkOffset + bundleOutcome.Chunks
					cl.l.Debugf("completed %d chunks at chunk %d", bundleOutcome.Chunks, bundleOutcome.ChunkOffset)
					// Here is where a readahead check 'is it time to read ahead' would sit and replace the time.After heuristic
				case Deferred:
					// Deferral isn't handled further down the pipeline yet.
					if bundleOutcome.Bundle == nil {
						err = errors.New("nil Bundle in Deferred bundleOutcome")
						queue <- &fetchGroup{err: err}
						cl.l.Error("encountered an error, shutting down scheduler")
						return
					}
					queue <- bundleOutcome.Bundle
				case Retry:
					err = errors.Errorf("client failed bundle %d %d", bundleOutcome.ChunkOffset, bundleOutcome.Chunks)
					queue <- &fetchGroup{err: err}
					cl.l.Error("encountered an error, shutting down scheduler")
					return
				}
			default:
				cl.l.Debug("No bundle outcomes to process")
				finishedOutcomes = true
			}
		}

		// First iteration or have not requested the last chunk
		if cl.chunkCount == nil || chunkRangeBegin < *cl.chunkCount {
			// Have bundles to request
			bundles, err = cl.requestBundles(ctx, path, chunkRangeBegin, 0)
			if err != nil {
				queue <- &fetchGroup{
					err: err,
				}
				cl.l.Error("encountered an error, shutting down scheduler")
				return
			}
		} else {
			if completedChunks >= *cl.chunkCount {
				cl.l.Info("got all bundles, terminating scheduler")
				return
			}

			bundles = []*ccmsg.TicketBundle{}
		}

		for _, bundle := range bundles {
			if cl.chunkCount == nil {
				// Cache the chunk count to permit completion detection when retrying non-terminal blocks
				_count := bundle.Metadata.ChunkCount()
				cl.chunkCount = &_count
				_size := bundle.Metadata.GetChunkSize()
				cl.chunkSize = &_size
			} else {
				if bundle.Metadata.ChunkCount() != *cl.chunkCount {
					err = errors.New("object chunk count changed mid retrieval")
					queue <- &fetchGroup{err: err}
					cl.l.Error("encountered an error, shutting down scheduler")
					return
				}
				if bundle.Metadata.GetChunkSize() != *cl.chunkSize {
					err = errors.New("object chunk size changed mid retrieval")
					queue <- &fetchGroup{err: err}
					cl.l.Error("encountered an error, shutting down scheduler")
					return
				}

			}
			chunks := len(bundle.TicketRequest)
			cl.l.WithFields(logrus.Fields{
				"len(chunks)": chunks,
			}).Info("pushing bundle to downloader")

			// For each chunk in TicketBundle, dispatch a request to the appropriate cache.
			chunkResults := make([]*chunkRequest, chunks)

			fetchGroup := &fetchGroup{
				bundle: bundle,
				notify: []chan DownloadResult{},
			}

			for i := 0; i < chunks; i++ {
				b := &chunkRequest{
					bundle: bundle,
					idx:    i,
				}
				chunkResults[i] = b

				ci := bundle.CacheInfo[i]
				pubKey := ci.Pubkey.GetPublicKey()
				cc, err := cl.GetCacheConnection(ctx, ci.Addr.ConnectionString(), pubKey)
				if err != nil {
					cl.l.WithError(err).Error("failed to connect to cache")
					// In future we should resubmit the bundle - but this is better than panicing.
					fetchGroup.err = err
					fetchGroup.bundle = nil
					queue <- fetchGroup
					return
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
			chunkRangeBegin += uint64(chunks)
			byteRangeBegin += uint64(chunks) * bundle.Metadata.ChunkSize

			minimumBacklogDepth = uint64(bundle.Metadata.MinimumBacklogDepth)
			bundleRequestInterval = int(bundle.Metadata.BundleRequestInterval)
		}
		cl.waitUntilNextRequest(schedulerNotify, minimumBacklogDepth, bundleRequestInterval)
	}
}

func (cl *client) waitUntilNextRequest(schedulerNotify chan bool, minimumBacklogDepth uint64, bundleRequestInterval int) {
	for {
		interval := time.Duration(bundleRequestInterval) * time.Second
		intervalRemaining := interval - time.Since(cl.lastBundleRequest)

		select {
		case <-schedulerNotify:
			cl.l.WithFields(logrus.Fields{
				"minimumBacklogDepth": minimumBacklogDepth,
			}).Debug("checking cache backlog depth")
			if cl.checkBacklogDepth(minimumBacklogDepth) {
				cl.l.Info("cache backlog is running low, requesting new bundle")
				return
			}
		case <-time.After(intervalRemaining):
			cl.l.WithFields(logrus.Fields{
				"interval": bundleRequestInterval,
			}).Info("interval reached, requesting new bundles")
			return
		}
	}
}

func (cl *client) checkBacklogDepth(n uint64) bool {
	for _, c := range cl.cacheConns {
		if c.BacklogLength() <= n {
			return true
		}
	}
	return false
}

type chunkRequest struct {
	bundle *ccmsg.TicketBundle
	idx    int

	encData []byte // Singly-encrypted data.
	err     error
}

func (cl *client) requestBundles(ctx context.Context, path string, chunkOffset uint64, chunkCount uint64) ([]*ccmsg.TicketBundle, error) {
	ctx, span := trace.StartSpan(ctx, "cachecash.com/Client/requestBundle")
	defer span.End()

	var byteRangeBegin uint64
	if chunkOffset != 0 {
		byteRangeBegin = *cl.chunkSize * chunkOffset
	}

	var byteRangeEnd uint64 // "continue to the end of the object"
	if chunkCount != 0 {
		byteRangeEnd = byteRangeBegin + *cl.chunkSize*chunkCount
	}

	cl.l.WithFields(logrus.Fields{
		"chunkOffset":    chunkOffset,
		"byteRangeBegin": byteRangeBegin,
	}).Info("requesting bundle")

	cl.l.Info("enumerating backlog length")

	backlogs := make(map[string]uint64)
	for _, cc := range cl.cacheConns {
		cl.l.WithFields(logrus.Fields{
			"cache": cc.PublicKey(),
		}).Info("backlog length: ", cc.BacklogLength())
		backlogs[string(cc.PublicKeyBytes())] = cc.BacklogLength()
	}

	req := &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            path,
		RangeBegin:      byteRangeBegin,
		RangeEnd:        byteRangeEnd,
		BacklogDepth:    backlogs,
	}
	cl.l.Infof("sending content request to publisher: %v", req)

	// Send request to publisher; get TicketBundle in response.
	resp, err := cl.publisherConn.GetContent(ctx, req)
	if err != nil {
		err = errors.Wrapf(err, "failed to fetch chunk-group at chunk offset %d", chunkOffset)
		return nil, err
	}
	bundles := resp.Bundles
	for _, bundle := range bundles {
		cl.l.Info("got ticket bundle from publisher for escrow: ", bundle.GetRemainder().GetEscrowId())
		// cl.l.Debugf("got ticket bundle from publisher: %v", proto.MarshalTextString(bundle))
	}

	cl.lastBundleRequest = time.Now()

	return bundles, nil
}
