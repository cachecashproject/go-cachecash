package publisher

import (
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"database/sql"
	"fmt"

	"github.com/cachecashproject/go-cachecash/batchsignature"
	"github.com/cachecashproject/go-cachecash/catalog"
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/publisher/models"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"golang.org/x/crypto/ed25519"
)

type ContentPublisher struct {
	// The ContentPublisher knows each cache's "inner master key" (aka "master key")?  This is an AES key.
	// For each cache, it also knows an IP address, a port number, and a public key.

	l  *logrus.Logger
	db *sql.DB

	signer  ed25519.PrivateKey
	catalog catalog.ContentCatalog

	escrows []*Escrow

	// XXX: It's obviously not great that this is necessary.
	// Maps object IDs to metadata; necessary to allow the publisher to generate cache-miss responses.
	reverseMapping map[common.ObjectID]reverseMappingEntry

	// XXX: Need cachecash.PublicKey to be an array of bytes, not a slice of bytes, or else we can't use it as a map key
	// caches map[cachecash.PublicKey]*ParticipatingCache
}

type CacheInfo struct {
	// ...
}

type reverseMappingEntry struct {
	path string
}

func NewContentPublisher(l *logrus.Logger, db *sql.DB, catalog catalog.ContentCatalog, signer ed25519.PrivateKey) (*ContentPublisher, error) {
	p := &ContentPublisher{
		l:              l,
		db:             db,
		signer:         signer,
		catalog:        catalog,
		reverseMapping: make(map[common.ObjectID]reverseMappingEntry),
	}

	return p, nil
}

// XXX: Temporary
func (p *ContentPublisher) AddEscrow(escrow *Escrow) error {
	p.escrows = append(p.escrows, escrow)
	return nil
}

