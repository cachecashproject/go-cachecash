package publisher

import (
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"math"
	"sort"

	"github.com/cachecashproject/go-cachecash/batchsignature"
	"github.com/cachecashproject/go-cachecash/catalog"
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/dbtx"
	"github.com/cachecashproject/go-cachecash/publisher/models"
	"github.com/dchest/siphash"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"golang.org/x/crypto/ed25519"
)

// publisherCache tracks the publishers data about a cache
type publisherCache struct {
	// can store multiple different length permutations as a single cache may be
	// in escrows of very different sizes - permutation size is set at the first
	// prime > 2**N that is at least 75x the escrows cache count. This usage is
	// to obtain the same permutation from different but similar size escrows
	// for efficiency, with more room for reuse and variation the larger the
	// escrows have grown.. The 75x ensures a minimum lookup count large enough
	// to be effective.
	permutations  map[int][]uint64
	participation *ParticipatingCache
}

// ContentPublisher is the main state for the publisher daemon.
//
// During startup the CLI entry point populates this by calling
// LoadFromDatabase, and the intent is that from that point all state is
// cached in RAM but the database is the source of truth with soft degradation.
// Some code may be inconsistent with this design principle - please fix if
// noticed.
// Reasoning:
// - for a single escrow cache counts are anticipated to be O(100's)
// - for a single publisher escrow counts are anticipated to be O(10's)
// - at least for the early / medium term: the network should support a great
//   many publishers and a great many escrows of course, but this component
//   itself is working on a 'fits in RAM' problem and thus we can optimise it
//   to be low latency
// - the publisher needs to be available for clients to obtain content, and
//   AWS performs regular outages to postgreSQL as part of regular maintenance and operations.
// - once truth is established the cached in RAM data can be operated on very quickly
//   e.g. the contents of an escrow cannot change after establishment
type ContentPublisher struct {
	// The ContentPublisher knows each cache's "inner master key" (aka "master key")?  This is an AES key.
	// For each cache, it also knows an IP address, a port number, and a public key.

	l *logrus.Logger

	signer  ed25519.PrivateKey
	catalog catalog.ContentCatalog

	escrows []*Escrow
	caches  map[string]*publisherCache

	// XXX: It's obviously not great that this is necessary.
	// Maps object IDs to metadata; necessary to allow the publisher to generate cache-miss responses.
	reverseMapping map[common.ObjectID]reverseMappingEntry

	PublisherAddr string
}

type reverseMappingEntry struct {
	path string
}

func NewContentPublisher(l *logrus.Logger, publisherAddr string, catalog catalog.ContentCatalog, signer ed25519.PrivateKey) (*ContentPublisher, error) {
	p := &ContentPublisher{
		l:              l,
		signer:         signer,
		catalog:        catalog,
		caches:         make(map[string]*publisherCache),
		reverseMapping: make(map[common.ObjectID]reverseMappingEntry),
		PublisherAddr:  publisherAddr,
	}

	return p, nil
}

func (p *ContentPublisher) LoadFromDatabase(ctx context.Context) (int, error) {
	escrows, err := models.Escrows().All(ctx, dbtx.ExecutorFromContext(ctx))
	if err != nil {
		return 0, errors.Wrap(err, "failed to query Escrows")
	}

	for _, e := range escrows {
		escrow := &Escrow{
			Inner:  *e,
			Caches: []*ParticipatingCache{},
		}

		// The default retrieval order would typically be the ID, but DB
		// clustering can cause that to vary, so we specify
		ecs, err := e.EscrowCaches(qm.Load("Cache"), qm.OrderBy(models.EscrowCacheColumns.CacheID)).All(ctx, dbtx.ExecutorFromContext(ctx))
		if err != nil {
			return 0, errors.Wrap(err, "failed to query EscrowsCaches")
		}
		for _, ec := range ecs {
			escrow.Caches = append(escrow.Caches, &ParticipatingCache{
				Cache:          *ec.R.Cache,
				InnerMasterKey: ec.InnerMasterKey,
			})
		}

		err = p.AddEscrow(escrow)
		if err != nil {
			return 0, errors.Wrap(err, "failed to query Cache")
		}
	}

	return len(escrows), nil
}

