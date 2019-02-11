package catalog

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/testutil"
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

	l   *logrus.Logger
	ts  *httptest.Server
	cat *catalog

	upstreamResponseDelay time.Duration
	objectData            []byte

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

	suite.objectData = testutil.RandBytesFromSource(rand.NewSource(dataSeed), 4*blockSize)
	suite.upstreamRequestQty = 0

	suite.ts = httptest.NewServer(http.HandlerFunc(suite.handleUpstreamRequest))

	upstream, err := NewHTTPUpstream(suite.l, suite.ts.URL)
	if err != nil {
		t.Fatalf("failed to create HTTP upstream: %v", err)
	}

	suite.cat, err = NewCatalog(suite.l, upstream)
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

	time.Sleep(suite.upstreamResponseDelay)

	switch r.URL.Path {
	case "/foo/bar":
		// TODO: Test behavior without this header (currently, causes catalog to panic; see comment in #16).
		w.Header().Add("Content-Length", fmt.Sprintf("%v", len(suite.objectData)))
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

	mm := make([]*ObjectMetadata, 2)
	for i := 0; i < len(mm); i++ {
		m, err := cat.GetData(ctx, &ccmsg.ContentRequest{Path: "/foo/bar"})
		assert.Nil(t, err)
		assert.NotNil(t, m)
		// assert.Nil(t, m.RespErr)
		assert.Equal(t, StatusOK, m.Status, "unexpected response status")

		mm[i] = m
	}

	// TODO: Add additional test cases covering how the object is split up into chunks.
	for i := 0; i < len(mm); i++ {
		assert.Equal(t, uint64(len(suite.objectData)), mm[i].ObjectSize())
	}

	// Due to caching, only a single request should be sent upstream.
	assert.Equal(t, 1, suite.upstreamRequestQty, "request for cached data should not create another upstream request")
}

func (suite *CatalogTestSuite) TestNotFound() {
	t := suite.T()
	cat := suite.cat

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	m, err := cat.GetData(ctx, &ccmsg.ContentRequest{Path: "/bogus"})
	assert.Nil(t, err)
	assert.NotNil(t, m)
	// assert.Nil(t, m.RespErr)
	assert.Equal(t, StatusNotFound, m.Status)
}

func (suite *CatalogTestSuite) TestUpstreamUnreachable() {
	t := suite.T()

	/// Deliberately pick a port we know nothing will be listening on.
	ts := httptest.NewServer(http.HandlerFunc(suite.handleUpstreamRequest))
	ts.Close()

	upstream, err := NewHTTPUpstream(suite.l, ts.URL)
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
	assert.Nil(t, err)
	assert.NotNil(t, m)

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
