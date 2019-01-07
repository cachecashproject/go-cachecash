package catalog

import (
	"context"
	"crypto/rand"
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

	return &FetchResult{
		header: http.Header{},
		data:   data,
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