// AddEscrow - internal helper?
// XXX: Temporary (we have nothing syncing from DB back to memory or maintaining memory integrity)
func (p *ContentPublisher) AddEscrow(escrow *Escrow) error {
	escrow.Publisher = p
	p.escrows = append(p.escrows, escrow)

	sort.Slice(escrow.Caches, func(i, j int) bool {
		return escrow.Caches[i].Cache.ID < escrow.Caches[j].Cache.ID
	})
	// setup a map from pubkey -> *cache
	for _, cache := range escrow.Caches {
		key := string(cache.PublicKey())
		pubCache, ok := p.caches[key]
		if ok {
			// update the cache - new network contact details etc
			pubCache.participation = cache
		} else {
			p.caches[key] = &publisherCache{
				permutations:  make(map[int][]uint64),
				participation: cache}
		}
		_, err := p.getCachePermutation(key, uint(len(escrow.Caches)))
		if err != nil {
			return errors.Wrap(err, "Failed to calculate lookup permutation")
		}
	}
	return escrow.CalculateLookup()
}

func (p *ContentPublisher) getEscrowByRequest(req *ccmsg.ContentRequest) (*Escrow, error) {
	// XXX This should find an escrow in RAM (cache) then fall back to the DB if
	// needed.
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

  - The _byte range_ is translated to a _chunk range_ depending on how the publisher would like to split the object.
    (Right now, we only support fixed-size chunks, but this is not inherent.)  The publisher may also choose how many
    chunks it would like to serve, and how many chunk-groups they will be divided into.  (The following steps are
    repeated for each chunk-group; the results are returned together in a single response to the client.)

  - The object's _path_ is used to ensure that the object exists, and that the specified chunks are in-cache and valid.
    (This may be satisfied by the content catalog's cache, or may require contacting an upstream.)  (A future
    enhancement might require that the publisher fetch only the cipher-blocks that will be used in puzzle generation,
    instead of all of the cipher-blocks in the chunks.)

  - The _path_ and _chunk range_ are mapped to a list of _chunk identifiers_.  These are arbitrarily assigned by the
    publisher.  (Our implementation uses the chunk's digest.)

  - The publisher selects a single escrow that will be used to service the request.

  - The publisher selects a set of caches that are enrolled in that escrow.  This selection should be designed to place
    the same chunks on the same caches (expanding the number in rotation as demand for the chunks grows), and to reuse
    the same caches for consecutive chunk-groups served to a single client (so that connection reuse and pipelining can
    improve performance, once implemented).

  - For each cache, the publisher chooses a logical slot index.  (For details, see documentation on the logical cache
    model.)  This slot index should be consistent between requests for the cache to serve the same chunk.

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

The publisher will also need to decide on LCM slot IDs for each chunk it asks a cache to serve.  These can vary per
cache, per escrow.  They should also be designed to support escrow rollover.

*/

const (
	chunksPerGroup    = 4
	bundlesPerRequest = 3
	// These keys were chosen randomly and are arbitrary unless and until we
	// decide to have multiple publishers collaborating (e.g. scale-out for a
	// single set of escrows). They should stay the same for the life of a
	// publisher at minimum to reduce chunk churn on caches.
	k0 uint64 = 0x103c1888a30d6f7f
	k1 uint64 = 0xea8e301f56febad4
)

// HandleContentRequest serves ticket bundles to clients.
//  This is only exported for integration_test.go.
func (p *ContentPublisher) HandleContentRequest(ctx context.Context, req *ccmsg.ContentRequest) ([]*ccmsg.TicketBundle, error) {
	p.l.WithFields(logrus.Fields{
		"path":       req.Path,
		"rangeBegin": req.RangeBegin,
		"rangeEnd":   req.RangeEnd,
	}).Info("content request")

	// - The _byte range_ is translated to a _chunk range_ depending on how the publisher would like to chunk the object.
	//   (Right now, we only support fixed-size chunks, but this is not inherent.)  The publisher may also choose how
	//   many chunks it would like to serve, and how many chunk-groups they will be divided into.  (The following steps
	//   are repeated for each chunk-group; the results are returned together in a single response to the client.)
	if req.RangeEnd != 0 && req.RangeEnd <= req.RangeBegin {
		// TODO: Return 4xx, since this is a bad request from the client.
		return nil, errors.New("invalid range")
	}

	for cache, status := range req.CacheStatus {
		p.l.WithFields(logrus.Fields{
			"cache":  base64.StdEncoding.EncodeToString([]byte(cache)),
			"status": status,
		}).Debug("received cache status")
	}

	// - The object's _path_ is used to ensure that the object exists, and that the specified chunks are in-cache and
	//   valid.  (This may be satisfied by the content catalog's cache, or may require contacting an upstream.)  (A
	//   future enhancement might require that the publisher fetch only the cipher-blocks that will be used in puzzle
	//   generation, instead of all of the cipher-blocks in the chunks.)
	p.l.Debug("pulling metadata and chunks into catalog")
	obj, err := p.catalog.GetData(ctx, &ccmsg.ContentRequest{Path: req.Path})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get metadata for requested object")
	}
	p.l.WithFields(logrus.Fields{
		"size": obj.ObjectSize(),
	}).Debug("received metadata and chunks")

	// - The publisher selects a single escrow that will be used to service the request.
	p.l.Debug("selecting escrow")
	escrow, err := p.getEscrowByRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get escrow for request")
	}

	p.l.Debug("selecting caches for bundle at offset")
	// path + content byte offset
	// path so that very short files don't all land on the same set of initial caches
	// content byte offset / chunk offset so that the content of files are spread over caches
	// This is v0 - see the content caching doc for future plans.
	prefix := []byte{}
	if req.ClientPublicKey == nil {
		return nil, errors.New("No client public key")
	}
	clientPublicKey := req.ClientPublicKey.PublicKey
	prefix = append(prefix, req.Path...)
	chunkRangeBegin := uint64(req.RangeBegin / obj.PolicyChunkSize())
	sequenceNo := req.SequenceNo

	numberOfBundles := bundlesPerRequest
	if req.RangeEnd != 0 {
		// TODO: if we need more than one bundle to reach RangeEnd, calculate
		// some number of bundles to balance read-ahead and latency
		numberOfBundles = 1
	}

	bundles := []*ccmsg.TicketBundle{}
	for bundleIdx := 0; bundleIdx < numberOfBundles; bundleIdx++ {
		bundle, err := p.generateBundle(ctx, escrow, obj, req.Path, chunkRangeBegin, clientPublicKey, sequenceNo, prefix, &req.CacheStatus)
		if err != nil {
			return nil, err
		}
		chunkRangeBegin += uint64(len(bundle.GetTicketRequest()))
		sequenceNo++
		bundles = append(bundles, bundle)

		if chunkRangeBegin >= uint64(obj.ChunkCount()) {
			// we've reached the end of the object
			break
		}
	}

	return bundles, nil
}

