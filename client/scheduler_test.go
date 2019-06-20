package client

import (
	"context"
	"errors"
	"testing"

	cachecash "github.com/cachecashproject/go-cachecash"
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestSchedulerTestSuite(t *testing.T) {
	suite.Run(t, new(SchedulerTestSuite))
}

type SchedulerTestSuite struct {
	suite.Suite

	l *logrus.Logger
}

func (suite *SchedulerTestSuite) SetupTest() {
	l := logrus.New()
	suite.l = l
}

func (suite *SchedulerTestSuite) newMock() (*client, *publisherMock) {
	mock := newPublisherMock()
	return &client{
		l:             suite.l,
		publisherConn: mock,
		cacheConns:    map[cacheID]cacheConnection{},
	}, mock
}

func (suite *SchedulerTestSuite) newContentResponse() *ccmsg.ContentResponse {
	return &ccmsg.ContentResponse{
		Bundles: []*ccmsg.TicketBundle{
			{
				Remainder: &ccmsg.TicketBundleRemainder{
					PuzzleInfo: &ccmsg.ColocationPuzzleInfo{
						Goal: []byte{},
					},
				},
				TicketRequest: []*ccmsg.TicketRequest{
					{
						ChunkIdx: 0,
						InnerKey: &ccmsg.BlockKey{
							Key: []byte{},
						},
						CachePublicKey: &ccmsg.PublicKey{
							PublicKey: []byte{},
						},
					},
					{
						ChunkIdx: 1,
						InnerKey: &ccmsg.BlockKey{
							Key: []byte{},
						},
						CachePublicKey: &ccmsg.PublicKey{
							PublicKey: []byte{},
						},
					},
				},
				CacheInfo: []*ccmsg.CacheInfo{
					suite.newCache([]byte{0, 1, 2, 3, 4}),
					suite.newCache([]byte{5, 6, 7, 8, 9}),
				},
				Metadata: &ccmsg.ObjectMetadata{
					ObjectSize: 512,
					ChunkSize:  128,
				},
			},
		},
	}
}

func (suite *SchedulerTestSuite) newCache(pubkey []byte) *ccmsg.CacheInfo {
	return &ccmsg.CacheInfo{
		Pubkey: &ccmsg.PublicKey{
			PublicKey: pubkey,
		},
		Addr: &ccmsg.NetworkAddress{},
	}
}

func (suite *SchedulerTestSuite) TestScheduler() {
	t := suite.T()
	cl, mock := suite.newMock()

	mock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            "/",
		RangeBegin:      0,
		RangeEnd:        0,
		BacklogDepth:    map[string]uint64{},
	}).Return(suite.newContentResponse(), nil)
	mock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            "/",
		RangeBegin:      256,
		RangeEnd:        0,
		BacklogDepth: map[string]uint64{
			"\000\001\002\003\004": 2,
		},
	}).Return(suite.newContentResponse(), nil)

	queue := make(chan *fetchGroup, 128)
	cl.schedule(context.Background(), "/", queue)

	group := <-queue
	assert.Nil(t, group.err)
	assert.NotNil(t, group.bundle)

	group = <-queue
	assert.Nil(t, group.err)
	assert.NotNil(t, group.bundle)

	assert.Zero(t, len(queue))
}

func (suite *SchedulerTestSuite) TestSchedulerError() {
	t := suite.T()
	cl, mock := suite.newMock()

	mock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            "/",
		RangeBegin:      0,
		RangeEnd:        0,
		BacklogDepth:    map[string]uint64{},
	}).Return((*ccmsg.ContentResponse)(nil), errors.New("this is an error"))

	queue := make(chan *fetchGroup, 128)
	cl.schedule(context.Background(), "/", queue)

	group := <-queue
	assert.Nil(t, group.bundle)
	assert.Equal(t, "failed to fetch chunk-group at offset 0: failed to request bundle from publisher: this is an error", group.err.Error())
	assert.Zero(t, len(queue))
}

func (suite *SchedulerTestSuite) TestRequestBundle() {
	t := suite.T()
	cl, mock := suite.newMock()

	mock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            "/",
		RangeBegin:      0,
		RangeEnd:        0,
		BacklogDepth:    map[string]uint64{},
	}).Return(&ccmsg.ContentResponse{
		Bundles: []*ccmsg.TicketBundle{},
	}, nil)

	resp, err := cl.requestBundles(context.Background(), "/", 0)
	assert.Nil(t, err, "failed to get bundle")
	assert.Equal(t, []*ccmsg.TicketBundle{}, resp)
}

func (suite *SchedulerTestSuite) TestRequestBundleError() {
	t := suite.T()
	cl, mock := suite.newMock()

	mock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            "/",
		RangeBegin:      0,
		RangeEnd:        0,
		BacklogDepth:    map[string]uint64{},
	}).Return((*ccmsg.ContentResponse)(nil), errors.New("this is an error"))

	resp, err := cl.requestBundles(context.Background(), "/", 0)
	assert.NotNil(t, err)
	assert.Nil(t, resp)
}
