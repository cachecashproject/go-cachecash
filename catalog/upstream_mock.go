package catalog

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"

	"github.com/kelleyk/go-cachecash/ccmsg"
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

// N.B.: Only the `Source` field should be populated in the return value.
func (up *MockUpstream) CacheMiss(path string, rangeBegin, rangeEnd uint64) (*ccmsg.CacheMissResponse, error) {
	panic("no impl")
}

func (up *MockUpstream) FetchData(ctx context.Context, path string, forceMetadata bool, rangeBegin, rangeEnd uint) (*FetchResult, error) {
	up.l.WithFields(logrus.Fields{
		"path":          path,
		"rangeBegin":    rangeBegin,
		"rangeEnd":      rangeEnd,
		"forceMetadata": forceMetadata,
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

	return &FetchResult{
		header: http.Header{
			"Content-Length": []string{fmt.Sprintf("%v", len(respData))},
			"Content-Range":  []string{fmt.Sprintf("bytes %v-%v/%v", rangeBegin, rangeEnd-1, len(data))},
		},
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

/*
func (up *MockUpstream) GetBlock(path string, blockIdx uint) ([]byte, error) {
	data, ok := up.Objects[path]
	if !ok {
		return nil, errors.
	}
	buf := data[
	return nil, nil
}
*/
