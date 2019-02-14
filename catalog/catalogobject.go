package catalog

import (
	"context"
	"crypto/aes"
	"fmt"
	"math"
	"sync"
	"time"

	cachecash "github.com/cachecashproject/go-cachecash"
	"github.com/cachecashproject/go-cachecash/cachecontrol"
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

/*

- The publisher can decide how each object is split into blocks; the cache must accept whatever decision the publisher
  made.

- The publisher won't use a CacheCash upstream; caches may be told to.


Things that need to be extended here:
- Upstream may not be HTTP.  Need interface.
- Fetches may time out or return transient/permanent errors.
- Periodically, we need to revalidate the metadata (and data) we have.
- Once we know that metadata is valid, we need to fetch any necessary blocks.
  This will need the same coalescing logic.

*/

type ObjectMetadata struct {
	c *catalog

	// blockStrategy describes how the object has been split into blocks.  This is necessary to map byte positions into
	// block positions and vice versa.
	// blockStrategy ...

	Status      ObjectStatus
	LastUpdate  time.Time
	LastAttempt time.Time
	ValidUntil  *time.Time
	Immutable   bool

	HTTPLastModified *string
	HTTPEtag         *string

	mu sync.RWMutex

	// Covered by `mu`.
	policy   *ObjectPolicy
	metadata *ccmsg.ObjectMetadata
	blocks   [][]byte
}

// ObjectPolicy contains publisher-determined metadata such as block size.  This is distinct from ccmsg.ObjectMetadata,
// which contains metadata cached from the upstream.
type ObjectPolicy struct {
	BlockSize            int
	DefaultCacheDuration time.Duration
}

func (policy *ObjectPolicy) ChunkIntoBlocks(buf []byte) [][]byte {
	blockSize := policy.BlockSize

	var block []byte
	blockCount := BlockCount(uint64(len(buf)), blockSize)
	blocks := make([][]byte, 0, blockCount)

	for len(buf) >= blockSize {
		block, buf = buf[:blockSize], buf[blockSize:]
		blocks = append(blocks, block)
	}

	// doing this afterwards so we don't need to branch inside the loop
	if len(buf) > 0 {
		blocks = append(blocks, buf[:len(buf)])
	}

	return blocks
}

var _ cachecash.ContentObject = (*ObjectMetadata)(nil)

func newObjectMetadata(c *catalog) *ObjectMetadata {
	return &ObjectMetadata{
		c:      c,
		blocks: make([][]byte, 0),
		policy: &ObjectPolicy{
			BlockSize:            128 * 1024,      // Fixed 128 KiB block size.  XXX: Don't hardwire this!
			DefaultCacheDuration: 5 * time.Minute, // XXX: Don't hardwire this!
		},
	}
}

// XXX: Is this a concurrency issue?
func (m *ObjectMetadata) Metadata() *ccmsg.ObjectMetadata {
	return m.metadata
}

func (m *ObjectMetadata) Fresh() bool {
	if m.Immutable {
		m.c.l.Debugln("Fresh() - Object is immutable")
		return true
	}

	if m.ValidUntil == nil {
		m.c.l.Debugln("Fresh() - ValidUntil hasn't been populated, force revalidation")
		return false
	}

	if m.c.clock.Now().Before(*m.ValidUntil) {
		m.c.l.Traceln("Fresh() - ObjectMetadata is still fresh, not revalidating")
		return true
	}

	m.c.l.Debugln("Fresh() - ObjectMetadata is stale, revalidating now")
	return false
}

func (m *ObjectMetadata) PolicyBlockSize() uint64 {
	return uint64(m.policy.BlockSize)
}

// BlockSize returns the size of a particular data block in bytes.
// N.B.: It's important that this return the actual size of the indicated block; otherwise, if we are generating a
//   puzzle that includes the last block in an object (which may be shorter than PolicyBlockSize() would suggest)
//   the colocation puzzle code may generate unsolvable puzzles (e.g. when the initial offset is chosen to be past
//   the end of the actual block).
func (m *ObjectMetadata) BlockSize(dataBlockIdx uint32) (int, error) {
	// XXX: More integer-typecasting nonsense.  Straighten this out!
	s := int(m.metadata.ObjectSize) - (int(m.policy.BlockSize) * int(dataBlockIdx))
	if s > m.policy.BlockSize {
		s = m.policy.BlockSize
	}
	return s, nil
}

func (m *ObjectMetadata) GetBlock(dataBlockIdx uint32) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.getBlock(dataBlockIdx)
}

func (m *ObjectMetadata) getBlock(dataBlockIdx uint32) ([]byte, error) {
	if int(dataBlockIdx) >= len(m.blocks) || m.blocks[dataBlockIdx] == nil {
		return nil, errors.New("block not in cache")
	}

	return m.blocks[dataBlockIdx], nil
}

func (m *ObjectMetadata) GetCipherBlock(dataBlockIdx, cipherBlockIdx uint32) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.getCipherBlock(dataBlockIdx, cipherBlockIdx)
}

