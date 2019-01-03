package provider

import (
	"context"
	"crypto"
	"fmt"

	"github.com/kelleyk/go-cachecash/batchsignature"
	"github.com/kelleyk/go-cachecash/catalog"
	"github.com/kelleyk/go-cachecash/ccmsg"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ed25519"
)

type ContentProvider struct {
	// The ContentProvider knows each cache's "inner master key" (aka "master key")?  This is an AES key.
	// For each cache, it also knows an IP address, a port number, and a public key.

	l *logrus.Logger

	signer  crypto.Signer
	catalog catalog.ContentCatalog

	escrows []*Escrow

	// XXX: Need cachecash.PublicKey to be an array of bytes, not a slice of bytes, or else we can't use it as a map key
	// caches map[cachecash.PublicKey]*ParticipatingCache
}

type CacheInfo struct {
	// ...
}

func NewContentProvider(l *logrus.Logger, catalog catalog.ContentCatalog, signer crypto.Signer) (*ContentProvider, error) {
	p := &ContentProvider{
		l:       l,
		signer:  signer,
		catalog: catalog,
	}

	return p, nil
}

// XXX: Temporary
func (p *ContentProvider) AddEscrow(escrow *Escrow) error {
	p.escrows = append(p.escrows, escrow)
	return nil
}

// XXX: Temporary
func (p *ContentProvider) getEscrowByRequest(req *ccmsg.ContentRequest) (*Escrow, error) {
	if len(p.escrows) == 0 {
		return nil, errors.New("no escrow for request")
	}
	return p.escrows[0], nil
}

/*
XXX: Temporary notes:

Object identifier (path) -> escrow-object (escrow & ID pair; do the IDs really matter?)

    The provider will probably want to maintain a list of existing escrow-ID pairs for each object;
    it may also, at its option, create a new pair and return that.  (That is, it can choose to serve
    the request out of an escrow that's already been used to serve the object, or it can choose to serve
    the request out of an escrow that hasn't been.)

    This should be designed so that cache rollover/reuse between escrows is possible.

The provider must also ensure that the metadata and data required to generate the puzzle is available
in the local catalog.  (The provider doesn't use the catalog yet; that needs to be implemented.)

The provider will also need to decide on LCM slot IDs for each block it asks a cache to serve.  These can vary per
cache, per escrow.  They should also be designed to support escrow rollover.

*/
func (p *ContentProvider) HandleContentRequest(ctx context.Context, req *ccmsg.ContentRequest) (*ccmsg.TicketBundle, error) {
	p.l.WithFields(logrus.Fields{"path": req.Path}).Info("content request")

	// Validate request.
	// - The request is for a list of blocks of a particular object.
	// - We have an valid escrow with enough active participating caches to serve the client.
	// XXX: TODO:

	// Select the escrow that will be used to serve the request.
	// TODO:
	escrow, err := p.getEscrowByRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get escrow for request")
	}

	// Select the caches that we'd like to use to serve the requested content.
	// XXX: TODO:
	if len(escrow.Caches) < len(req.BlockIdx) {
		return nil, errors.New(fmt.Sprintf("not enough caches: have %v; need %v", len(escrow.Caches), len(req.BlockIdx)))
	}
	caches := escrow.Caches[0:len(req.BlockIdx)]

	// Reserve a lottery ticket for each cache.  (Recall that lottery ticket numbers must be unique, and we are limited
	// in the number that we can issue during each blockchain block to the number that we declared in our begin-escrow
	// transaction.)
	// XXX: We need to make sure that these numbers are released to be reused if the request fails.
	ticketNos, err := escrow.reserveTicketNumbers(len(caches))
	if err != nil {
		return nil, errors.Wrap(err, "failed to reserve ticket numbers")
	}

	// XXX: If the object doesn't exist, we shouldn't reserve ticket numbers to satisfy the request!
	obj, objID, err := escrow.GetObjectByPath(ctx, req.Path)
	if err != nil {
		return nil, errors.Wrap(err, "no object for path")
	}

	bp := &BundleParams{
		Escrow:            escrow,
		RequestSequenceNo: req.SequenceNo,
		ClientPublicKey:   ed25519.PublicKey(req.ClientPublicKey.PublicKey),
		Object:            obj,
		ObjectID:          objID,
	}
	for i, idx := range req.BlockIdx {
		bp.Entries = append(bp.Entries, BundleEntryParams{
			TicketNo: ticketNos[i],
			BlockIdx: uint32(idx), // XXX: fix typing
			Cache:    caches[i],
		})
	}

	batchSigner, err := batchsignature.NewTrivialBatchSigner(p.signer)
	if err != nil {
		return nil, err
	}
	gen := NewBundleGenerator(batchSigner)
	bundle, err := gen.GenerateTicketBundle(bp)
	if err != nil {
		return nil, err
	}

	return bundle, nil
}
