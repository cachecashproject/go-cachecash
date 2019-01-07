package catalog

import (
	"context"
	"crypto/aes"
	"net/http"
	"sync"
	"time"

	cachecash "github.com/kelleyk/go-cachecash"
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

	Nonempty bool // XXX: This will probably be replaced with something else (e.g. the real data members) shortly.

	RespHeader http.Header // Probably don't want to store these directly.
	RespData   []byte
	RespErr    error

	mu        sync.Mutex
	reqDoneCh chan struct{}
}

var _ cachecash.ContentObject = (*ObjectMetadata)(nil)

func newObjectMetadata(c *catalog) *ObjectMetadata {
	return &ObjectMetadata{c: c}
}

func (m *ObjectMetadata) Fresh() bool {
	if !m.Nonempty {
		return false
	}

	return true
}

const (
	defaultBlockSize = 512 * 1024 // Fixed 512 KiB block size.
)

// BlockSize returns the size of a particular data block in bytes.
// TODO: Do we really need this?
func (m *ObjectMetadata) BlockSize(dataBlockIdx uint32) (int, error) {

	return defaultBlockSize, nil
}

func (m *ObjectMetadata) GetBlock(dataBlockIdx uint32) ([]byte, error) {
	// XXX: Do better range checking/etc.  Obviously, this won't work once we are doing multiple partial fetches.
	block := m.RespData[defaultBlockSize*dataBlockIdx : defaultBlockSize*(dataBlockIdx+1)]
	m.c.l.Debugf("ObjectMetadata.GetBlock() len(rval)=%v", len(block))
	return block, nil
}

// GetCipherBlock returns an individual cipher block (aka "sub-block") of a particular data block (a protocol-level
// block).  The return value will be aes.BlockSize bytes long (16 bytes).  ciperBlockIdx is taken modulo the number
// of whole cipher blocks that exist in the data block.
func (m *ObjectMetadata) GetCipherBlock(dataBlockIdx, cipherBlockIdx uint32) ([]byte, error) {
	dataBlock, err := m.GetBlock(dataBlockIdx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get data block")
	}

	cipherBlockIdx = cipherBlockIdx % uint32(len(dataBlock)/aes.BlockSize)
	cipherBlock := dataBlock[cipherBlockIdx*aes.BlockSize : (cipherBlockIdx+1)*aes.BlockSize]
	m.c.l.Debugf("ObjectMetadata.GetCipherBlock() len(rval)=%v", len(cipherBlock))
	return cipherBlock, nil
}

// BlockCount returns the number of blocks in this object.
func (m *ObjectMetadata) BlockCount() int {
	panic("no impl")
	// return 0
}

func (m *ObjectMetadata) BlockDigest(dataBlockIdx uint32) ([]byte, error) {
	panic("no impl")
	// return nil, nil
}

// ensureFresh returns immediately if the object's metadata is already in-cache and fresh.  Otherwise, it ensures that
// exactly one request for the metadata is made.
func (m *ObjectMetadata) ensureFresh(ctx context.Context, path string) error {
	// N.B.: At this point, all goroutines will have a pointer to the same m.

	if m.Fresh() {
		return nil
	}

	// m is not fresh; either it's an empty/new metadata object or the metadata we have has expired.
	// We want exactly one upstream request to update it.

	m.mu.Lock()
	if m.reqDoneCh == nil {
		m.reqDoneCh = make(chan struct{})
		go m.fetchMetadata(ctx, path, m.reqDoneCh)
	}
	reqDoneCh := m.reqDoneCh
	m.mu.Unlock()

	// XXX: What if the request finishes before this point is reached?
	select {
	case <-reqDoneCh:
	case <-ctx.Done():
		return ctx.Err()
	}

	// XXX: TODO: We don't actually want all of the requesters to get the outcome out the upstream request!  We want it
	// to be processed once, and for the objectMetadata to be updated.  All of the other requesters need only be
	// notified once the objectMetadata struct is ready for them to use!

	// XXX: The metadata object itself may indicate that the object does not exist, etc.  Should we translate that into
	// an error here?
	return nil
}

// XXX: We take doneCh as an argument but ignore it in favor of m.reqDoneCh.
func (m *ObjectMetadata) fetchMetadata(ctx context.Context, path string, doneCh chan struct{}) {
	defer close(m.reqDoneCh)

	r, err := m.c.upstream.FetchData(ctx, path, true, 0, 0)
	if err != nil {
		m.c.l.WithError(err).Error("failed to fetch from upstream")
		return
	}

	// XXX: I'm not sure that we still want to do this.
	m.RespHeader = r.header
	m.RespData = r.data
	m.Status = r.status
	// TODO: Update last-fetched/last-attempted times based on status.
}
