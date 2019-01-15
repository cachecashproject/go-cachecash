package catalog

import (
	"context"
	"fmt"
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

// XXX: What is the difference between an error returned from this function and an error stored in the FetchResult
// struct?  When should we do one vs. the other?
func (up *httpUpstream) FetchData(ctx context.Context, path string, forceMetadata bool, rangeBegin, rangeEnd uint) (*FetchResult, error) {
	up.l.WithFields(logrus.Fields{
		"path":          path,
		"rangeBegin":    rangeBegin,
		"rangeEnd":      rangeEnd,
		"forceMetadata": forceMetadata,
	}).Info("upstream fetch")

	if rangeEnd != 0 && rangeEnd <= rangeBegin {
		return nil, errors.New("invalid byte range")
	}

	u, err := up.upstreamURL(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get upstream URL")
	}

	// TODO: Check Accept-Ranges header via HEAD; check Content-Range header in response; handle 416 responses.
	req, _ := http.NewRequest("GET", u, nil)
	if rangeBegin != 0 {
		if rangeEnd != 0 {
			// N.B.: HTTP ranges are inclusive; our ranges are [inclusive, exclusive).
			req.Header.Add("Range", fmt.Sprintf("bytes=%v-%v", rangeBegin, rangeEnd-1))
		} else {
			req.Header.Add("Range", fmt.Sprintf("bytes=%v-", rangeBegin))
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
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
		return nil, errors.Wrap(err, "unhandled HTTP status code from upstream")
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

func (up *httpUpstream) BlockSource(req *ccmsg.CacheMissRequest, path string, policy *ObjectPolicy) (*ccmsg.CacheMissResponse, error) {
	u, err := up.upstreamURL(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get upstream URL")
	}

	var rangeEnd uint64
	if req.RangeEnd != 0 {
		rangeEnd = req.RangeEnd * uint64(policy.BlockSize)
	}

	return &ccmsg.CacheMissResponse{
		Source: &ccmsg.CacheMissResponse_Http{
			Http: &ccmsg.BlockSourceHTTP{
				Url:        u,
				RangeBegin: req.RangeBegin * uint64(policy.BlockSize),
				RangeEnd:   rangeEnd,
			},
		},
	}, nil
}
