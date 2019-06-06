package client

import (
	"context"
	"fmt"

	cachecash "github.com/cachecashproject/go-cachecash"
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
	cacheConns map[cacheID]*cacheConnection
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

		cacheConns: make(map[cacheID]*cacheConnection),
	}, nil
}

func (cl *client) GetObject(ctx context.Context, path string) (Object, error) {
	// TODO: User, RawQuery, Fragment are currently ignored.  Should either pass to server or throw an error if they are
	// provided.

	queue := make(chan *ccmsg.TicketBundle, 1)
	go func() {
		defer close(queue)

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
			queue <- bundle
			rangeBegin += chunks

			if rangeBegin >= bundle.Metadata.BlockCount() {
				cl.l.Info("got all bundles")
				break
			}
			blockSize = bundle.Metadata.BlockSize
		}
	}()

	var rangeBegin uint64
	var data []byte
	for bundle := range queue {
		// XXX: `rangeBegin` here must be in bytes.
		bg, err := cl.requestBlockGroup(ctx, path, bundle)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to fetch block-group at offset %v", rangeBegin)
		}

		// XXX: There's lots of copying going on here that is probably unnecessary.
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
	cl.cacheConns = make(map[cacheID]*cacheConnection)

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

/*
// Dispatches to cacheConnection.requestBlock() for the correct cache.  If no connection exists for the cache, creates
// one.
func (cl *client) requestBlock(ctx context.Context, cid cacheID, req *blockRequest) (bool, error) {
	return false, errors.New("no impl")
}
*/

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

func (cl *client) requestBlockGroup(ctx context.Context, path string, bundle *ccmsg.TicketBundle) (*blockGroup, error) {
	cacheQty := len(bundle.CacheInfo)

	// For each block in TicketBundle, dispatch a request to the appropriate cache.
	var cacheConns []*cacheConnection
	blockResults := make([]*blockRequest, cacheQty)
	blockResultCh := make(chan *blockRequest)
	for i := 0; i < cacheQty; i++ {
		cid := (cacheID)(bundle.CacheInfo[i].Addr.ConnectionString())

		cc, ok := cl.cacheConns[cid]
		if !ok {
			var err error
			// XXX: It's problematic to pass ctx here, because canceling the context will destroy the cache connections!
			// (It should only cancel this particular block-group request.)
			cc, err = newCacheConnection(context.Background(), cl.l, bundle.CacheInfo[i].Addr.ConnectionString())
			if err != nil {
				return nil, errors.Wrap(err, "failed to connect to cache")
			}
			cl.cacheConns[cid] = cc
		}
		cacheConns = append(cacheConns, cc)

		blockResults[i] = &blockRequest{
			bundle: bundle,
			idx:    i,
		}
		go cl.requestBlock(ctx, cc, blockResults[i], blockResultCh)
	}

	// We receive either an error (in which case we abort the other requests and return an error to our parent) or we
	// receive singly-encrypted data from each cache.
	for i := 0; i < cacheQty; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case b := <-blockResultCh:
			if b.err != nil {
				return nil, errors.Wrap(b.err, "got error in block request; aborting any that remain in group")
			}
		}
		// ..
	}

	cl.l.Info("got singly-encrypted data from each cache")

	// Solve colocation puzzle.
	tt := common.StartTelemetryTimer(cl.l, "solvePuzzle")
	var singleEncryptedBlocks [][]byte
	for i := 0; i < cacheQty; i++ {
		singleEncryptedBlocks = append(singleEncryptedBlocks, blockResults[i].encData)
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
	for i := 0; i < cacheQty; i++ {
		req, err := bundle.BuildClientCacheRequest(&ccmsg.TicketL2Info{
			EncryptedTicketL2: bundle.EncryptedTicketL2,
			PuzzleSecret:      secret,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to build L2 ticket request")
		}
		tt := common.StartTelemetryTimer(cl.l, "exchangeTicketL2")
		_, err = cacheConns[i].grpcClient.ExchangeTicketL2(ctx, req)
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

func (cl *client) requestBlock(ctx context.Context, cc *cacheConnection, b *blockRequest, blockResultCh chan<- *blockRequest) {
	if err := cl.requestBlockC(ctx, cc, b); err != nil {
		cc.l.WithError(err).Error("error in requestBlockC")
		b.err = err
	}
	blockResultCh <- b
}

func (cl *client) requestBlockC(ctx context.Context, cc *cacheConnection, b *blockRequest) error {
	// Send request ticket to cache; await data.
	reqData, err := b.bundle.BuildClientCacheRequest(b.bundle.TicketRequest[b.idx])
	if err != nil {
		return errors.Wrap(err, "failed to build client-cache request")
	}
	tt := common.StartTelemetryTimer(cl.l, "getBlock")
	msgData, err := cc.grpcClient.GetBlock(ctx, reqData)
	if err != nil {
		return errors.Wrap(err, "failed to exchange request ticket with cache")
	}
	tt.Stop()
	cl.l.WithFields(logrus.Fields{
		"blockIdx": b.bundle.TicketRequest[b.idx].BlockIdx,
		"len":      len(msgData.Data),
	}).Infof("got data response from cache")

	// Send L1 ticket to cache; await outer decryption key.
	reqL1, err := b.bundle.BuildClientCacheRequest(b.bundle.TicketL1[b.idx])
	if err != nil {
		return errors.Wrap(err, "failed to build client-cache request")
	}
	tt = common.StartTelemetryTimer(cl.l, "exchangeTicketL1")
	msgL1, err := cc.grpcClient.ExchangeTicketL1(ctx, reqL1)
	if err != nil {
		return errors.Wrap(err, "failed to exchange request ticket with cache")
	}
	tt.Stop()
	cl.l.Infof("got L1 response from cache")

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