func (p *ContentPublisher) generateBundle(ctx context.Context, escrow *Escrow, obj *catalog.ObjectMetadata, path string,
	chunkRangeBegin uint64, clientPublicKey ed25519.PublicKey, sequenceNo uint64, prefix []byte,
	cacheStatus *map[string]*ccmsg.ContentRequest_ClientCacheStatus) (*ccmsg.TicketBundle, error) {
	// XXX: this doesn't work with empty files
	if chunkRangeBegin >= uint64(obj.ChunkCount()) {
		return nil, errors.New("chunkRangeBegin beyond last chunk")
	}

	// TODO: Return multiple chunk-groups if appropriate.
	rangeEnd := chunkRangeBegin + chunksPerGroup
	if rangeEnd > uint64(obj.ChunkCount()) {
		rangeEnd = uint64(obj.ChunkCount())
	}

	p.l.WithFields(logrus.Fields{
		"chunkRangeBegin": chunkRangeBegin,
		"chunkRangeEnd":   rangeEnd,
	}).Info("content request")

	// - The _path_ and _chunk range_ are mapped to a list of _chunk identifiers_.  These are arbitrarily assigned by
	// the publisher.  (Our implementation uses the chunk's digest.)
	p.l.Debug("mapping chunk indices into chunk identifiers")
	chunkIndices := make([]uint64, 0, rangeEnd-chunkRangeBegin)
	chunkIDs := make([]common.ChunkID, 0, rangeEnd-chunkRangeBegin)
	for chunkIdx := chunkRangeBegin; chunkIdx < rangeEnd; chunkIdx++ {
		chunkID, err := p.getChunkID(obj, chunkIdx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get chunk ID")
		}

		chunkIDs = append(chunkIDs, chunkID)
		chunkIndices = append(chunkIndices, chunkIdx)
	}

	// XXX: Should be based on the upstream path, which the current implementation conflates with the request path.
	objID, err := generateObjectID(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate object ID")
	}
	p.reverseMapping[objID] = reverseMappingEntry{path: path}

	// We use the consistent hash to pick the first of the caches in the escrow.
	// Then each cache is assigned in round robin fashion ignoring error status.
	// Then failed caches are replaced with unused good caches also in round
	// robin order.
	// Start: escrow has: A B C D E
	// first index picks (say) C, giving C,D,E,A
	// if A is errored, A is replaced with B giving C,D,E,B.
	// if two caches are errored, an error is given.
	bundleInput := append(prefix, string(chunkRangeBegin)...)
	h64 := siphash.Hash(k0, k1, bundleInput)
	firstIndex := (*escrow.lookup)[h64%uint64(len(*escrow.lookup))]
	caches := []*ParticipatingCache{}
	// Allocate the ideal cache -> chunks without consideration of errors to
	// keep allocation stable
	for i := firstIndex; len(caches) != chunksPerGroup; i = (i + 1) % len(escrow.Caches) {
		cache := escrow.Caches[i]
		for j := 0; j < len(caches); j++ {
			if caches[j] == cache {
				return nil, errors.New("insufficient caches to create Bundle - would have to reuse cache")
			}
		}
		caches = append(caches, cache)
	}
	p.l.WithFields(logrus.Fields{
		"chunkIdx":    chunkRangeBegin,
		"objectID":    objID,
		"bundleInput": bundleInput,
		"h64":         h64,
		"firstIndex":  firstIndex,
		"caches":      caches,
	}).Debug("Allocating Caches")

	// substitute individual faulting caches with other ones with good status
	badcaches := 0
	for i := range caches {
		cache := caches[i]
		if (*cacheStatus)[string(cache.PublicKey())].GetStatus() == ccmsg.ContentRequest_ClientCacheStatus_DEFAULT {
			// Good cache. Could also consider backlog length or other signal.
			continue
		}
		// The next possible cache is located at the hash offset + the caches we
		// allocated in this chunk + the badcaches we consumed thus far.
		offset := firstIndex + chunksPerGroup + badcaches
		for j := 0; j < len(escrow.Caches)-chunksPerGroup-badcaches; j++ {
			index := (offset + j) % len(escrow.Caches)
			cache := escrow.Caches[index]
			if (*cacheStatus)[string(cache.PublicKey())].GetStatus() == ccmsg.ContentRequest_ClientCacheStatus_DEFAULT {
				caches[i] = cache
				break
			}
		}
		cache = caches[i]
		if (*cacheStatus)[string(cache.PublicKey())].GetStatus() != ccmsg.ContentRequest_ClientCacheStatus_DEFAULT {
			// Failed to find a good cache.
			return nil, errors.New("insufficient caches to create Bundle - would have to use cache in error state")
		}
		badcaches++
	}
	p.l.WithFields(logrus.Fields{
		"chunkIdx":    chunkRangeBegin,
		"objectID":    objID,
		"bundleInput": bundleInput,
		"h64":         h64,
		"firstIndex":  firstIndex,
		"caches":      caches,
	}).Debug("Caches after error conditions")

	// XXX: Should be redundant - unwind the differing checks and make sure, then remove.
	if len(caches) < len(chunkIndices) {
		return nil, errors.New(fmt.Sprintf("not enough caches: have %v; need %v", len(caches), len(chunkIndices)))
	}
	// ditto
	caches = caches[0:len(chunkIndices)]

	// - For each cache, the publisher chooses a logical slot index.  (For details, see documentation on the logical
	//   cache model.)  This slot index should be consistent between requests for the cache to serve the same chunk.

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

	// Reserve a lottery ticket for each cache.  (Recall that lottery ticket numbers must be unique, and we are limited
	// in the number that we can issue during each blockchain block to the number that we declared in our begin-escrow
	// transaction.)
	// XXX: We need to make sure that these numbers are released to be reused if the request fails.
	//      <-- robertc: we can't guarante that (things do fail on ec2 etc - does it cost money if this doesn't unwind, or just limit bundles issued?)
	p.l.Debug("reserving tickets")
	ticketNos, err := escrow.reserveTicketNumbers(len(caches))
	if err != nil {
		return nil, errors.Wrap(err, "failed to reserve ticket numbers")
	}

	p.l.Debug("building bundle parameters")
	bp := &BundleParams{
		Escrow:            escrow,
		RequestSequenceNo: sequenceNo,
		ClientPublicKey:   clientPublicKey,
		ObjectID:          objID,
	}
	for i, chunkIdx := range chunkIndices {
		// XXX: Need this to be non-zero; otherwise all of our chunks collide!
		bp.Entries = append(bp.Entries, BundleEntryParams{
			TicketNo: ticketNos[i],
			ChunkIdx: uint32(chunkIdx),
			ChunkID:  chunkIDs[i],
			Cache:    *caches[i],
		})

		b, err := obj.GetChunk(uint32(chunkIdx))
		if err != nil {
			return nil, errors.Wrap(err, "failed to get chunk")
		}
		bp.PlaintextChunks = append(bp.PlaintextChunks, b)
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
		ChunkSize:  obj.PolicyChunkSize(),
		ObjectSize: obj.ObjectSize(),

		// TODO: don't hardcode those
		MinimumBacklogDepth:   2,
		BundleRequestInterval: 5,
	}

	p.l.Debug("done; returning bundle")
	return bundle, nil
}

