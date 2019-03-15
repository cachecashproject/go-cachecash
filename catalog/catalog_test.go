package catalog

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/testutil"
	"github.com/jonboulle/clockwork"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	dataSeed  = 0xDEADBEEF
	blockSize = 128 * 1024 // in bytes; used to generate ObjectPolicy structs
)

type CatalogTestSuite struct {
	suite.Suite

	l     *logrus.Logger
	ts    *httptest.Server
	cat   *catalog
	clock clockwork.FakeClock

	upstreamResponseDelay time.Duration
	blockSize             int
	objectData            []byte
	policy                ObjectPolicy

	muMetrics          sync.Mutex
	upstreamRequestQty int
}

func TestCatalogTestSuite(t *testing.T) {
	suite.Run(t, new(CatalogTestSuite))
}

func (suite *CatalogTestSuite) SetupTest() {
	t := suite.T()

	suite.l = logrus.New()
	suite.l.SetLevel(logrus.DebugLevel)
	suite.clock = clockwork.NewFakeClock()

	suite.objectData = testutil.RandBytesFromSource(rand.NewSource(dataSeed), 4*blockSize)
	suite.upstreamRequestQty = 0
	suite.blockSize = 128 * 1024 // TODO: Ensure this is used everywhere.
	suite.policy = ObjectPolicy{
		BlockSize: suite.blockSize,
	}

	suite.ts = httptest.NewServer(http.HandlerFunc(suite.handleUpstreamRequest))

	upstream, err := NewHTTPUpstream(suite.l, suite.ts.URL, 5*time.Minute)
	if err != nil {
		t.Fatalf("failed to create HTTP upstream: %v", err)
	}

	suite.cat, err = NewCatalog(suite.l, upstream)
	suite.cat.clock = suite.clock
	if err != nil {
		t.Fatalf("failed to create catalog: %v", err)
	}
}

func (suite *CatalogTestSuite) TearDownTest() {
	suite.ts.Close()
}