// XXX: Temporary
func (p *ContentPublisher) getEscrowByRequest(req *ccmsg.ContentRequest) (*Escrow, error) {
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

  - The _byte range_ is translated to a _block range_ depending on how the publisher would like to chunk the object.
    (Right now, we only support fixed-size blocks, but this is not inherent.)  The publisher may also choose how many
    blocks it would like to serve, and how many block-groups they will be divided into.  (The following steps are
    repeated for each block-group; the results are returned together in a single response to the client.)

  - The object's _path_ is used to ensure that the object exists, and that the specified blocks are in-cache and valid.
    (This may be satisfied by the content catalog's cache, or may require contacting an upstream.)  (A future
    enhancement might require that the publisher fetch only the cipher-blocks that will be used in puzzle generation,
    instead of all of the cipher-blocks in the data blocks.)

  - The _path_ and _block range_ are mapped to a list of _block identifiers_.  These are arbitrarily assigned by the
    publisher.  (Our implementation uses the block's digest.)

  - The publisher selects a single escrow that will be used to service the request.

  - The publisher selects a set of caches that are enrolled in that escrow.  This selection should be designed to place
    the same blocks on the same caches (expanding the number in rotation as demand for the chunks grows), and to reuse
    the same caches for consecutive block-groups served to a single client (so that connection reuse and pipelining can
    improve performance, once implemented).

  - For each cache, the publisher chooses a logical slot index.  (For details, see documentation on the logical cache
    model.)  This slot index should be consistent between requests for the cache to serve the same block.

***************************************************************

XXX: Temporary notes:

Object identifier (path) -> escrow-object (escrow & ID pair; do the IDs really matter?)

    The publisher will probably want to maintain a list of existing escrow-ID pairs for each object;
    it may also, at its option, create a new pair and return that.  (That is, it can choose to serve
    the request out of an escrow that's already been used to serve the object, or it can choose to serve
    the request out of an escrow that hasn't been.)

    This should be designed so that cache rollover/reuse between escrows is possible.

The publisher must also ensure that the metadata and data required to generate the puzzle is available
in the local catalog.  (The publisher doesn't use the catalog yet; that needs to be implemented.)

The publisher will also need to decide on LCM slot IDs for each block it asks a cache to serve.  These can vary per
cache, per escrow.  They should also be designed to support escrow rollover.

*/

const (
	blocksPerGroup = 4
)

func (p *ContentPublisher) HandleContentRequest(ctx context.Context, req *ccmsg.ContentRequest) (*ccmsg.TicketBundle, error) {
	p.l.WithFields(logrus.Fields{
		"path":       req.Path,
		"rangeBegin": req.RangeBegin,
		"rangeEnd":   req.RangeEnd,
	}).Info("content request")

	// - The object's _path_ is used to ensure that the object exists, and that the specified blocks are in-cache and
	//   valid.  (This may be satisfied by the content catalog's cache, or may require contacting an upstream.)  (A
	//   future enhancement might require that the publisher fetch only the cipher-blocks that will be used in puzzle
	//   generation, instead of all of the cipher-blocks in the data blocks.)
	p.l.Debug("pulling metadata and blocks into catalog")
	obj, err := p.catalog.GetData(ctx, &ccmsg.ContentRequest{Path: req.Path})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get metadata for requested object")
	}

	// - The _byte range_ is translated to a _block range_ depending on how the publisher would like to chunk the object.
	//   (Right now, we only support fixed-size blocks, but this is not inherent.)  The publisher may also choose how
	//   many blocks it would like to serve, and how many block-groups they will be divided into.  (The following steps
	//   are repeated for each block-group; the results are returned together in a single response to the client.)
	if req.RangeEnd != 0 && req.RangeEnd <= req.RangeBegin {
		// TODO: Return 4xx, since this is a bad request from the client.
		return nil, errors.New("invalid range")
	}
	rangeBegin := uint64(req.RangeBegin / obj.PolicyBlockSize())
	rangeEnd := uint64(req.RangeEnd / obj.PolicyBlockSize()) // XXX: This probably needs a ceil()

	// XXX: this doesn't work with empty files
	if rangeBegin >= uint64(obj.BlockCount()) {
		return nil, errors.New("rangeBegin beyond last block")
	}

	// TODO: Return multiple block-groups if appropriate.
	rangeEnd = rangeBegin + blocksPerGroup
	if rangeEnd > uint64(obj.BlockCount()) {
		rangeEnd = uint64(obj.BlockCount())
	}

	p.l.WithFields(logrus.Fields{
		"blockRangeBegin": rangeBegin,
		"blockRangeEnd":   rangeEnd,
	}).Info("content request")

	// - The _path_ and _block range_ are mapped to a list of _block identifiers_.  These are arbitrarily assigned by
	// the publisher.  (Our implementation uses the block's digest.)
	p.l.Debug("mapping block indices into block identifiers")
	blockIndices := make([]uint64, 0, rangeEnd-rangeBegin)
	blockIDs := make([]common.BlockID, 0, rangeEnd-rangeBegin)
	for blockIdx := rangeBegin; blockIdx < rangeEnd; blockIdx++ {
		blockID, err := p.getBlockID(obj, blockIdx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get block ID")
		}

		blockIDs = append(blockIDs, blockID)
		blockIndices = append(blockIndices, blockIdx)
	}

	// - The publisher selects a single escrow that will be used to service the request.
	p.l.Debug("selecting escrow")
	escrow, err := p.getEscrowByRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get escrow for request")
	}

	// - The publisher selects a set of caches that are enrolled in that escrow.  This selection should be designed to
	//   place the same blocks on the same caches (expanding the number in rotation as demand for the chunks grows), and
	//   to reuse the same caches for consecutive block-groups served to a single client (so that connection reuse and
	//   pipelining can improve performance, once implemented).
	p.l.Debug("selecting caches")

	ecs, err := models.EscrowCaches(qm.Load("Cache"), qm.Where("escrow_id = ?", escrow.Inner.ID)).All(ctx, p.db)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query database")
	}

	caches := []ParticipatingCache{}
	for _, participant := range ecs {
		caches = append(caches, ParticipatingCache{
			InnerMasterKey: participant.InnerMasterKey,
			Cache:          *participant.R.Cache,
		})
	}

	if len(caches) < len(blockIndices) {
		return nil, errors.New(fmt.Sprintf("not enough caches: have %v; need %v", len(caches), len(blockIndices)))
	}
	caches = caches[0:len(blockIndices)]

	// - For each cache, the publisher chooses a logical slot index.  (For details, see documentation on the logical
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

	// XXX: Should be based on the upstream path, which the current implementation conflates with the request path.
	objID, err := generateObjectID(req.Path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate object ID")
	}
	p.reverseMapping[objID] = reverseMappingEntry{path: req.Path}

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
		ObjectID:          objID,
	}
	for i, blockIdx := range blockIndices {
		// XXX: Need this to be non-zero; otherwise all of our blocks collide!
		bp.Entries = append(bp.Entries, BundleEntryParams{
			TicketNo: ticketNos[i],
			BlockIdx: uint32(blockIdx),
			BlockID:  blockIDs[i],
			Cache:    caches[i],
		})

		b, err := obj.GetBlock(uint32(blockIdx))
		if err != nil {
			return nil, errors.Wrap(err, "failed to get data block")
		}
		bp.PlaintextBlocks = append(bp.PlaintextBlocks, b)
	}

	p.l.Debug("generating and signing bundle")
	batchSigner, err := batchsignature.NewTrivialBatchSigner(p.signer)
	if err != nil {
		return nil, err
	}
	gen := NewBundleGenerator(p.l, batchSigner)
	bundle, err := gen.GenerateTicketBundle(bp)
	if err != nil {
		return nil, err
	}

	// Attach metadata.
	// XXX: This needs to be covered by unit tests.
	bundle.Metadata = &ccmsg.ObjectMetadata{
		BlockSize:  obj.PolicyBlockSize(),
		ObjectSize: obj.ObjectSize(),
	}

	p.l.Debug("done; returning bundle")
	return bundle, nil
}