func (p *ContentPublisher) assignSlot(path string, chunkIdx uint64, chunkID uint64) uint64 {
	// XXX: should depend on number of slots available to cache, etc.
	return chunkIdx
}

// TODO: XXX: Since object policy is, by definition, something that the publisher can set arbitrarily on a per-object
// basis, this should be the only place that these values are hardcoded.
func (p *ContentPublisher) objectPolicy(path string) (*catalog.ObjectPolicy, error) {
	return &catalog.ObjectPolicy{
		ChunkSize: 128 * 1024,
	}, nil
}

func (p *ContentPublisher) getChunkID(obj *catalog.ObjectMetadata, chunkIdx uint64) (common.ChunkID, error) {
	data, err := obj.GetChunk(uint32(chunkIdx))
	if err != nil {
		return common.ChunkID{}, errors.Wrap(err, "failed to get chunk data to generate ID")
	}

	var id common.ChunkID
	digest := sha512.Sum384(data)
	copy(id[:], digest[0:common.ChunkIDSize])

	p.l.WithFields(logrus.Fields{
		"chunkIdx": chunkIdx,
		"chunkID":  id,
	}).Debug("generating chunk ID")

	return id, nil
}

func (p *ContentPublisher) cacheMiss(ctx context.Context, req *ccmsg.CacheMissRequest) (*ccmsg.CacheMissResponse, error) {
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
	// if req.RangeEnd <= number-of-chunks-in-object ... invalid range

	objMeta, err := p.catalog.GetMetadata(ctx, path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get metadata for object")
	}

	// Convert object policy, which is required to convert chunk range into byte range.
	pol, err := p.objectPolicy(path)
	if err != nil {
		return nil, errors.New("failed to get object policy")
	}

	resp := ccmsg.CacheMissResponse{
		Chunks: []*ccmsg.Chunk{},
	}

	// XXX: Shouldn't we be telling the cache what chunk IDs it should expect, and providing enough information for it
	// to verify that it's getting the right data (e.g. a digest)?

	// Select logical cache slot for each chunk.
	for i := req.RangeBegin; i < req.RangeEnd; i++ {
		chunkID := i // XXX: Not true!
		slotIdx := p.assignSlot(path, i, chunkID)
		chunk, err := p.catalog.ChunkSource(ctx, req, path, objMeta)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get chunk source")
		}

		// TODO: we shouldn't need to modify the chunk afterwards
		// chunk.ChunkId = ChunkID,
		chunk.SlotIdx = slotIdx

		resp.Chunks = append(resp.Chunks, chunk)
	}

	resp.Metadata = &ccmsg.ObjectMetadata{
		ObjectSize: objMeta.ObjectSize(),
		ChunkSize:  uint64(pol.ChunkSize),
	}

	return &resp, nil
}

