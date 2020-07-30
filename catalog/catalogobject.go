package catalog

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/cachecashproject/go-cachecash/cachecontrol"
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

/*

- The publisher can decide how each object is split into chunks; the cache must accept whatever decision the publisher
  made.

- The publisher won't use a CacheCash upstream; caches may be told to.


Things that need to be extended here:
- Upstream may not be HTTP.  Need interface.
- Fetches may time out or return transient/permanent errors.
- Periodically, we need to revalidate the metadata (and data) we have.
- Once we know that metadata is valid, we need to fetch any necessary chunks.
  This will need the same coalescing logic.

*/

type ObjectMetadata struct {
	c *catalog

	// chunkStrategy describes how the object has been split into chunks.  This is necessary to map byte positions into
	// chunk positions and vice versa.
	// chunkStrategy ...

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
	chunks   [][]byte
}

// ObjectPolicy contains publisher-determined metadata such as chunk size.  This is distinct from ccmsg.ObjectMetadata,
// which contains metadata cached from the upstream.
type ObjectPolicy struct {
	ChunkSize            int
	DefaultCacheDuration time.Duration
}

func (policy *ObjectPolicy) SplitIntoChunks(buf []byte) [][]byte {
	chunkSize := policy.ChunkSize

	var chunk []byte
	chunkCount := ChunkCount(uint64(len(buf)), chunkSize)
	chunks := make([][]byte, 0, chunkCount)

	for len(buf) >= chunkSize {
		chunk, buf = buf[:chunkSize], buf[chunkSize:]
		chunks = append(chunks, chunk)
	}

	// doing this afterwards so we don't need to branch inside the loop
	if len(buf) > 0 {
		chunks = append(chunks, buf)
	}

	return chunks
}

