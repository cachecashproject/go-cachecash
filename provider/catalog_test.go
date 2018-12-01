package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CatalogTestSuite struct {
	suite.Suite

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

	suite.upstreamRequestQty = 0
	suite.ts = httptest.NewServer(http.HandlerFunc(suite.handleUpstreamRequest))

	var err error
	suite.cat, err = newCatalog(suite.ts.URL)
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
	fmt.Fprintln(w, "Hello, client")
}

func (suite *CatalogTestSuite) TestSimple() {
	t := suite.T()
	cat := suite.cat

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	m, err := cat.getObjectMetadata(ctx, "/foo/bar")
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

	mm := make([]*objectMetadata, 2)
	for i := 0; i < len(mm); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			m, err := cat.getObjectMetadata(ctx, "/foo/bar")
			assert.Nil(t, err)
			assert.NotNil(t, m)

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

	mm := make([]*objectMetadata, 2)
	for i := 0; i < len(mm); i++ {
		m, err := cat.getObjectMetadata(ctx, "/foo/bar")
		assert.Nil(t, err)
		assert.NotNil(t, m)

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

// TestCacheExpired
//  - case: object has not changed; revalidated
//  - case: object has changed; not revalidated
