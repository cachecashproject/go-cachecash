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
}

var _ publisherConnection = (*publisherMock)(nil)

func newPublisherMock() *publisherMock {
	return &publisherMock{}
}

func (pc *publisherMock) newCacheConnection(ctx context.Context, l *logrus.Logger, addr string, pubkey ed25519.PublicKey) (cacheConnection, error) {
	return newCacheMock(addr, pubkey), nil
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
