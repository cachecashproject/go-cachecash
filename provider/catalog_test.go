package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CatalogTestSuite struct {
	suite.Suite

	ts  *httptest.Server
	cat *catalog

	upstreamRequestQty int
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