// GetCipherBlock returns an individual cipher block (aka "sub-block") of a particular data block (a protocol-level
// block).  The return value will be aes.BlockSize bytes long (16 bytes).  ciperBlockIdx is taken modulo the number
// of whole cipher blocks that exist in the data block.
func (m *ObjectMetadata) getCipherBlock(dataBlockIdx, cipherBlockIdx uint32) ([]byte, error) {
	dataBlock, err := m.getBlock(dataBlockIdx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get data block")
	}

	cipherBlockIdx = cipherBlockIdx % uint32(len(dataBlock)/aes.BlockSize)
	cipherBlock := dataBlock[cipherBlockIdx*aes.BlockSize : (cipherBlockIdx+1)*aes.BlockSize]
	m.c.l.Debugf("ObjectMetadata.GetCipherBlock() len(rval)=%v", len(cipherBlock))
	return cipherBlock, nil
}

func BlockCount(size uint64, blockSize int) int {
	return int(math.Ceil(float64(size) / float64(blockSize)))
}

// BlockCount returns the number of blocks in this object.
// XXX: This is a problem; m.metadata may be nil if we don't know anything about the object.
func (m *ObjectMetadata) BlockCount() int {
	return BlockCount(m.metadata.ObjectSize, m.policy.BlockSize)
}

func (m *ObjectMetadata) ObjectSize() uint64 {
	return m.metadata.ObjectSize
}

func (m *ObjectMetadata) BlockDigest(dataBlockIdx uint32) ([]byte, error) {
	panic("no impl")
	// return nil, nil
}

// Converts a byte range to a block range.  An end value of 0, which indicates that the range continues to the end of
// the object, converts to a 0.
func (m *ObjectMetadata) blockRange(rangeBegin, rangeEnd uint64) (uint64, uint64) {
	blockRangeBegin := rangeBegin / uint64(m.policy.BlockSize)

	var blockRangeEnd uint64
	if rangeEnd != 0 {
		blockRangeEnd = uint64(math.Ceil(float64(rangeEnd) / float64(m.policy.BlockSize)))
	}

	return blockRangeBegin, blockRangeEnd
}

// ensureFresh ensures that the object's metadata is valid (i.e. has not changed/expired), and that the block(s)
// described by req are present in cache.
func (m *ObjectMetadata) ensureFresh(ctx context.Context, req *ccmsg.ContentRequest) error {
	m.mu.Lock()
	fresh := m.Fresh()
	m.mu.Unlock()

	m.c.l.WithFields(log.Fields{
		"path": req.Path,
	}).Debugf("ensureFresh for byte range [%v, %v] -> fresh=%v", req.RangeBegin, req.RangeEnd, fresh)
	if fresh {
		return nil
	}

	doneCh := make(chan struct{})
	go m.fetchData(ctx, req, doneCh)

	select {
	case <-doneCh:
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func (m *ObjectMetadata) fetchData(ctx context.Context, req *ccmsg.ContentRequest, doneCh chan struct{}) {
	defer close(doneCh)

	log := m.c.l.WithFields(log.Fields{
		"path": req.Path,
	})

	// XXX: this is set to 0, 0 to fetch the whole file. This function might be used by the cache in the future, so range support is still needed
	r, err := m.c.upstream.FetchData(ctx, req.Path, m, 0, 0)
	if err != nil {
		log.WithError(err).Error("failed to fetch from upstream")
		// XXX: Should set m.metadata.Status, right?  Why isn't this covered by the test suite?
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	blockRangeBegin, blockRangeEnd := m.blockRange(req.RangeBegin, req.RangeEnd)
	log.Debugf("fetchData for requested blockRange [%v, %v]", blockRangeBegin, blockRangeEnd)

	// Populate metadata.
	if m.metadata == nil {
		m.metadata = &ccmsg.ObjectMetadata{}
	}

	size, err := r.ObjectSize()
	if err != nil {
		panic(fmt.Sprintf("error parsing metadata: %v", err))
	}
	m.metadata.ObjectSize = uint64(size)
	log.Debugf("fetchData populates metadata; ObjectSize=%v", m.metadata.ObjectSize)

	// XXX: don't expose the status, this is handled in here
	log.Debugf("fetchData - r.status=%v", r.status)
	m.Status = r.status

	switch r.status {
	case StatusOK:
		log.Debugln("fetchData - got response, slicing into blocks")
		m.blocks = m.policy.ChunkIntoBlocks(r.data)
		log.Debugf("fetchData - populated cache with %v blocks", len(m.blocks))

	case StatusNotModified:
		log.Debugln("fetchData - upstream wasn't modified, our data is still fresh")

	default:
		log.Errorf("fetchData - received unexpected http status: %v", r.status)
		return
	}

	// set freshness values accordingly
	cacheControl := r.header.Get("Cache-Control")
	if cacheControl != "" {
		cc, err := cachecontrol.Parse(cacheControl)
		if err == nil {
			if cc.MaxAge != nil {
				validUntil := m.c.clock.Now().Add(*cc.MaxAge)
				m.ValidUntil = &validUntil
			}
			m.Immutable = cc.Immutable
		}
	}

	if m.ValidUntil == nil && !m.Immutable {
		log.Warnf("fetchData - no cache control rules found, using default: %s", m.policy.DefaultCacheDuration)
		validUntil := m.c.clock.Now().Add(m.policy.DefaultCacheDuration)
		m.ValidUntil = &validUntil
	}
}
