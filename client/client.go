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

// XXX:
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

	publisherConn     publisherConnection
	lastBundleRequest time.Time

	// Unclear what key type should be.
	cacheConns map[cacheID]cacheConnection
}

var _ Client = (*client)(nil)

// New creates a new client connecting to the supplied publisher with a
// lazy-connecting grpc client.
func New(ctx context.Context, l *logrus.Logger, addr string) (Client, error) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate keypair")
	}

	pc, err := newPublisherConnection(ctx, l, addr)
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

	queue := make(chan *fetchGroup, 128)
	go cl.schedule(ctx, path, queue)

	var rangeBegin uint64
	var pos = 0

	for group := range queue {
		if group.err != nil {
			output <- &OutputChunk{nil, group.err}
			return
		}

		bundle := group.bundle
		cacheQty := len(group.notify)
		chunkResults := make([]*chunkRequest, cacheQty)
		cacheConns := make([]cacheConnection, cacheQty)

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
					output <- &OutputChunk{nil, errors.Wrap(b.err, "got error in chunk request; aborting any that remain in group")}
					return
				}
			}
		}

		cl.l.Info("got singly-encrypted data from each cache")
		bg, err := cl.decryptPuzzle(ctx, bundle, chunkResults, cacheConns)
		if err != nil {
			output <- &OutputChunk{nil, errors.Wrap(err, "failed to decrypt puzzle")}
			return
		}

		for i, d := range bg.data {
			if bg.chunkIdx[i] != rangeBegin+uint64(i) {
				output <- &OutputChunk{nil, fmt.Errorf("chunk at position %v has index %v, but expected %v",
					i, bg.chunkIdx[i], rangeBegin+uint64(i))}
				return
			}
			cl.l.WithFields(logrus.Fields{
				"current position": pos,
				"len(newChunk)":    len(d),
			}).Debug("sending chunk to receiver")
			output <- &OutputChunk{d, nil}
			pos += len(d)
		}

		rangeBegin += uint64(len(bg.data))
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
