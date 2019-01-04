package catalog

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/kelleyk/go-cachecash/ccmsg"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type httpUpstream struct {
	l       *logrus.Logger
	baseURL *url.URL
}

var _ Upstream = (*httpUpstream)(nil)

func NewHTTPUpstream(l *logrus.Logger, baseURL string) (Upstream, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse upstream URL")
	}

	return &httpUpstream{
		l:       l,
		baseURL: u,
	}, nil
}

func (up *httpUpstream) upstreamURL(path string) (string, error) {

	// XXX: Need to ensure that `pathURL` is not an absolute URL; that could be used to make the publisher fetch
	// arbitrary data.
	pathURL, err := url.Parse(path)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse object path")
	}
	u := up.baseURL.ResolveReference(pathURL)

	return u.String(), nil
}

// N.B.: Only the `Source` field should be populated in the return value.
func (up *httpUpstream) CacheMiss(path string, rangeBegin, rangeEnd uint64) (*ccmsg.CacheMissResponse, error) {
	panic("no impl")
}

// XXX: What is the difference between an error returned from this function and an error stored in the FetchResult
// struct?  When should we do one vs. the other?
func (up *httpUpstream) FetchData(ctx context.Context, path string, forceMetadata bool, blockOffset, blockCount int) (*FetchResult, error) {
	up.l.WithFields(logrus.Fields{"path": path}).Info("upstream fetch")

	u, err := up.upstreamURL(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get upstream URL")
	}

	resp, err := http.Get(u)
	if err != nil {
		return nil, errors.Wrap(err, "failed HTTP fetch")
	}

	// XXX: Should be using a HEAD request instead.
	// XXX: Should be acting on HTTP status code.

	defer func() {
		_ = resp.Body.Close()
	}()

	// Need to handle
	// 203 Non-Authoritative Information
	// 204 No Content
	// 206 Partial Content
	// 3xx redirections (handled by the Golang HTTP client?)
	// 304 Not Changed
	// 4xx client errors (probably point to a configuration problem?)
	var status ObjectStatus
	switch {
	case resp.StatusCode == http.StatusOK:
		status = StatusOK
	case resp.StatusCode == http.StatusNotFound:
		status = StatusNotFound
	case resp.StatusCode >= 500 && resp.StatusCode < 600:
		status = StatusUpstreamError
	default:
		panic("unhandled HTTP status code from upstream")
	}

	var body []byte
	if status == StatusOK {
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read request body")
		}
	}

	return &FetchResult{
		header: resp.Header,
		data:   body,
		status: status,
	}, nil
}
