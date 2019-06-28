package client

import (
	"context"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/ed25519"
)

type publisherMock struct {
	mock.Mock
	chunks [][]byte
}

var _ publisherConnection = (*publisherMock)(nil)

func newPublisherMock() *publisherMock {
	return &publisherMock{}
}

func (pc *publisherMock) newCacheConnection(l *logrus.Logger, addr string, pubkey ed25519.PublicKey) (cacheConnection, error) {
	args := pc.Called(l, addr, pubkey)
	cr := args.Get(0).(*cacheMock)
	err := args.Error(1)
	return cr, err
}

func (pc *publisherMock) GetContent(ctx context.Context, req *ccmsg.ContentRequest) (*ccmsg.ContentResponse, error) {
	args := pc.Called(req)
	cr := args.Get(0).(*ccmsg.ContentResponse)
	err := args.Error(1)
	return cr, err
}

func (pc *publisherMock) Close(ctx context.Context) error {
	return nil
}

func (pc *publisherMock) QueueChunks(chunks [][]byte) {
	pc.chunks = chunks
}

func (pc *publisherMock) makeNewCacheCall(l *logrus.Logger, addr string, pubkey string, chunks ...[]byte) {
	pubkeyKey := ed25519.PublicKey(([]byte)(pubkey))
	pc.On("newCacheConnection", l,
		addr, pubkeyKey).Return(newCacheMock(addr, pubkeyKey, chunks), nil).Once()
}
