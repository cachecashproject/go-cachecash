package catalog

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/kelleyk/go-cachecash/ccmsg"
)

//go:generate stringer -type=ObjectStatus

type ObjectStatus int

const (
	StatusUnknown ObjectStatus = iota
	StatusOK
	StatusNotFound
	StatusInternalError
	StatusUpstreamUnreachable
	StatusUpstreamError
)

// FetchResult describes the metadata, data, and/or errors returned by an Upstream in response to a single request.
// They are consumed by the catalog, which uses them to update its cache.
type FetchResult struct {
	// XXX: This should be changed; this struct is supposed to be protocol-agnostic.
	header http.Header
	data   []byte
	status ObjectStatus
}

// ObjectSize returns the size of the entire object; the response might actually contain data for only some part of it.
// TODO: Support '*' in the Content-Range header.
//
// Valid formats for the 'Content-Range' header:
//   'bytes 0-49/500'    // The first 50 bytes of a 500-byte object.
//   'bytes 0-49/*'      // The first 50 bytes of an object whose length is unknown.
//   'bytes */1234'      // Used in 416 (Range Not Satisfiable) responses.
//
func (r *FetchResult) ObjectSize() (int, error) {
	cr := r.header.Get("Content-Range")
	cl := r.header.Get("Content-Length")

	if cr == "" {
		if cl == "" {
			return 0, errors.New("response contains neither Content-Length nor Content-Range header")
		}
		return strconv.Atoi(cl)
	}

	parts := strings.Fields(cr)
	if len(parts) != 2 {
		return 0, errors.New("Content-Range header has unexpected number of words")
	}
	if parts[0] != "bytes" {
		return 0, errors.New("Content-Range header contains unsupported length unit")
	}
	var rangeBegin, rangeEnd, objectSize int
	if _, err := fmt.Sscanf(parts[1], "%d-%d/%d", &rangeBegin, &rangeEnd, &objectSize); err != nil {
		return 0, errors.New("Content-Range header has unexpected format")
	}
	return objectSize, nil
}

type Upstream interface {
	// FetchData ensures that metadata is fresh, and also that the indicated blocks are available in the cache.  An
	// empty list of block indices is allowed; this ensures metadata freshness but does not pull any data blocks.
	//
	// forceMetadata requires that object metadata be fetched even if it would not otherwise be fetched.
	//
	// rangeEnd must be >= rangeBegin.  rangeEnd == 0 means "continue to he end of the object".
	//
	// Cases:
	// - We want to fetch metadata only.
	// - We want to fetch metadata *and* a series of blocks.
	// - We have metadata that we already believe to be be valid, so we don't necessarily need to fetch it, if that's
	//   any extra effort.  We want to fetch a series of blocks.
	//
	// Questions:
	// - Some upstreams will require CacheCash (not upstream) metadata for the object.  For example, the HTTP upstream
	//   will need to know block sizes in order to translate block indices into byte ranges.  How should this be done?
	//
	FetchData(ctx context.Context, path string, forceMetadata bool, rangeBegin, rangeEnd uint) (*FetchResult, error)

	CacheMiss(path string, rangeBegin, rangeEnd uint64) (*ccmsg.CacheMissResponse, error)
}

type ContentCatalog interface {
	GetData(ctx context.Context, req *ccmsg.ContentRequest) (*ObjectMetadata, error)

	// XXX: Temporary; remove once refactoring is complete.
	GetObjectMetadata(ctx context.Context, path string) (*ObjectMetadata, error)

	CacheMiss(path string, rangeBegin, rangeEnd uint64) (*ccmsg.CacheMissResponse, error)
}

type ContentLocator interface {
	GetContentSource(ctx context.Context, req *ccmsg.CacheMissRequest) (*ccmsg.CacheMissResponse, error)
}
