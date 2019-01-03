package catalog

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CatalogTestSuite struct {
	suite.Suite

	l   *logrus.Logger
	ts  *httptest.Server
	cat *catalog

	upstreamRequestQty    int
	upstreamResponseDelay time.Duration
}

func TestCatalogTestSuite(t *testing.T) {
	suite.Run(t, new(CatalogTestSuite))
}

func (suite *CatalogTestSuite) SetupTest() {
	t := suite.T()

	suite.l = logrus.New()
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

func (suite *CatalogTestSuite) handleUpstreamRequest(w http.ResponseWriter, r *http.Request) {
	suite.upstreamRequestQty++
	time.Sleep(suite.upstreamResponseDelay)

	switch r.URL.Path {
	case "/foo/bar":
		fmt.Fprintln(w, "Hello, client")
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (suite *CatalogTestSuite) TestSimple() {
	t := suite.T()
	cat := suite.cat

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	m, err := cat.GetObjectMetadata(ctx, "/foo/bar")
	assert.Nil(t, err)
	assert.NotNil(t, m)

	// This should indicate a Content-Length of 14 ("Hello, client\n").
	// TODO: Replace with an assertion once the test server is serving actual data.
	fmt.Printf("object metadata: %v\n", m)
}

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

			m, err := cat.GetObjectMetadata(ctx, "/foo/bar")
			assert.Nil(t, err)
			assert.NotNil(t, m)
			assert.Nil(t, m.RespErr)
			assert.Equal(t, StatusOK, m.Status)

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

// Once we have a valid cache entry, receiving another downstream request should not cause us to make an upstream
// request.
func (suite *CatalogTestSuite) TestCacheValid() {
	t := suite.T()
	cat := suite.cat

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	mm := make([]*ObjectMetadata, 2)
	for i := 0; i < len(mm); i++ {
		m, err := cat.GetObjectMetadata(ctx, "/foo/bar")
		assert.Nil(t, err)
		assert.NotNil(t, m)
		assert.Nil(t, m.RespErr)
		assert.Equal(t, StatusOK, m.Status)

		mm[i] = m
	}

	// This should indicate a Content-Length of 14 ("Hello, client\n").
	// TODO: Replace with an assertion once the test server is serving actual data.
	for i := 0; i < len(mm); i++ {
		fmt.Printf("object metadata %v: %v\n", i, mm[i])
	}

	// Due to caching, only a single request should be sent upstream.
	assert.Equal(t, 1, suite.upstreamRequestQty)
}

func (suite *CatalogTestSuite) TestNotFound() {
	t := suite.T()
	cat := suite.cat

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	m, err := cat.GetObjectMetadata(ctx, "/bogus")
	assert.Nil(t, err)
	assert.NotNil(t, m)
	assert.Nil(t, m.RespErr)
	assert.Equal(t, StatusNotFound, m.Status)
}

func (suite *CatalogTestSuite) TestUpstreamUnreachable() {
	t := suite.T()

	/// Deliberately pick a port we know nothing will be listening on.
	ts := httptest.NewServer(http.HandlerFunc(suite.handleUpstreamRequest))
	ts.Close()

	upstream, err := NewHTTPUpstream(ts.l, ts.URL)
	if err != nil {
		t.Fatalf("failed to create HTTP upstream: %v", err)
	}

	cat, err := NewCatalog(suite.l, upstream)
	if err != nil {
		t.Fatalf("failed to create catalog: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	m, err := cat.GetObjectMetadata(ctx, "/foo/bar")
	assert.Nil(t, err)
	assert.NotNil(t, m)

	// This should indicate a Content-Length of 14 ("Hello, client\n").
	// TODO: Replace with an assertion once the test server is serving actual data.
	fmt.Printf("object metadata: %v\n", m)

}

// TestUpstreamTimeout covers what happens if the upstream does not respond before our request times out.
func (suite *CatalogTestSuite) TestUpstreamTimeout() {
	t := suite.T()
	cat := suite.cat

	suite.upstreamResponseDelay = 1 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	m, err := cat.GetObjectMetadata(ctx, "/foo/bar")
	assert.Equal(t, context.DeadlineExceeded, err)
	assert.Nil(t, m)
}

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
