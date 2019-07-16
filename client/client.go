package client

import (
	"context"
	"fmt"
	"time"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/colocationpuzzle"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/trace"
	"golang.org/x/crypto/ed25519"
)

/*
The API of client should mirror that of `http` and `ctxhttp`:
- https://godoc.org/pkg/http
- https://godoc.org/golang.org/x/net/context/ctxhttp

Simplifications made in this implementation of the client:
- A client interacts with a single content publisher.
- A client requests only a single object at a time.
*/

// OutputChunk contains either some ready to use bytes or an error.
type OutputChunk struct {
	Data []byte
	Err  error
}

// Client provides a simple interface to retrieve files from the network.
type Client interface {
	GetObject(ctx context.Context, name string, output chan *OutputChunk)
	Close(ctx context.Context) error
}

// The public key bytes are used for cached connection lookups.
type cacheID string

type chunkGroup struct {
	data     [][]byte
	chunkIdx []uint64
	metadata *ccmsg.ObjectMetadata
}

// What if one chunk is a member of multiple chunk-groups?
type client struct {
	l *logrus.Logger

	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey
	cacheConns map[cacheID]cacheConnection

	// TODO: publishers should be pooled and cached too
	publisherConn publisherConnection

	// These fields are actually single-object specific and should be factored into a dedicated stream object.
	lastBundleRequest time.Time
	chunkCount        *uint64
	chunkSize         *uint64
}

var _ Client = (*client)(nil)

// New creates a new client connecting to the supplied publisher with a
// lazy-connecting grpc client.
func New(l *logrus.Logger, addr string) (Client, error) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate keypair")
	}

	pc, err := newPublisherConnection(l, addr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to publisher")
	}

	return &client{
		l:          l,
		publicKey:  pub,
		privateKey: priv,

		publisherConn: pc,

		cacheConns: make(map[cacheID]cacheConnection),
	}, nil
}

// GetCacheConnection returns a connection from the cache connection pool.
// If none is available, a new connection is initiated and return.
func (cl *client) GetCacheConnection(ctx context.Context, addr string, pubKey ed25519.PublicKey) (cacheConnection, error) {
	cid := (cacheID)(string(pubKey))
	cc, ok := cl.cacheConns[cid]
	if ok {
		return cc, nil
	}
	cc, err := cl.publisherConn.newCacheConnection(cl.l, addr, pubKey)
	if err != nil {
		cl.l.WithError(err).Error("failed to connect to cache")
		return nil, err
	}
	cl.cacheConns[cid] = cc
	go cc.Run(ctx)
	return cc, nil
}

// GetObject streams an object back to the caller on the supplied chan
//
// for chunk := range output {
//     if chunk.Err != nil {
//         ...
//     }
//     // use chunk.Data
// }
func (cl *client) GetObject(ctx context.Context, path string, output chan *OutputChunk) {
	// TODO: User, RawQuery, Fragment are currently ignored.  Should either pass to server or throw an error if they are
	// provided.
	defer close(output)
	ctx, span := trace.StartSpan(ctx, "cachecash.com/Client/GetObject")
	defer span.End()

	// TODO - we need to model and assess the buffer sizes here
	queue := make(chan *fetchGroup, 128)
	completions := make(chan BundleOutcome, 128)
	go cl.schedule(ctx, path, queue, completions)

	// counts chunks
	var rangeBegin uint64
	// counts bytes
	var pos = 0

	// number of times a given chunk has been retried
	// there is no sleep associated with the retry as the
	// fast error case we want to fail fast, and in the slow
	// remote case we want to ask the publisher for new bundles
	// immediately.
	retries := 0

outer:
	for group := range queue {
		outcome := BundleOutcome{ChunkOffset: rangeBegin}
		if group.err != nil {
			output <- &OutputChunk{nil, group.err}
			// Scheduler self-terminates if it is signalling a failure to us.
			return
		}

		if retries > 5 {
			// Too many retries on a single chunk, give it up.
			err := errors.Errorf("Too many retries on chunk %d", rangeBegin)
			output <- &OutputChunk{nil, err}
			// Scheduler self-terminates if it is signalling a failure to us.
			return
		}

		bundle := group.bundle
		chunks := uint64(len(group.notify))
		chunkResults := make([]*chunkRequest, chunks)
		cacheConns := make([]cacheConnection, chunks)
		outcome = BundleOutcome{ChunkOffset: rangeBegin, Chunks: chunks}

		if len(bundle.TicketRequest) > 0 && bundle.TicketRequest[0].ChunkIdx != rangeBegin {
			// Skip over read-ahead chunks with no processing until we get to
			// the bundle we are looking for
			outcome.Outcome = Deferred
			outcome.Bundle = group
			outcome.ChunkOffset = bundle.TicketRequest[0].ChunkIdx
			completions <- outcome
			continue
		}

		// We receive either an error (in which case we abort the other requests and return an error to our parent) or we
		// receive singly-encrypted data from each cache.
		for i, notify := range group.notify {
			select {
			case <-ctx.Done():
				output <- &OutputChunk{nil, ctx.Err()}
				return
			case result := <-notify:
				b := result.resp
				chunkResults[i] = b
				cacheConns[i] = result.cache
				if b.err != nil {
					retries++
					outcome.Outcome = Retry
					completions <- outcome
					continue outer
				}
			}
		}

		cl.l.Info("got singly-encrypted data from each cache")

		bg, err := cl.decryptPuzzle(ctx, bundle, chunkResults, cacheConns)
		if err != nil {
			output <- &OutputChunk{nil, errors.Wrap(err, "failed to decrypt puzzle")}
			// TODO: track which caches are implicated, to allow triangulation of bad cache
			retries++
			outcome.Outcome = Retry
			completions <- outcome
			continue outer
		}

		for i, d := range bg.data {
			if bg.chunkIdx[i] != rangeBegin+uint64(i) {
				output <- &OutputChunk{nil, fmt.Errorf("chunk at position %v has index %v, but expected %v",
					i, bg.chunkIdx[i], rangeBegin+uint64(i))}
				// This is arguably an internal logic error - the scheduler
				// should detect and discard much earlier, but we can come back
				// to tidy things up
				retries++
				outcome.Outcome = Retry
				completions <- outcome
				continue outer
			}
			cl.l.WithFields(logrus.Fields{
				"current position": pos,
				"len(newChunk)":    len(d),
			}).Debug("sending chunk to receiver")
			output <- &OutputChunk{d, nil}
			pos += len(d)
		}

		rangeBegin += chunks
		retries = 0
		outcome.Outcome = Completed
		completions <- outcome
		group.schedulerNotify <- true
	}
}

