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

	reverseMapping map[uint64]reverseMappingEntry

	// XXX: Need cachecash.PublicKey to be an array of bytes, not a slice of bytes, or else we can't use it as a map key
	// caches map[cachecash.PublicKey]*ParticipatingCache
}

type CacheInfo struct {
	// ...
}

type reverseMappingEntry struct {
	path     string
	blockIdx uint64
}

func NewContentProvider(l *logrus.Logger, catalog catalog.ContentCatalog, signer crypto.Signer) (*ContentProvider, error) {
	p := &ContentProvider{
		l:              l,
		signer:         signer,
		catalog:        catalog,
		reverseMapping: make(map[uint64]reverseMappingEntry),
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
The process of satisfying a request

  - A request arrives for an object, identified by a _path_, which is actually an opaque series of bytes.  (In our
    implementation, they're HTTP-like paths.)  It (optionally) includes a _byte range_, which may be open-ended (which
    means "continue until the end of the object").

  - The _byte range_ is translated to a _block range_ depending on how the provider would like to chunk the object.
    (Right now, we only support fixed-size blocks, but this is not inherent.)  The provider may also choose how many
    blocks it would like to serve, and how many block-groups they will be divided into.  (The following steps are
    repeated for each block-group; the results are returned together in a single response to the client.)

  - The object's _path_ is used to ensure that the object exists, and that the specified blocks are in-cache and valid.
    (This may be satisfied by the content catalog's cache, or may require contacting an upstream.)  (A future
    enhancement might require that the provider fetch only the cipher-blocks that will be used in puzzle generation,
    instead of all of the cipher-blocks in the data blocks.)

  - The _path_ and _block range_ are mapped to a list of _block identifiers_.  These are arbitrarily assigned by the
    provider.  (Our implementation uses the block's digest.)

  - The provider selects a single escrow that will be used to service the request.

  - The provider selects a set of caches that are enrolled in that escrow.  This selection should be designed to place
    the same blocks on the same caches (expanding the number in rotation as demand for the chunks grows), and to reuse
    the same caches for consecutive block-groups served to a single client (so that connection reuse and pipelining can
    improve performance, once implemented).

  - For each cache, the provider chooses a logical slot index.  (For details, see documentation on the logical cache
    model.)  This slot index should be consistent between requests for the cache to serve the same block.

***************************************************************

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

const (
	defaultBlockSize = 512 * 1024
	blocksPerGroup   = 4
)

func (p *ContentProvider) HandleContentRequest(ctx context.Context, req *ccmsg.ContentRequest) (*ccmsg.TicketBundle, error) {
	p.l.WithFields(logrus.Fields{"path": req.Path}).Info("content request")

	// - The _byte range_ is translated to a _block range_ depending on how the provider would like to chunk the object.
	//   (Right now, we only support fixed-size blocks, but this is not inherent.)  The provider may also choose how
	//   many blocks it would like to serve, and how many block-groups they will be divided into.  (The following steps
	//   are repeated for each block-group; the results are returned together in a single response to the client.)
	if req.RangeEnd != 0 && req.RangeEnd <= req.RangeBegin {
		// TODO: Return 4xx, since this is a bad request from the client.
		return nil, errors.New("invalid range")
	}
	// XXX: We also have `ContentObject.BlockSize()`; should pick one or the other.
	rangeBegin := uint64(req.RangeBegin / defaultBlockSize)
	rangeEnd := uint64(req.RangeEnd / defaultBlockSize) // XXX: This probably needs a ceil()
	// TODO: Return multiple block-groups if appropriate.
	rangeEnd = rangeBegin + blocksPerGroup

	// - The object's _path_ is used to ensure that the object exists, and that the specified blocks are in-cache and
	//   valid.  (This may be satisfied by the content catalog's cache, or may require contacting an upstream.)  (A
	//   future enhancement might require that the provider fetch only the cipher-blocks that will be used in puzzle
	//   generation, instead of all of the cipher-blocks in the data blocks.)
	p.l.Debug("pulling metadata and blocks into catalog")
	objMeta, err := p.catalog.GetObjectMetadata(ctx, req.Path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get metadata for requested object")
	}

	// - The _path_ and _block range_ are mapped to a list of _block identifiers_.  These are arbitrarily assigned by
	// the provider.  (Our implementation uses the block's digest.)
	p.l.Debug("mapping block indices into block identifiers")
	blockIDs := make([]uint64, 0, rangeEnd-rangeBegin)
	for blockIdx := rangeBegin; blockIdx < rangeEnd; blockIdx++ {
		//blockIDs = append(blockIDs, objMeta.GetBlockID(blockIdx))
		// XXX: TEMP: Use block index as block ID.  This WILL NOT WORK as soon as we have multiple objects.
		// XXX: TEMP: The need to maintain this mapping is a flaw.
		blockID := blockIdx
		p.reverseMapping[blockIdx] = reverseMappingEntry{path: req.Path, blockIdx: blockIdx}
		blockIDs = append(blockIDs, blockID)
	}

	// - The provider selects a single escrow that will be used to service the request.
	p.l.Debug("selecting escrow")
	escrow, err := p.getEscrowByRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get escrow for request")
	}

	// - The provider selects a set of caches that are enrolled in that escrow.  This selection should be designed to
	//   place the same blocks on the same caches (expanding the number in rotation as demand for the chunks grows), and
	//   to reuse the same caches for consecutive block-groups served to a single client (so that connection reuse and
	//   pipelining can improve performance, once implemented).
	p.l.Debug("selecting caches")
	if len(escrow.Caches) < len(blockIDs) {
		return nil, errors.New(fmt.Sprintf("not enough caches: have %v; need %v", len(escrow.Caches), len(blockIDs)))
	}
	caches := escrow.Caches[0:len(blockIDs)]

	// - For each cache, the provider chooses a logical slot index.  (For details, see documentation on the logical
	//   cache model.)  This slot index should be consistent between requests for the cache to serve the same block.

	// *********************************************************

	/*
		// XXX: If the object doesn't exist, we shouldn't reserve ticket numbers to satisfy the request!
		// XXX: This is what we need to remove to make the change to the content catalog; the `obj` here allows access to
		//      the entire contents of the object!
		obj, objID, err := escrow.GetObjectByPath(ctx, req.Path)
		if err != nil {
			return nil, errors.Wrap(err, "no object for path")
		}
	*/
	var objID uint64
	obj := objMeta // XXX: ...

	// Reserve a lottery ticket for each cache.  (Recall that lottery ticket numbers must be unique, and we are limited
	// in the number that we can issue during each blockchain block to the number that we declared in our begin-escrow
	// transaction.)
	// XXX: We need to make sure that these numbers are released to be reused if the request fails.
	p.l.Debug("reserving tickets")
	ticketNos, err := escrow.reserveTicketNumbers(len(caches))
	if err != nil {
		return nil, errors.Wrap(err, "failed to reserve ticket numbers")
	}

	p.l.Debug("building bundle parameters")
	bp := &BundleParams{
		Escrow:            escrow,
		RequestSequenceNo: req.SequenceNo,
		ClientPublicKey:   ed25519.PublicKey(req.ClientPublicKey.PublicKey),
		Object:            obj,
		ObjectID:          objID,
	}
	for i, bid := range blockIDs {
		bp.Entries = append(bp.Entries, BundleEntryParams{
			TicketNo: ticketNos[i],
			// XXX: fix typing; also, we're stuffing a block ID into a block index!  Should we change the message to use
			// block ID, or the code to provide block index?
			BlockIdx: uint32(bid),
			Cache:    caches[i],
		})
	}

	p.l.Debug("generating and signing bundle")
	batchSigner, err := batchsignature.NewTrivialBatchSigner(p.signer)
	if err != nil {
		return nil, err
	}
	gen := NewBundleGenerator(batchSigner)
	bundle, err := gen.GenerateTicketBundle(bp)
	if err != nil {
		return nil, err
	}

	p.l.Debug("done; returning bundle")
	return bundle, nil
}

func (p *ContentProvider) assignSlot(path string, blockIdx uint64) uint64 {
	// XXX: should depend on number of slots available to cache, etc.
	return blockIdx
}

func (p *ContentProvider) CacheMiss(ctx context.Context, req *ccmsg.CacheMissRequest) (*ccmsg.CacheMissResponse, error) {
	// TODO: How do we identify the cache submitting the request?

	/*
			// First, at least for the HTTP upstream, we need to map the block ID back to a (path, block index) tuple.
			//
			// TODO: This is clearly not a great thing to make the publisher maintain; should investigate ways to remove this
			// requirement.
			//
			// TODO: Should also verify that this cache has a reason to be requesting this content.  Could have the cache
			// provide the request ticket from the client.  Might also be a good way to avoid the reverse-mapping issue.
			rme, ok := p.reverseMapping[req.BlockId]
			if !ok {
				return nil, errors.New("no reverse mapping found for block ID")
			}

			// XXX:
			blockIdx := rme.blockIdx
			path := rme.path

		// Get source information from the upstream module (in the case of HTTP, that's a URL and byte range)/
		msg, err := p.catalog.CacheMiss(path, blockIdx, blockIdx+1)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate source information")
		}

		// Select logical cache slots for each block.
		msg.SlotIdx = []uint64{p.assignSlot(path, blockIdx)}

		return msg, nil
	*/

	panic("no impl")
}
