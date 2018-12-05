package catalog

import (
	"context"
	"sync"
)

// - Different objects may have different block strategies: different sizes; fixed-size vs. rolling-hash blocks, etc.

/*
type upstreamRequest struct {
	doneCh chan struct{}

	resp *http.Response
	err  error
}

func makeUpstreamRequest(path string) *upstreamRequest {
	r := &upstreamRequest{
		doneCh: make(chan struct{}),
	}
	defer close(r.doneCh)
	go r.fetch(path)
	return r
}
func (r *upstreamRequest) fetch(path string) {

	// XXX: This should become a HEAD request.
}
*/

// policy describes how an object will be divided into blocks, etc.  This is information other than what is returned
// from the upstream HTTP source.
type policy struct {
}

type catalog struct {
	upstream Upstream

	mu      sync.Mutex
	objects map[string]*objectMetadata
}

func newCatalog(upstream Upstream) (*catalog, error) {
	return &catalog{
		upstream: upstream,
		objects:  make(map[string]*objectMetadata),
	}, nil
}

func (c *catalog) getObjectPolicy(path string) (*policy, error) {
	return nil, nil
}

func (c *catalog) getObjectMetadata(ctx context.Context, path string) (*objectMetadata, error) {
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
