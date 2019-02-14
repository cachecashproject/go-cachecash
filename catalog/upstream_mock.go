package catalog

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/sirupsen/logrus"
)

type MockUpstream struct {
	l       *logrus.Logger
	Objects map[string][]byte
}

var _ Upstream = (*MockUpstream)(nil)

func NewMockUpstream(l *logrus.Logger) (*MockUpstream, error) {
	return &MockUpstream{
		l:       l,
		Objects: make(map[string][]byte),
	}, nil
}

func (up *MockUpstream) FetchData(ctx context.Context, path string, metadata *ObjectMetadata, rangeBegin, rangeEnd uint) (*FetchResult, error) {
	up.l.WithFields(logrus.Fields{
		"path":          path,
		"rangeBegin":    rangeBegin,
		"rangeEnd":      rangeEnd,
	}).Info("upstream fetch")

	data, ok := up.Objects[path]
	if !ok {
		return &FetchResult{status: StatusNotFound}, nil
	}

	if rangeEnd != 0 && rangeEnd > uint(len(data)) {
		return nil, errors.New("invalid range")
	}

	respData := data
	if rangeEnd == 0 {
		rangeEnd = uint(len(respData))
	}
	respData = respData[rangeBegin:rangeEnd]

	up.l.Debugf("mock upstream fetch: responding to request for bytes [%v, %v) with %v bytes", rangeBegin, rangeEnd, len(respData))

	h := http.Header{
		"Content-Length": []string{fmt.Sprintf("%v", len(respData))},
	}
	if rangeBegin != 0 || rangeEnd != 0 {
		// TODO: This still isn't quite the same as "if they specified a range in the request..."
		h["Content-Range"] = []string{fmt.Sprintf("bytes %v-%v/%v", rangeBegin, rangeEnd-1, len(data))}
	}

	return &FetchResult{
		header: h,
		data:   respData,
		status: StatusOK,
	}, nil
}

func (up *MockUpstream) AddRandomObject(path string, size uint) {
	data := make([]byte, size)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}
	up.Objects[path] = data
}

func (up *MockUpstream) BlockSource(req *ccmsg.CacheMissRequest, path string, meta *ccmsg.ObjectMetadata,
	policy *ObjectPolicy) (*ccmsg.CacheMissResponse, error) {

	return nil, errors.New("mock upstream does not support BlockSource()")
}