func newObjectMetadata(c *catalog) *ObjectMetadata {
	return &ObjectMetadata{
		c:      c,
		chunks: make([][]byte, 0),
		policy: &ObjectPolicy{
			ChunkSize:            128 * 1024,      // Fixed 128 KiB chunk size.  XXX: Don't hardwire this!
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

func (m *ObjectMetadata) PolicyChunkSize() uint64 {
	return uint64(m.policy.ChunkSize)
}

// ChunkSize returns the size of a particular chunk in bytes.
// N.B.: It's important that this return the actual size of the indicated chunk; otherwise, if we are generating a
//   puzzle that includes the last chunk in an object (which may be shorter than PolicyChunkSize() would suggest)
//   the colocation puzzle code may generate unsolvable puzzles (e.g. when the initial offset is chosen to be past
//   the end of the actual chunk).
func (m *ObjectMetadata) ChunkSize(chunkIdx uint32) (int, error) {
	// XXX: More integer-typecasting nonsense.  Straighten this out!
	s := int(m.metadata.ObjectSize) - (int(m.policy.ChunkSize) * int(chunkIdx))
	if s > m.policy.ChunkSize {
		s = m.policy.ChunkSize
	}
	return s, nil
}

func (m *ObjectMetadata) GetChunk(chunkIdx uint32) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.getChunk(chunkIdx)
}

func (m *ObjectMetadata) getChunk(chunkIdx uint32) ([]byte, error) {
	if int(chunkIdx) >= len(m.chunks) || m.chunks[chunkIdx] == nil {
		return nil, errors.New("chunk not in cache")
	}

	return m.chunks[chunkIdx], nil
}

func (m *ObjectMetadata) ChunkRange(rangeBegin uint64, rangeEnd uint64) ([][]byte, error) {
	if int(rangeEnd) > len(m.chunks) {
		return nil, errors.New("chunk end out of range")
	} else if rangeEnd == 0 {
		rangeEnd = uint64(len(m.chunks))
	}
	return m.chunks[rangeBegin:rangeEnd], nil
}

func ChunkCount(size uint64, chunkSize int) int {
	return int(math.Ceil(float64(size) / float64(chunkSize)))
}

// ChunkCount returns the number of chunks in this object.
// XXX: This is a problem; m.metadata may be nil if we don't know anything about the object.
func (m *ObjectMetadata) ChunkCount() int {
	return ChunkCount(m.metadata.ObjectSize, m.policy.ChunkSize)
}

func (m *ObjectMetadata) ObjectSize() uint64 {
	return m.metadata.ObjectSize
}

// Converts a byte range to a chunk range.  An end value of 0, which indicates that the range continues to the end of
// the object, converts to a 0.
func (m *ObjectMetadata) chunkRange(rangeBegin, rangeEnd uint64) (uint64, uint64) {
	chunkRangeBegin := rangeBegin / uint64(m.policy.ChunkSize)

	var chunkRangeEnd uint64
	if rangeEnd != 0 {
		chunkRangeEnd = uint64(math.Ceil(float64(rangeEnd) / float64(m.policy.ChunkSize)))
	}

	return chunkRangeBegin, chunkRangeEnd
}

// ensureFresh ensures that the object's metadata is valid (i.e. has not changed/expired), and that the chunk(s)
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

	doneCh := make(chan error)
	go m.fetchData(ctx, req, doneCh)

	select {
	case err := <-doneCh:
		if err != nil {
			log.Error(err)
		}
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *ObjectMetadata) fetchData(ctx context.Context, req *ccmsg.ContentRequest, doneCh chan error) {
	defer close(doneCh)

	log := m.c.l.WithFields(log.Fields{
		"path": req.Path,
	})

	// XXX: this is set to 0, 0 to fetch the whole file. This function might be used by the cache in the future, so range support is still needed
	r, err := m.c.upstream.FetchData(ctx, req.Path, m, 0, 0)
	if err != nil {
		// XXX: Should set m.metadata.Status, right?  Why isn't this covered by the test suite?
		doneCh <- errors.Wrap(err, "failed to fetch from upstream")
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	chunkRangeBegin, chunkRangeEnd := m.chunkRange(req.RangeBegin, req.RangeEnd)
	log.Debugf("fetchData for requested chunkRange [%v, %v]", chunkRangeBegin, chunkRangeEnd)

	// Populate metadata.
	if m.metadata == nil {
		m.metadata = &ccmsg.ObjectMetadata{}
	}

	if r == nil {
		log.Debug("FetchData returned nil")
		doneCh <- errors.Wrap(err, "request response is nil")
		return
	}

	// XXX: don't expose the status, this is handled in here
	log.Debugf("fetchData - r.status=%v", r.status)
	m.Status = r.status

	switch r.status {
	case StatusOK:
		log.Debugln("fetchData - got response, slicing into chunks")
		m.chunks = m.policy.SplitIntoChunks(r.data)
		log.Debugf("fetchData - populated cache with %v chunks", len(m.chunks))

		size, err := r.ObjectSize()
		if err != nil {
			doneCh <- errors.Wrap(err, "error parsing metadata")
			return
		}
		m.metadata.ObjectSize = uint64(size)

	case StatusNotModified:
		log.Debugln("fetchData - upstream wasn't modified, our data is still fresh")

	default:
		// log.Errorf("fetchData - received unexpected http status: %v", r.status)
		doneCh <- fmt.Errorf("Received unexpected http status: %v", r.status)
		return
	}

	log.Debugf("fetchData populates metadata; ObjectSize=%v", m.metadata.ObjectSize)

	// set freshness values accordingly
	m.ValidUntil = nil

	cacheControl := r.header.Get("Cache-Control")
	if cacheControl != "" {
		cc := cachecontrol.Parse(cacheControl)

		if cc.MaxAge != nil {
			validUntil := m.c.clock.Now().Add(*cc.MaxAge)
			m.ValidUntil = &validUntil
		}
		m.Immutable = cc.Immutable
	}

	if m.ValidUntil == nil && !m.Immutable {
		log.Warnf("fetchData - no cache control rules found, using default: %s", m.policy.DefaultCacheDuration)
		validUntil := m.c.clock.Now().Add(m.policy.DefaultCacheDuration)
		m.ValidUntil = &validUntil
	}
}
