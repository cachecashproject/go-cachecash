package catalog

import (
	"context"
	"sync"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/jonboulle/clockwork"
	"github.com/sirupsen/logrus"
)

type catalog struct {
	l     *logrus.Logger
	clock clockwork.Clock

	upstream Upstream

	mu      sync.Mutex
	objects map[string]*ObjectMetadata
}

var _ ContentCatalog = (*catalog)(nil)

func NewCatalog(l *logrus.Logger, upstream Upstream) (*catalog, error) {
	return &catalog{
		l:        l,
		clock:    clockwork.NewRealClock(),
		upstream: upstream,
		objects:  make(map[string]*ObjectMetadata),
	}, nil
}

func (c *catalog) GetMetadata(ctx context.Context, path string) (*ObjectMetadata, error) {
	// XXX: This works, but is NOT a good long-term solution: it may cause a fetch of the first block of the object.
	// XXX: This should trigger a HEAD request instead of a range request
	return c.GetData(ctx, &ccmsg.ContentRequest{
		Path:       path,
		RangeBegin: 0,
		RangeEnd:   1,
	})
}

// XXX: Should return type be different?
func (c *catalog) GetData(ctx context.Context, req *ccmsg.ContentRequest) (*ObjectMetadata, error) {
	// There are several cases to consider.
	// - There may be no record of the object in the cache.
	// - There may be a record of the object, and...
	//   - ...it is valid.  It may be used immediately.
	//   - ...it has expired.  A request should be made to see if it has changed.
	//     If so, remove and replace the metadata; otherwise, change the expiry.

	// TODO: What if our requests to the upstream origin fail with retryable errors?  We should back off and retry
	// without discarding data.

	// XXX: We don't need to hold a global lock this entire time, and we absolutely shouldn't hold it while we are
	// watiing for requests in flight to resolve.
	c.mu.Lock()
	m, ok := c.objects[req.Path]
	if !ok {
		m = newObjectMetadata(c)
		c.objects[req.Path] = m
	}
	c.mu.Unlock()

	if err := m.ensureFresh(ctx, req); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *catalog) Upstream(path string) (Upstream, error) {
	return c.upstream, nil
}
