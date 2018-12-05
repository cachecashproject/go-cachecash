package catalog

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

type httpUpstream struct {
	baseURL *url.URL
}

var _ Upstream = (*httpUpstream)(nil)

func NewHTTPUpstream(baseURL string) (Upstream, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse upstream URL")
	}

	return &httpUpstream{
		baseURL: u,
	}, nil
}

// XXX: What is the difference between an error returned from this function and an error stored in the FetchResult
// struct?  When should we do one vs. the other?
func (up *httpUpstream) FetchData(ctx context.Context, path string, forceMetadata bool, blockOffset, blockCount int) (*FetchResult, error) {

	pathURL, err := url.Parse(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse object path")
	}
	u := up.baseURL.ResolveReference(pathURL)

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed HTTP fetch")
	}

	// XXX: Should be using a HEAD request instead.
	// XXX: Should be acting on HTTP status code.

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read request body")
	}

	return &FetchResult{
		header: resp.Header,
		data:   body,
	}, nil
}
