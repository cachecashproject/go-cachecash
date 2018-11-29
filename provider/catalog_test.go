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
}

func TestCatalogTestSuite(t *testing.T) {
	suite.Run(t, new(CatalogTestSuite))
}

func (suite *CatalogTestSuite) handleUpstreamRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, client")
}

func (suite *CatalogTestSuite) TestSimple() {
	t := suite.T()

	ts := httptest.NewServer(http.HandlerFunc(suite.handleUpstreamRequest))
	defer ts.Close()

	cat, err := newCatalog(ts.URL)
	assert.Nil(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	m, err := cat.getObjectMetadata(ctx, "/foo/bar")
	assert.Nil(t, err)
	assert.NotNil(t, m)

	fmt.Printf("object metadata: %v\n", m)
}