func (p *ContentPublisher) assignSlot(path string, blockIdx uint64, blockID uint64) uint64 {
	// XXX: should depend on number of slots available to cache, etc.
	return blockIdx
}

// TODO: XXX: Since object policy is, by definition, something that the publisher can set arbitrarily on a per-object
// basis, this should be the only place that these values are hardcoded.
func (p *ContentPublisher) objectPolicy(path string) (*catalog.ObjectPolicy, error) {
	return &catalog.ObjectPolicy{
		BlockSize: 128 * 1024,
	}, nil
}

func (p *ContentPublisher) getBlockID(obj *catalog.ObjectMetadata, blockIdx uint64) (common.BlockID, error) {
	data, err := obj.GetBlock(uint32(blockIdx))
	if err != nil {
		return common.BlockID{}, errors.Wrap(err, "failed to get block data to generate ID")
	}

	var id common.BlockID
	digest := sha512.Sum384(data)
	copy(id[:], digest[0:common.BlockIDSize])

	p.l.WithFields(logrus.Fields{
		"blockIdx": blockIdx,
		"blockID":  id,
	}).Debug("generating block ID")

	return id, nil
}

func (p *ContentPublisher) CacheMiss(ctx context.Context, req *ccmsg.CacheMissRequest) (*ccmsg.CacheMissResponse, error) {
	// TODO: How do we identify the cache submitting the request?

	objectID, err := common.BytesToObjectID(req.ObjectId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to interpret object ID")
	}

	rme, ok := p.reverseMapping[objectID]
	if !ok {
		return nil, errors.New("no reverse mapping found for object ID")
	}
	path := rme.path

	if req.RangeEnd != 0 && req.RangeEnd <= req.RangeBegin {
		return nil, errors.New("invalid range")
	}
	// if req.RangeEnd <= number-of-blocks-in-object ... invalid range

	objMeta, err := p.catalog.GetMetadata(ctx, path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get metadata for object")
	}

	// Convert object policy, which is required to convert block range into byte range.
	pol, err := p.objectPolicy(path)
	if err != nil {
		return nil, errors.New("failed to get object policy")
	}

	resp := ccmsg.CacheMissResponse{
		Chunks: []*ccmsg.Chunk{},
	}

	// XXX: Shouldn't we be telling the cache what block IDs it should expect, and providing enough information for it
	// to verify that it's getting the right data (e.g. a digest)?

	// Select logical cache slot for each block.
	for i := req.RangeBegin; i < req.RangeEnd; i++ {
		blockID := i // XXX: Not true!
		slotIdx := p.assignSlot(path, i, blockID)
		chunk, err := p.catalog.BlockSource(ctx, req, path, objMeta)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get block source")
		}

		// TODO: we shouldn't need to modify the chunk afterwards
		// chunk.BlockId = BlockID,
		chunk.SlotIdx = slotIdx

		resp.Chunks = append(resp.Chunks, chunk)
	}

	resp.Metadata = &ccmsg.ObjectMetadata{
		ObjectSize: objMeta.ObjectSize(),
		BlockSize:  uint64(pol.BlockSize),
	}

	return &resp, nil
}

// XXX: This is, obviously, temporary.  We should be using object IDs that are larger than 64 bits, among other
// problems.  We also must account for the fact that the object stored at a path may change (e.g. when the mtime/etag
// are updated).
func generateObjectID(path string) (common.ObjectID, error) {
	digest := sha256.Sum256([]byte(path))
	return common.BytesToObjectID(digest[0:common.ObjectIDSize])
}
