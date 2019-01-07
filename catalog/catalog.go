package catalog

import (
	"context"
	"sync"

	"github.com/kelleyk/go-cachecash/ccmsg"
	"github.com/sirupsen/logrus"
)

type catalog struct {
	l *logrus.Logger

	upstream Upstream

	mu      sync.Mutex
	objects map[string]*ObjectMetadata
}

var _ ContentCatalog = (*catalog)(nil)

func NewCatalog(l *logrus.Logger, upstream Upstream) (*catalog, error) {
	return &catalog{
		l:        l,
		upstream: upstream,
		objects:  make(map[string]*ObjectMetadata),
	}, nil
}

func (c *catalog) GetObjectMetadata(ctx context.Context, path string) (*ObjectMetadata, error) {
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
	m, ok := c.objects[path]
	if !ok {
		m = newObjectMetadata(c)
		c.objects[path] = m
	}
	c.mu.Unlock()

	if err := m.ensureFresh(ctx, path); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *catalog) CacheMiss(path string, rangeBegin, rangeEnd uint64) (*ccmsg.CacheMissResponse, error) {
	return c.upstream.CacheMiss(path, rangeBegin, rangeEnd)
}
