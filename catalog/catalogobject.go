package catalog

import (
	"context"
	"crypto/aes"
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	cachecash "github.com/kelleyk/go-cachecash"
	"github.com/kelleyk/go-cachecash/ccmsg"
	"github.com/pkg/errors"
)

/*

- The provider can decide how each object is split into blocks; the cache must accept whatever decision the provider
  made.

- The provider won't use a CacheCash upstream; caches may be told to.


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

	mu sync.RWMutex

	// Covered by `mu`.
	policy   *ObjectPolicy
	metadata *ccmsg.ObjectMetadata
	blocks   [][]byte
}

// ObjectPolicy contains provider-determined metadata such as block size.  This is distinct from ccmsg.ObjectMetadata,
// which contains metadata cached from the upstream.
type ObjectPolicy struct {
	BlockSize int
}

var _ cachecash.ContentObject = (*ObjectMetadata)(nil)

func newObjectMetadata(c *catalog) *ObjectMetadata {
	return &ObjectMetadata{
		c:      c,
		blocks: make([][]byte, 0),
		policy: &ObjectPolicy{
			BlockSize: 512 * 1024, // Fixed 512 KiB block size.  XXX: Don't hardwire this!
		},
	}
}

// XXX: Needs real implementation.
func (m *ObjectMetadata) Fresh() bool {
	return true
}

// BlockSize returns the size of a particular data block in bytes.
// TODO: Do we really need this?
func (m *ObjectMetadata) BlockSize(dataBlockIdx uint32) (int, error) {
	return m.policy.BlockSize, nil
}

func (m *ObjectMetadata) GetBlock(dataBlockIdx uint32) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.getBlock(dataBlockIdx)
}

func (m *ObjectMetadata) getBlock(dataBlockIdx uint32) ([]byte, error) {
	if int(dataBlockIdx) > len(m.blocks) || m.blocks[dataBlockIdx] == nil {
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

// BlockCount returns the number of blocks in this object.
// XXX: This is a problem; m.metadata may be nil if we don't know anything about the object.
func (m *ObjectMetadata) BlockCount() int {
	return int(math.Ceil(float64(m.metadata.ObjectSize) / float64(m.policy.BlockSize)))
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

// Requires that the caller hold a read lock on `m.mu`.
func (m *ObjectMetadata) rangeInCache(rangeBegin, rangeEnd uint64) bool {
	blockRangeBegin, blockRangeEnd := m.blockRange(rangeBegin, rangeEnd)

	if blockRangeEnd == 0 {
		if m.metadata == nil || m.metadata.ObjectSize == 0 {
			// We don't know how long the object is yet, so we must need to make an upstream request.
			m.c.l.Debugf("rangeInCache() - unfilled metadata")
			return false
		}
		blockRangeEnd = uint64(m.BlockCount())
	}

	if int(blockRangeEnd) > len(m.blocks) {
		m.c.l.Debugf("rangeInCache() - cache too short to cover request")
		return false
	}
	for i := blockRangeBegin; i < blockRangeEnd; i++ {
		if m.blocks[i] == nil {
			m.c.l.Debugf("rangeInCache() - cache missing block %v", i)
			return false
		}
	}
	return true
}

// ensureFresh ensures that the object's metadata is valid (i.e. has not changed/expired), and that the block(s)
// described by req are present in cache.
func (m *ObjectMetadata) ensureFresh(ctx context.Context, req *ccmsg.ContentRequest) error {
	m.mu.Lock()
	covered := m.rangeInCache(req.RangeBegin, req.RangeEnd)
	fresh := m.Fresh()
	m.mu.Unlock()

	m.c.l.Debugf("ensureFresh for byte range [%v, %v) -> covered=%v fresh=%v", req.RangeBegin, req.RangeEnd, covered, fresh)
	if covered && fresh {
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

	r, err := m.c.upstream.FetchData(ctx, req.Path, true, uint(req.RangeBegin), uint(req.RangeEnd))
	if err != nil {
		m.c.l.WithError(err).Error("failed to fetch from upstream")
		// XXX: Should set m.metadata.Status, right?  Why isn't this covered by the test suite?
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	blockRangeBegin, blockRangeEnd := m.blockRange(req.RangeBegin, req.RangeEnd)
	m.c.l.Debugf("fetchData for requested blockRange [%v, %v)", blockRangeBegin, blockRangeEnd)

	// Populate metadata.
	if m.metadata == nil {
		m.metadata = &ccmsg.ObjectMetadata{}
	}
	// XXX: What if header is absent?
	size, err := strconv.Atoi(r.header.Get("Content-Length"))
	if err != nil {
		panic(fmt.Sprintf("error parsing metadata: %v", err))
	}
	m.metadata.ObjectSize = uint64(size)
	m.c.l.Debugf("fetchData populates metadata; ObjectSize=%v", m.metadata.ObjectSize)

	if blockRangeEnd == 0 {
		blockRangeEnd = uint64(m.BlockCount())
	}
	m.c.l.Debugf("fetchData using blockRange [%v, %v)", blockRangeBegin, blockRangeEnd)

	// XXX: Error responses from upstream shouldn't make us immediately drop the entire object on the floor.
	m.c.l.Debugf("fetchData - r.status=%v (%d)", r.status, r.status)
	m.Status = r.status
	if m.Status != StatusOK {
		return
	}

	// XXX: Handle case where we need to invalidate the data that's already in the cache.

	// Ensure len(m.blocks) >= blockRangeEnd.
	if len(m.blocks) <= int(blockRangeEnd) {
		m.blocks = append(m.blocks, make([][]byte, int(blockRangeEnd)-len(m.blocks))...)
	}
	m.c.l.Debugf("fetchData: after extend, len(m.blocks)=%v", len(m.blocks))

	// XXX: Should not assume that response range matches request range.
	blockSize := m.policy.BlockSize // XXX:
	for i := 0; i < int(blockRangeEnd-blockRangeBegin); i++ {
		m.c.l.Debugf("inserting object block into cache: %v", i+int(blockRangeBegin))

		blockEnd := (i + 1) * blockSize
		if blockEnd > len(r.data) {
			blockEnd = len(r.data)
		}
		buf := r.data[i*blockSize : blockEnd]

		m.blocks[i+int(blockRangeBegin)] = buf
	}
}
