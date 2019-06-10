package client

import (
	"context"
	"fmt"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/colocationpuzzle"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

type Client interface {
	GetObject(ctx context.Context, name string) (Object, error)
	Close(ctx context.Context) error
}

type Object interface {
	Data() []byte
}

// XXX:
type cacheID string

type blockGroup struct {
	data     [][]byte
	blockIdx []uint64
	metadata *ccmsg.ObjectMetadata
}

// What if one block is a member of multiple block-groups?
type client struct {
	l *logrus.Logger

	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey

	publisherConn *publisherConnection

	// Unclear what key type should be.
	cacheConns map[cacheID]*Downloader

	queue chan *fetchGroup
}

var _ Client = (*client)(nil)

func New(l *logrus.Logger, addr string) (Client, error) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate keypair")
	}

	ctx := context.Background() // XXX:
	pc, err := newPublisherConnection(ctx, l, addr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to publisher")
	}

	return &client{
		l:          l,
		publicKey:  pub,
		privateKey: priv,

		publisherConn: pc,

		cacheConns: make(map[cacheID]*Downloader),
		queue:      make(chan *fetchGroup),
	}, nil
}

func (cl *client) GetObject(ctx context.Context, path string) (Object, error) {
	// TODO: User, RawQuery, Fragment are currently ignored.  Should either pass to server or throw an error if they are
	// provided.

	go cl.schedule(ctx, path)

	var rangeBegin uint64
	var data []byte

	for group := range cl.queue {
		bundle := group.bundle
		cacheQty := len(group.notify)
		chunkResults := make([]*blockRequest, cacheQty)
		cacheConns := make([]*Downloader, cacheQty)

		// We receive either an error (in which case we abort the other requests and return an error to our parent) or we
		// receive singly-encrypted data from each cache.
		for i, notify := range group.notify {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case result := <-notify:
				b := result.resp
				chunkResults[i] = b
				cacheConns[i] = result.cache
				if b.err != nil {
					return nil, errors.Wrap(b.err, "got error in block request; aborting any that remain in group")
				}
			}
		}

		cl.l.Info("got singly-encrypted data from each cache")
		bg, err := cl.decryptPuzzle(ctx, bundle, chunkResults, cacheConns)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decrypt puzzle")
		}

		for i, d := range bg.data {
			if bg.blockIdx[i] != rangeBegin+uint64(i) {
				return nil, fmt.Errorf("block at position %v has index %v, but expected %v",
					i, bg.blockIdx[i], rangeBegin+uint64(i))
			}
			cl.l.WithFields(logrus.Fields{
				"len(outputBuffer)": len(data),
				"len(newBlock)":     len(d),
			}).Debug("appending data block to output buffer")
			data = append(data, d...)
		}

		rangeBegin += uint64(len(bg.data))
	}

	return &object{
		data: data,
	}, nil
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
	cl.cacheConns = make(map[cacheID]*Downloader)

	if err := cl.publisherConn.Close(ctx); err != nil {
		retErr = err
	}

	cl.l.Info("client.Close() - exit")
	return retErr
}

type object struct {
	data []byte
}

var _ Object = (*object)(nil)

func (o *object) Data() []byte {
	return o.data
}

func (cl *client) decryptPuzzle(ctx context.Context, bundle *ccmsg.TicketBundle, chunkResults []*blockRequest, cacheConns []*Downloader) (*blockGroup, error) {
	// Solve colocation puzzle.
	tt := common.StartTelemetryTimer(cl.l, "solvePuzzle")
	var singleEncryptedBlocks [][]byte
	for _, result := range chunkResults {
		singleEncryptedBlocks = append(singleEncryptedBlocks, result.encData)
	}
	pi := bundle.Remainder.PuzzleInfo
	secret, _, err := colocationpuzzle.Solve(colocationpuzzle.Parameters{
		Rounds:      pi.Rounds,
		StartOffset: uint32(pi.StartOffset),
		StartRange:  uint32(pi.StartRange),
	}, singleEncryptedBlocks, pi.Goal)
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

	// Decrypt singly-encrypted blocks to produce final plaintext.
	tt = common.StartTelemetryTimer(cl.l, "decryptData")
	var plaintextBlocks [][]byte
	var blockIdx []uint64
	for i, ciphertext := range singleEncryptedBlocks {
		plaintext, err := util.EncryptDataBlock(
			bundle.TicketRequest[i].BlockIdx,
			bundle.Remainder.RequestSequenceNo,
			ticketL2.InnerSessionKey[i].Key,
			ciphertext)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decrypt singly-encrypted block")
		}
		plaintextBlocks = append(plaintextBlocks, plaintext)
		blockIdx = append(blockIdx, bundle.TicketRequest[i].BlockIdx)
	}
	tt.Stop()

	// Return data to parent.
	cl.l.Info("block-group fetch completed without error")
	return &blockGroup{
		data:     plaintextBlocks,
		blockIdx: blockIdx,
		metadata: bundle.Metadata,
	}, nil
}
