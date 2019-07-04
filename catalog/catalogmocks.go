package catalog

import (
	"context"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/stretchr/testify/mock"
)

// ContentCatalogMock mocks the ContentCatalog interface for testing.
type ContentCatalogMock struct {
	mock.Mock
}

var _ ContentCatalog = (*ContentCatalogMock)(nil)

// NewContentCatalogMock creates a new mock
func NewContentCatalogMock() *ContentCatalogMock {
	return &ContentCatalogMock{}
}

// GetData is defined on the ContentCatalog interface
func (ccm *ContentCatalogMock) GetData(ctx context.Context, req *ccmsg.ContentRequest) (*ObjectMetadata, error) {
	args := ccm.Called(ctx, req)
	om := args.Get(0).(*ObjectMetadata)
	err := args.Error(1)
	return om, err
}

// GetMetadata is defined on the ContentCatalog interface
func (ccm *ContentCatalogMock) GetMetadata(ctx context.Context, path string) (*ObjectMetadata, error) {
	args := ccm.Called(ctx, path)
	om := args.Get(0).(*ObjectMetadata)
	err := args.Error(1)
	return om, err
}

// ChunkSource is defined on the ContentCatalog interface
func (ccm *ContentCatalogMock) ChunkSource(ctx context.Context, req *ccmsg.CacheMissRequest, path string, metadata *ObjectMetadata) (*ccmsg.Chunk, error) {
	args := ccm.Called(ctx, req, path, metadata)
	chunk := args.Get(0).(*ccmsg.Chunk)
	err := args.Error(1)
	return chunk, err

}

// Upstream is defined on the ContentCatalog interface
func (ccm *ContentCatalogMock) Upstream(path string) (Upstream, error) {
	args := ccm.Called(path)
	upstream := args.Get(0).(Upstream)
	err := args.Error(1)
	return upstream, err

}