func (p *ContentPublisher) AddEscrowToDatabase(ctx context.Context, escrow *Escrow) error {
	if err := p.AddEscrow(escrow); err != nil {
		return errors.Wrap(err, "failed to add escrow to publisher")
	}
	if err := escrow.Inner.Insert(ctx, dbtx.ExecutorFromContext(ctx), boil.Infer()); err != nil {
		return errors.Wrap(err, "failed to add escrow to database")
	}

	for _, c := range escrow.Caches {
		p.l.Info("Adding cache to database: ", c)
		err := c.Cache.Upsert(ctx, dbtx.ExecutorFromContext(ctx), true, []string{"public_key"}, boil.Whitelist("inetaddr", "port"), boil.Infer())
		if err != nil {
			return errors.Wrap(err, "failed to add cache to database")
		}

		ec := models.EscrowCache{
			EscrowID:       escrow.Inner.ID,
			CacheID:        c.Cache.ID,
			InnerMasterKey: c.InnerMasterKey,
		}
		err = ec.Upsert(ctx, dbtx.ExecutorFromContext(ctx), false, []string{"escrow_id", "cache_id"}, boil.Whitelist("inner_master_key"), boil.Infer())
		if err != nil {
			return errors.Wrap(err, "failed to link cache to escrow")
		}
	}

	return nil
}