// TODO: Will need to be extended when we want to test HEAD requests and/or range requests that produce 206 responses.
func (suite *CatalogTestSuite) handleUpstreamRequest(w http.ResponseWriter, r *http.Request) {
	t := suite.T()

	suite.muMetrics.Lock()
	suite.upstreamRequestQty++
	suite.muMetrics.Unlock()

	// This needs to be an actual sleep because of TestUpstreamTimeout
	time.Sleep(suite.upstreamResponseDelay)

	switch r.URL.Path {
	// This path can be cached for 1h
	case "/foo/bar":
		w.Header().Add("Content-Length", fmt.Sprintf("%v", len(suite.objectData)))

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(suite.objectData); err != nil {
			t.Fatalf("failed to write response in HTTP handler: %v", err)
		}

	// This path is immutable and cached forever
	case "/forever":
		w.Header().Add("Cache-Control", "public, immutable")

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(suite.objectData); err != nil {
			t.Fatalf("failed to write response in HTTP handler: %v", err)
		}

	// This path is cachable for 1sec, a revalidation returns a 304
	case "/renew/1sec":
		w.Header().Add("Cache-Control", "public,max-age=1")
		w.Header().Add("Last-Modified", "Wed, 21 Oct 2015 07:28:00 GMT")
		w.Header().Add("ETag", "asdf")

		if r.Header.Get("If-None-Match") == "asdf" {
			w.WriteHeader(http.StatusNotModified)
		} else {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write(suite.objectData); err != nil {
				t.Fatalf("failed to write response in HTTP handler: %v", err)
			}
		}

	// This path is cachable for 1sec, a revalidation returns a 200
	case "/update/1sec":
		w.Header().Add("Cache-Control", "public,max-age=1")
		w.Header().Add("Last-Modified", "Wed, 21 Oct 2015 07:28:00 GMT")
		w.Header().Add("ETag", "asdf")

		// always changed/200
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(suite.objectData); err != nil {
			t.Fatalf("failed to write response in HTTP handler: %v", err)
		}

	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (suite *CatalogTestSuite) TestSimple() {
	t := suite.T()
	cat := suite.cat

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	m, err := cat.GetData(ctx, &ccmsg.ContentRequest{Path: "/foo/bar"})
	assert.Nil(t, err)
	assert.NotNil(t, m)

	assert.Equal(t, uint64(len(suite.objectData)), m.ObjectSize())
}

/*
XXX: Removed coalescing support.
// Tests that when a downstream request is received when an upstream request is already in flight, a new upstream
// request is not made; the result of the single upstream request is provided to all downstream requests.
func (suite *CatalogTestSuite) TestCoalescing() {
	t := suite.T()
	cat := suite.cat

	suite.upstreamResponseDelay = 1 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	mm := make([]*ObjectMetadata, 2)
	for i := 0; i < len(mm); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			m, err := cat.GetData(ctx, &ccmsg.ContentRequest{Path: "/foo/bar"})
			assert.Nil(t, err)
			assert.NotNil(t, m)
			// assert.Nil(t, m.RespErr)

			// m.mu.Lock()
			assert.Equal(t, StatusOK, m.Status)
			// m.mu.Unlock()

			mm[i] = m
		}(i)
	}
	wg.Wait()

	// This should indicate a Content-Length of 14 ("Hello, client\n").
	// TODO: Replace with an assertion once the test server is serving actual data.
	for i := 0; i < len(mm); i++ {
		fmt.Printf("object metadata %v: %v\n", i, mm[i])
	}

	// Due to coalescing, only a single request should be sent upstream.
	assert.Equal(t, 1, suite.upstreamRequestQty)
}
*/

// Once we have a valid cache entry, receiving another downstream request should not cause us to make an upstream
// request.
func (suite *CatalogTestSuite) TestCacheValid() {
	t := suite.T()
	cat := suite.cat

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	for i := 0; i < 2; i++ {
		m, err := cat.GetData(ctx, &ccmsg.ContentRequest{Path: "/foo/bar"})
		assert.Nil(t, err)
		assert.NotNil(t, m)
		// assert.Nil(t, m.RespErr)
		assert.Equal(t, StatusOK, m.Status, "unexpected response status")

		assert.Equal(t, uint64(len(suite.objectData)), m.ObjectSize())
	}

	// Due to caching, only a single request should be sent upstream.
	assert.Equal(t, 1, suite.upstreamRequestQty, "request for cached data should not create another upstream request")
}

func (suite *CatalogTestSuite) TestCacheImmutable() {
	t := suite.T()
	cat := suite.cat

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	for i := 0; i < 2; i++ {
		m, err := cat.GetData(ctx, &ccmsg.ContentRequest{Path: "/forever"})
		assert.Nil(t, err)
		assert.NotNil(t, m)
		assert.Equal(t, StatusOK, m.Status, "unexpected response status")
		assert.Equal(t, uint64(len(suite.objectData)), m.ObjectSize())

		// Delay for a year
		suite.clock.Advance(365 * 24 * time.Hour)
	}

	// Due to caching, only a single request should be sent upstream.
	assert.Equal(t, 1, suite.upstreamRequestQty, "request for cached data should not create another upstream request")
}

func (suite *CatalogTestSuite) TestCacheRevalidate() {
	t := suite.T()
	cat := suite.cat

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First request should populate cache
	m, err := cat.GetData(ctx, &ccmsg.ContentRequest{Path: "/renew/1sec"})
	assert.Nil(t, err)
	assert.NotNil(t, m)

	assert.Equal(t, StatusOK, m.Status, "unexpected response status")
	assert.Equal(t, uint64(len(suite.objectData)), m.ObjectSize())

	// Delay until cache is stale
	suite.clock.Advance(2 * time.Second)

	// Second request should be a cache miss and trigger a revalidation
	m, err = cat.GetData(ctx, &ccmsg.ContentRequest{Path: "/renew/1sec"})
	assert.Nil(t, err)
	assert.NotNil(t, m)

	assert.Equal(t, StatusNotModified, m.Status, "unexpected response status")
	assert.Equal(t, uint64(len(suite.objectData)), m.ObjectSize())

	// Due to caching, only a single request should be sent upstream.
	assert.Equal(t, 2, suite.upstreamRequestQty, "request for cached data should not create another upstream request")
}

func (suite *CatalogTestSuite) TestCacheUpdate() {
	t := suite.T()
	cat := suite.cat

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First request to populate cache
	m, err := cat.GetData(ctx, &ccmsg.ContentRequest{Path: "/update/1sec"})
	assert.Nil(t, err)
	assert.NotNil(t, m)

	assert.Equal(t, StatusOK, m.Status, "unexpected response status")
	assert.Equal(t, uint64(len(suite.objectData)), m.ObjectSize())

	// Delay until cache is stale
	suite.clock.Advance(2 * time.Second)

	// Second request should update cache
	m, err = cat.GetData(ctx, &ccmsg.ContentRequest{Path: "/update/1sec"})
	assert.Nil(t, err)
	assert.NotNil(t, m)

	assert.Equal(t, StatusOK, m.Status, "unexpected response status")
	assert.Equal(t, uint64(len(suite.objectData)), m.ObjectSize())

	// Due to caching, only a single request should be sent upstream.
	assert.Equal(t, 2, suite.upstreamRequestQty, "request for cached data should not create another upstream request")
}

func (suite *CatalogTestSuite) TestNotFound() {
	t := suite.T()
	cat := suite.cat

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	m, err := cat.GetData(ctx, &ccmsg.ContentRequest{Path: "/bogus"})
	assert.NotNil(t, err)
	assert.Nil(t, m)
	// assert.Nil(t, m.RespErr)
	// assert.Equal(t, StatusNotFound, m.Status)
}

func (suite *CatalogTestSuite) TestUpstreamUnreachable() {
	t := suite.T()

	/// Deliberately pick a port we know nothing will be listening on.
	ts := httptest.NewServer(http.HandlerFunc(suite.handleUpstreamRequest))
	ts.Close()

	upstream, err := NewHTTPUpstream(suite.l, ts.URL, 5*time.Minute)
	if err != nil {
		t.Fatalf("failed to create HTTP upstream: %v", err)
	}

	cat, err := NewCatalog(suite.l, upstream)
	if err != nil {
		t.Fatalf("failed to create catalog: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	m, err := cat.GetData(ctx, &ccmsg.ContentRequest{Path: "/foo/bar"})
	assert.NotNil(t, err)
	assert.Nil(t, m)

	// XXX: Test that the object metadata contains an error response.
}

// TestUpstreamTimeout covers what happens if the upstream does not respond before our request times out.
func (suite *CatalogTestSuite) TestUpstreamTimeout() {
	t := suite.T()
	cat := suite.cat

	suite.upstreamResponseDelay = 1 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	m, err := cat.GetData(ctx, &ccmsg.ContentRequest{Path: "/foo/bar"})
	assert.Equal(t, context.DeadlineExceeded, err)
	assert.Nil(t, m)
}

// TODO:
// - Test that metadata is populated after a successful request.

// TestCacheExpired
//  - case: object has not changed; revalidated
//  - case: object has changed; not revalidated
// Test that coalescing, caching, etc. work with error responses (e.g. 404s) too.

// Test sub-object granularity:
//  - metadata-only requests should generate HEAD requests to upstream
//  - request range for [a,b] with...
//     - nothing in cache
//     - [a,b] in cache
//     - cache contains data overlapping a or b but not both; one subrange request is generated
//     - cache contains data from within [a,b] but not overlapping [a,b]; two subrange requests are generated
// Also, behavior of above when error(s) are generated

// XXX: Should these BlockSource() tests actually be tests of ContentPublisher.CacheMiss? A lot of the logic here, where
// we turn the object ID into a path and then use that to look up metadata and policy, duplicates actual logic in that
// function.
func (suite *CatalogTestSuite) testBlockSource(req *ccmsg.CacheMissRequest, blockSource BlockSource, expectedSrc interface{}) {
	t := suite.T()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	path := "/foo/bar"

	_, err := suite.cat.GetData(ctx, &ccmsg.ContentRequest{Path: path})
	if err != nil {
		t.Fatalf("failed to pull object into catalog: %v", err)
	}

	// XXX: Duplicates logic in ContentPublisher.CacheMiss.
	objMeta, err := suite.cat.GetMetadata(ctx, path)
	if err != nil {
		t.Fatalf("failed to get metadata for object: %v", err)
	}

	suite.cat.blockSource = blockSource
	chunk, err := suite.cat.BlockSource(ctx, req, path, objMeta)

	assert.Nil(t, err)
	assert.NotNil(t, chunk)

	// XXX: TODO: These fields are not yet populated.
	// // Test that slot_idx and block_id contain enough elements to cover the entire object.
	// expectedBlockCount := int(math.Ceil(float64(len(suite.objectData)) / float64(suite.blockSize)))
	// assert.Equal(t, expectedBlockCount, len(resp.BlockId))
	// assert.Equal(t, expectedBlockCount, len(resp.SlotIdx))

	// XXX: TODO: Test fails; metadata is not populated.
	// // Test that object metadata was correctly passed along.
	// // TODO: Other fields, such as etag/last_modified, are not yet implemented.
	// assert.NotNil(t, resp.Metadata)
	// assert.Equal(t, uint64(len(suite.objectData)), resp.Metadata.ObjectSize)
	// assert.Equal(t, uint64(suite.blockSize), resp.Metadata.BlockSize)

	switch blockSource {
	case BlockSourceInline:
		bsrc, ok := chunk.Source.(*ccmsg.Chunk_Inline)
		if !ok {
			t.Fatalf("unexpected CacheMissResponse source type")
		}

		expectedSrc, ok := expectedSrc.(*ccmsg.BlockSourceInline)
		if !ok {
			t.Fatalf("failed to cast expected blocksource result")
		}

		assert.Equal(t, expectedSrc.Block, bsrc.Inline.Block)
	case BlockSourceHTTP:
		bsrc, ok := chunk.Source.(*ccmsg.Chunk_Http)
		if !ok {
			t.Fatalf("unexpected CacheMissResponse source type")
		}

		expectedSrc, ok := expectedSrc.(*ccmsg.BlockSourceHTTP)
		if !ok {
			t.Fatalf("failed to cast expected blocksource result")
		}

		assert.Equal(t, expectedSrc.Url, bsrc.Http.Url)
		assert.Equal(t, expectedSrc.RangeBegin, bsrc.Http.RangeBegin)
		assert.Equal(t, expectedSrc.RangeEnd, bsrc.Http.RangeEnd)
	}
}

// N.B. This works even with objectID unset in CacheMissRequest because we are translating back to the path "/foo/bar"
// in the call to BlockSource().  This is messy.
func (suite *CatalogTestSuite) TestBlockSource_Inline_WholeObject() {
	suite.testBlockSource(&ccmsg.CacheMissRequest{}, BlockSourceInline,
		&ccmsg.BlockSourceInline{
			Block: suite.policy.ChunkIntoBlocks(suite.objectData[:]),
		})
}

func (suite *CatalogTestSuite) TestBlockSource_HTTP_WholeObject() {
	suite.testBlockSource(&ccmsg.CacheMissRequest{}, BlockSourceHTTP,
		&ccmsg.BlockSourceHTTP{
			Url:        suite.ts.URL + "/foo/bar",
			RangeBegin: 0,
			RangeEnd:   0,
		})
}

func (suite *CatalogTestSuite) TestBlockSource_Inline_FirstBlock() {
	suite.testBlockSource(&ccmsg.CacheMissRequest{
		RangeBegin: 0,
		RangeEnd:   1,
	}, BlockSourceInline, &ccmsg.BlockSourceInline{
		Block: suite.policy.ChunkIntoBlocks(suite.objectData[:blockSize]),
	})
}

func (suite *CatalogTestSuite) TestBlockSource_HTTP_FirstBlock() {
	suite.testBlockSource(&ccmsg.CacheMissRequest{
		RangeBegin: 0,
		RangeEnd:   1,
	}, BlockSourceHTTP, &ccmsg.BlockSourceHTTP{
		Url:        suite.ts.URL + "/foo/bar",
		RangeBegin: 0,
		RangeEnd:   uint64(suite.blockSize),
	})
}

// Tests that the publisher/catalog returns the actual end of a partial final block instead of simply computing where
// the block would end were it full-size.
func (suite *CatalogTestSuite) TestBlockSource_Inline_PartialFinalBlock() {
	// Remove the last half-block.
	suite.objectData = suite.objectData[0 : len(suite.objectData)-(suite.blockSize/2)]

	expectedBlockCount := int(math.Ceil(float64(len(suite.objectData)) / float64(suite.blockSize)))

	blocks := suite.policy.ChunkIntoBlocks(suite.objectData[:])
	suite.testBlockSource(&ccmsg.CacheMissRequest{
		RangeBegin: uint64(expectedBlockCount - 1),
		RangeEnd:   uint64(expectedBlockCount),
	}, BlockSourceInline, &ccmsg.BlockSourceInline{
		Block: blocks[len(blocks)-1:],
	})
}

// Tests that the publisher/catalog returns the actual end of a partial final block instead of simply computing where
// the block would end were it full-size.
func (suite *CatalogTestSuite) TestBlockSource_HTTP_PartialFinalBlock() {
	// Remove the last half-block.
	suite.objectData = suite.objectData[0 : len(suite.objectData)-(suite.blockSize/2)]

	expectedBlockCount := int(math.Ceil(float64(len(suite.objectData)) / float64(suite.blockSize)))

	suite.testBlockSource(&ccmsg.CacheMissRequest{
		RangeBegin: uint64(expectedBlockCount - 1),
		RangeEnd:   uint64(expectedBlockCount),
	}, BlockSourceHTTP, &ccmsg.BlockSourceHTTP{
		Url:        suite.ts.URL + "/foo/bar",
		RangeBegin: uint64((expectedBlockCount - 1) * suite.blockSize),
		RangeEnd:   uint64(len(suite.objectData)),
	})
}

// TODO: Test with different blockSize.  See issue #17.