func (cl *client) Close(ctx context.Context) error {
	// XXX: Open question: how should cache/publisher connection management work?  Probably needs to be a setting, or on
	// a timer, where we have some way of automatically closing connections we aren't using any longer but do allow for
	// reuse.  In the meantime, this function (which may not be appropriately named) manually closes them.  It isn't
	// concurrency-safe, which is probably an issue.

	cl.l.Infof("client.Close() - enter - %v cache conns open", len(cl.cacheConns))

	var retErr error
	for _, cc := range cl.cacheConns {
		// cl.l.Infof("client.Close() - closing cc.co=%p", cc.co)
		if err := cc.Close(ctx); err != nil {
			retErr = err
		}
	}
	cl.cacheConns = make(map[cacheID]cacheConnection)

	if err := cl.publisherConn.Close(ctx); err != nil {
		retErr = err
	}

	cl.l.Info("client.Close() - exit")
	return retErr
}

func (cl *client) decryptPuzzle(ctx context.Context, bundle *ccmsg.TicketBundle, chunkResults []*chunkRequest, cacheConns []cacheConnection) (*chunkGroup, error) {
	// Solve colocation puzzle.
	ctx, span := trace.StartSpan(ctx, "cachecash.com/Client/solvePuzzle")
	defer span.End()
	tt := common.StartTelemetryTimer(cl.l, "solvePuzzle")
	var singleEncryptedChunks [][]byte
	for _, result := range chunkResults {
		singleEncryptedChunks = append(singleEncryptedChunks, result.encData)
	}
	pi := bundle.Remainder.PuzzleInfo
	secret, _, err := colocationpuzzle.Solve(colocationpuzzle.Parameters{
		Rounds:      pi.Rounds,
		StartOffset: uint32(pi.StartOffset),
		StartRange:  uint32(pi.StartRange),
	}, singleEncryptedChunks, pi.Goal)
	tt.Stop()
	if err != nil {
		return nil, errors.Wrap(err, "failed to solve colocation puzzle")
	}

	// Decrypt L2 ticket.
	ticketL2, err := common.DecryptTicketL2(secret, bundle.EncryptedTicketL2)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt L2 ticket")
	}

	// Send the L2 ticket to each cache.
	// XXX: This should not be serialized.
	// l2ResultCh := make(chan l2Result)
	for _, conn := range cacheConns {
		req, err := bundle.BuildClientCacheRequest(&ccmsg.TicketL2Info{
			EncryptedTicketL2: bundle.EncryptedTicketL2,
			PuzzleSecret:      secret,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to build L2 ticket request")
		}
		tt := common.StartTelemetryTimer(cl.l, "exchangeTicketL2")
		err = conn.ExchangeTicketL2(ctx, req)
		tt.Stop()
		if err != nil {
			// TODO: This should not cause us to abort sending the L2 ticket to other caches, and should not prevent us
			// from returning the plaintext data.
			return nil, errors.Wrap(err, "failed to send L2 ticket to cache")
		}
	}

	// Decrypt singly-encrypted chunks to produce final plaintext.
	tt = common.StartTelemetryTimer(cl.l, "decryptData")
	var plaintextChunks [][]byte
	var chunkIdx []uint64
	for i, ciphertext := range singleEncryptedChunks {
		plaintext, err := util.EncryptChunk(
			bundle.TicketRequest[i].ChunkIdx,
			bundle.Remainder.RequestSequenceNo,
			ticketL2.InnerSessionKey[i].Key,
			ciphertext)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decrypt singly-encrypted chunk")
		}
		plaintextChunks = append(plaintextChunks, plaintext)
		chunkIdx = append(chunkIdx, bundle.TicketRequest[i].ChunkIdx)
	}
	tt.Stop()

	// Return data to parent.
	cl.l.Info("chunk-group fetch completed without error")
	return &chunkGroup{
		data:     plaintextChunks,
		chunkIdx: chunkIdx,
		metadata: bundle.Metadata,
	}, nil
}