// XXX: This is, obviously, temporary.  We should be using object IDs that are larger than 64 bits, among other
// problems.  We also must account for the fact that the object stored at a path may change (e.g. when the mtime/etag
// are updated).
func generateObjectID(path string) (common.ObjectID, error) {
	digest := sha256.Sum256([]byte(path))
	return common.BytesToObjectID(digest[0:common.ObjectIDSize])
}

// get from cache or calculate if needed the maglev-hash permutations of a cache
func (p *ContentPublisher) getCachePermutation(pubkey string, escrowSize uint) ([]uint64, error) {
	bits := math.Ceil(math.Log2(float64((escrowSize * 75))))
	// Lookup table sizes are primes, so we have approximate to actual
	if escrowSize > 100 {
		return []uint64{}, errors.New("cannot handle escrows with more than 100 caches yet (more primes needed)")
	}
	approximate := uint64(math.Exp2(bits))
	// 128 -> 131
	// 256 -> 257
	// 512 -> 521
	// 1024 -> etc
	approxToLength := map[uint64]uint64{
		128:  131,
		256:  257,
		512:  521,
		1024: 1031,
		2048: 2053,
		4096: 4099,
		8192: 8209,
	}
	length := approxToLength[approximate]

	cache, found := p.caches[pubkey]
	if !found {
		return []uint64{}, errors.New("Cannot get permutation for unknown cache")
	}
	prior, ok := cache.permutations[int(length)]
	if ok {
		return prior, nil
	}

	h64 := siphash.Hash(k0, k1, []byte(pubkey))
	// The largest permutation we might ever need is probably order (500k ) or 20
	// bits, so just use the one siphash for both mod and offset calculation.
	if length > 500000 {
		return []uint64{}, errors.New("cannot handle escrows with more than 500K caches")
	}
	offset := h64 % length
	skip := (h64>>uint(bits))%(length-1) + 1
	permutation := make([]uint64, length)
	//	permutation[i][j]←(offset+j×skip)modM
	var j uint64
	for j = 0; j < length; j++ {
		permutation[j] = (offset + j*skip) % length
	}
	// Cache the result
	// This may require a lock for memory safety from
	// here ----
	cache.permutations[int(length)] = permutation
	// to here ----
	return permutation, nil
}
