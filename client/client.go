package client

import (
	"context"

	cachecash "github.com/kelleyk/go-cachecash"
	"github.com/kelleyk/go-cachecash/ccmsg"
	"github.com/kelleyk/go-cachecash/colocationpuzzle"
	"github.com/kelleyk/go-cachecash/common"
	"github.com/kelleyk/go-cachecash/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ed25519"
)

/*
The API of client should mirror that of `http` and `ctxhttp`:
- https://godoc.org/pkg/http
- https://godoc.org/golang.org/x/net/context/ctxhttp

Simplifications made in this implementation of the client:
- A client interacts with a single content provider.
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

// What if one block is a member of multiple block-groups?
type client struct {
	l *logrus.Logger

	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey

	providerConn *providerConnection

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
	pc, err := newProviderConnection(ctx, l, addr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to provider")
	}

	return &client{
		l:          l,
		publicKey:  pub,
		privateKey: priv,

		providerConn: pc,

		cacheConns: make(map[cacheID]*cacheConnection),
	}, nil
}

func (cl *client) GetObject(ctx context.Context, path string) (Object, error) {

	/*
		uri, err := url.ParseRequestURI(rawURI)
		if err != nil {
			return errors.Wrap(err, "failed to parse URI")
		}

		if uri.Scheme != "cachecash" {
			return errors.New("invalid URI: unexpected scheme")
		}

		// TODO: User, RawQuery, Fragment are currently ignored.  Should either pass to server or throw an error if they are
		// provided.
	*/

	// TODO: No spec for this---somehow we need to get metadata describing how many blocks there are to fetch.
	// TODO: Let's assume the object is four blocks, which is true for our hardwired test object.

	// bgr := cl.requestBlockGroup(ctx,

	// XXX: This is hardwired for test purposes.
	_, err := cl.requestBlockGroup(ctx, path)
	return nil, err
}

func (cl *client) Close(ctx context.Context) error {
	// XXX: Open question: how should cache/provider connection management work?  Probably needs to be a setting, or on
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

	if err := cl.providerConn.Close(ctx); err != nil {
		retErr = err
	}

	cl.l.Info("client.Close() - exit")
	return retErr
}

type object struct {
}

var _ Object = (*object)(nil)

func (o *object) Data() []byte {
	return []byte{}
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

func (cl *client) requestBlockGroup(ctx context.Context, path string) ([][]byte, error) {
	req := &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            path,
		RangeBegin:      0,
		RangeEnd:        0, // "continue to the end of the object"
	}

	// Send request to provider; get TicketBundle in response.
	resp, err := cl.providerConn.grpcClient.GetContent(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to request bundle from provider")
	}
	bundle := resp.Bundle
	cl.l.Infof("got ticket bundle from provider: %v", bundle)

	cacheQty := len(bundle.CacheInfo)
	// cacheQty = 1 // XXX: Make troubleshooting easier in development!

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
		_, err = cacheConns[i].grpcClient.ExchangeTicketL2(ctx, req)
		if err != nil {
			// TODO: This should not cause us to abort sending the L2 ticket to other caches, and should not prevent us
			// from returning the plaintext data.
			return nil, errors.Wrap(err, "failed to send L2 ticket to cache")
		}
	}

	// Decrypt singly-encrypted blocks to produce final plaintext.
	var plaintextBlocks [][]byte
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
	}

	// Return data to parent.
	cl.l.Info("block-group fetch completed without error")
	return plaintextBlocks, nil
}

type l2Result struct {
	idx int
	err error
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
	msgData, err := cc.grpcClient.GetBlock(ctx, reqData)
	if err != nil {
		return errors.Wrap(err, "failed to exchange request ticket with cache")
	}
	cl.l.Infof("got data response from cache")

	// Send L1 ticket to cache; await outer decryption key.
	reqL1, err := b.bundle.BuildClientCacheRequest(b.bundle.TicketL1[b.idx])
	if err != nil {
		return errors.Wrap(err, "failed to build client-cache request")
	}
	msgL1, err := cc.grpcClient.ExchangeTicketL1(ctx, reqL1)
	if err != nil {
		return errors.Wrap(err, "failed to exchange request ticket with cache")
	}
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
