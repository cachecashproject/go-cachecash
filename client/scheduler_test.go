package client

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	cachecash "github.com/cachecashproject/go-cachecash"
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/ed25519"
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
	suite.l.SetLevel(logrus.DebugLevel)
}

func (suite *SchedulerTestSuite) newMock() (*client, *publisherMock) {
	mock := newPublisherMock()
	return &client{
		l:             suite.l,
		publisherConn: mock,
		cacheConns:    map[cacheID]cacheConnection{},
	}, mock
}

type CROptions struct {
	bundles      uint64
	objectSize   uint64
	chunkSize    uint64
	bundleOffset uint64
}

func (suite *SchedulerTestSuite) newContentResponse(options ...CROptions) *ccmsg.ContentResponse {
	// Defaults
	opts := CROptions{
		bundles:      0,
		objectSize:   512,
		chunkSize:    128,
		bundleOffset: 0,
	}
	// Merge explicit choices
	for _, opt := range options {
		if opt.bundles != 0 {
			opts.bundles = opt.bundles
		}
		if opt.objectSize != 0 {
			opts.objectSize = opt.objectSize
		}
		if opt.chunkSize != 0 {
			opts.chunkSize = opt.chunkSize
		}
		if opt.bundleOffset != 0 {
			opts.bundleOffset = opt.bundleOffset
		}
	}

	bundles := []*ccmsg.TicketBundle{}

	for i := opts.bundleOffset; i < opts.bundles+opts.bundleOffset; i++ {
		chunkIdx := i * 2 // 2 chunks per bundle
		bundles = append(bundles, &ccmsg.TicketBundle{
			Remainder: &ccmsg.TicketBundleRemainder{
				PuzzleInfo: &ccmsg.ColocationPuzzleInfo{
					Goal: []byte{},
				},
			},
			TicketRequest: []*ccmsg.TicketRequest{
				{
					ChunkIdx: chunkIdx,
					InnerKey: &ccmsg.BlockKey{
						Key: []byte{},
					},
					CachePublicKey: &ccmsg.PublicKey{
						PublicKey: []byte{},
					},
				},
				{
					ChunkIdx: chunkIdx + 1,
					InnerKey: &ccmsg.BlockKey{
						Key: []byte{},
					},
					CachePublicKey: &ccmsg.PublicKey{
						PublicKey: []byte{},
					},
				},
			},
			CacheInfo: []*ccmsg.CacheInfo{
				suite.newCache(net.ParseIP("192.0.2.1"), 1001, []byte{0, 1, 2, 3, 4}),
				suite.newCache(net.ParseIP("192.0.2.2"), 1002, []byte{5, 6, 7, 8, 9}),
			},
			Metadata: &ccmsg.ObjectMetadata{
				ObjectSize: opts.objectSize,
				ChunkSize:  opts.chunkSize,
			},
		})
	}
	return &ccmsg.ContentResponse{
		Bundles: bundles,
	}
}

func (suite *SchedulerTestSuite) newCache(ip net.IP, port uint32, pubkey []byte) *ccmsg.CacheInfo {
	return &ccmsg.CacheInfo{
		Pubkey: &ccmsg.PublicKey{
			PublicKey: pubkey,
		},
		Addr: &ccmsg.NetworkAddress{
			Inetaddr: ip,
			Port:     port,
		},
	}
}

func (suite *SchedulerTestSuite) TestSchedulerOneBundle() {
	t := suite.T()
	cl, mock := suite.newMock()

	mock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            "/",
		RangeBegin:      0,
		RangeEnd:        0,
		CacheStatus:     map[string]*ccmsg.ContentRequest_ClientCacheStatus{},
	}).Return(suite.newContentResponse(CROptions{bundles: 1}), nil).Once()
	mock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            "/",
		RangeBegin:      256,
		RangeEnd:        0,
		CacheStatus: map[string]*ccmsg.ContentRequest_ClientCacheStatus{
			"\x00\x01\x02\x03\x04": &ccmsg.ContentRequest_ClientCacheStatus{
				BacklogDepth: 0x1,
			},
			"\x05\x06\x07\x08\x09": &ccmsg.ContentRequest_ClientCacheStatus{
				BacklogDepth: 0x1,
			},
		},
	}).Return(suite.newContentResponse(CROptions{bundles: 1}), nil).Once()
	mock.makeNewCacheCall(cl.l, "192.0.2.1:1001", "\x00\x01\x02\x03\x04")
	mock.makeNewCacheCall(cl.l, "192.0.2.2:1002", "\x05\x06\x07\x08\x09")

	queue := make(chan *fetchGroup, 128)
	bundleCompletions := make(chan BundleOutcome, 128)
	bundleCompletions <- BundleOutcome{Outcome: Completed, ChunkOffset: 0, Chunks: 2}
	bundleCompletions <- BundleOutcome{Outcome: Completed, ChunkOffset: 2, Chunks: 2}
	cl.schedule(context.Background(), "/", queue, bundleCompletions)

	group := <-queue
	assert.Nil(t, group.err)
	assert.NotNil(t, group.bundle)

	group = <-queue
	assert.Nil(t, group.err)
	assert.NotNil(t, group.bundle)

	assert.Zero(t, len(queue))
}

func (suite *SchedulerTestSuite) TestSchedulerZeroBundles() {
	t := suite.T()
	cl, mock := suite.newMock()

	mock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            "/",
		RangeBegin:      0,
		RangeEnd:        0,
		CacheStatus:     map[string]*ccmsg.ContentRequest_ClientCacheStatus{},
	}).Return(suite.newContentResponse(), nil).Once()
	mock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            "/",
		RangeBegin:      0,
		RangeEnd:        0,
		CacheStatus:     map[string]*ccmsg.ContentRequest_ClientCacheStatus{},
	}).Return(suite.newContentResponse(CROptions{bundles: 1}), nil).Once()
	mock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            "/",
		RangeBegin:      256,
		RangeEnd:        0,
		CacheStatus: map[string]*ccmsg.ContentRequest_ClientCacheStatus{
			"\x00\x01\x02\x03\x04": &ccmsg.ContentRequest_ClientCacheStatus{
				BacklogDepth: 0x1,
			},
			"\x05\x06\x07\x08\x09": &ccmsg.ContentRequest_ClientCacheStatus{
				BacklogDepth: 0x1,
			},
		},
	}).Return(suite.newContentResponse(), nil).Once()
	mock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            "/",
		RangeBegin:      256,
		RangeEnd:        0,
		CacheStatus: map[string]*ccmsg.ContentRequest_ClientCacheStatus{
			"\x00\x01\x02\x03\x04": &ccmsg.ContentRequest_ClientCacheStatus{
				BacklogDepth: 0x1,
			},
			"\x05\x06\x07\x08\x09": &ccmsg.ContentRequest_ClientCacheStatus{
				BacklogDepth: 0x1,
			},
		},
	}).Return(suite.newContentResponse(CROptions{bundles: 1}), nil).Once()
	mock.makeNewCacheCall(cl.l, "192.0.2.1:1001", "\x00\x01\x02\x03\x04")
	mock.makeNewCacheCall(cl.l, "192.0.2.2:1002", "\x05\x06\x07\x08\x09")

	queue := make(chan *fetchGroup, 128)
	bundleCompletions := make(chan BundleOutcome, 128)
	bundleCompletions <- BundleOutcome{Outcome: Completed, ChunkOffset: 0, Chunks: 2}
	bundleCompletions <- BundleOutcome{Outcome: Completed, ChunkOffset: 2, Chunks: 2}
	cl.schedule(context.Background(), "/", queue, bundleCompletions)

	group := <-queue
	assert.Nil(t, group.err)
	assert.NotNil(t, group.bundle)

	group = <-queue
	assert.Nil(t, group.err)
	assert.NotNil(t, group.bundle)

	assert.Zero(t, len(queue))
}

func (suite *SchedulerTestSuite) TestSchedulerAllBundlesAtOnce() {
	t := suite.T()
	cl, mock := suite.newMock()

	mock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            "/",
		RangeBegin:      0,
		RangeEnd:        0,
		CacheStatus:     map[string]*ccmsg.ContentRequest_ClientCacheStatus{},
	}).Return(suite.newContentResponse(CROptions{bundles: 2}), nil).Once()
	mock.makeNewCacheCall(cl.l, "192.0.2.1:1001", "\x00\x01\x02\x03\x04")
	mock.makeNewCacheCall(cl.l, "192.0.2.2:1002", "\x05\x06\x07\x08\x09")

	queue := make(chan *fetchGroup, 128)
	bundleCompletions := make(chan BundleOutcome, 128)
	bundleCompletions <- BundleOutcome{Outcome: Completed, ChunkOffset: 0, Chunks: 2}
	bundleCompletions <- BundleOutcome{Outcome: Completed, ChunkOffset: 2, Chunks: 2}
	cl.schedule(context.Background(), "/", queue, bundleCompletions)

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
		CacheStatus:     map[string]*ccmsg.ContentRequest_ClientCacheStatus{},
	}).Return((*ccmsg.ContentResponse)(nil), errors.New("this is an error")).Once()

	queue := make(chan *fetchGroup, 128)
	bundleCompletions := make(chan BundleOutcome, 128)
	cl.schedule(context.Background(), "/", queue, bundleCompletions)

	group := <-queue
	assert.Nil(t, group.bundle)
	assert.Equal(t, "failed to fetch chunk-group at chunk offset 0: this is an error", group.err.Error())
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
		CacheStatus:     map[string]*ccmsg.ContentRequest_ClientCacheStatus{},
	}).Return(&ccmsg.ContentResponse{
		Bundles: []*ccmsg.TicketBundle{},
	}, nil).Once()

	resp, err := cl.requestBundles(context.Background(), "/", 0, 0)
	assert.Nil(t, err, "failed to get bundle")
	assert.Equal(t, []*ccmsg.TicketBundle{}, resp)
}

func (suite *SchedulerTestSuite) TestRequestLimitedBundle() {
	t := suite.T()
	cl, mock := suite.newMock()
	var chunkSize uint64 = 512
	cl.chunkSize = &chunkSize

	mock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            "/",
		RangeBegin:      0,
		RangeEnd:        1024,
		CacheStatus:     map[string]*ccmsg.ContentRequest_ClientCacheStatus{},
	}).Return(&ccmsg.ContentResponse{
		Bundles: []*ccmsg.TicketBundle{},
	}, nil).Once()

	resp, err := cl.requestBundles(context.Background(), "/", 0, 2)
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
		CacheStatus:     map[string]*ccmsg.ContentRequest_ClientCacheStatus{},
	}).Return((*ccmsg.ContentResponse)(nil), errors.New("this is an error")).Once()

	resp, err := cl.requestBundles(context.Background(), "/", 0, 0)
	assert.NotNil(t, err)
	assert.Nil(t, resp)
}

func (suite *SchedulerTestSuite) TestCacheConnectionError() {
	t := suite.T()
	cl, mock := suite.newMock()

	mock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            "/",
		RangeBegin:      0,
		RangeEnd:        0,
		CacheStatus:     map[string]*ccmsg.ContentRequest_ClientCacheStatus{},
	}).Return(suite.newContentResponse(), nil).Once()
	mock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            "/",
		RangeBegin:      0,
		RangeEnd:        0,
		CacheStatus:     map[string]*ccmsg.ContentRequest_ClientCacheStatus{},
	}).Return(suite.newContentResponse(CROptions{bundles: 1}), nil).Once()
	mock.On("newCacheConnection", cl.l, "192.0.2.1:1001",
		ed25519.PublicKey(([]byte)("\x00\x01\x02\x03\x04"))).Return(
		(*cacheMock)(nil), errors.New("cache connection failure")).Once()

	queue := make(chan *fetchGroup, 128)
	bundleCompletions := make(chan BundleOutcome, 128)
	cl.schedule(context.Background(), "/", queue, bundleCompletions)

	group := <-queue
	assert.NotNil(t, group.err)
	assert.Nil(t, group.bundle)

	assert.Zero(t, len(queue))
}

func (suite *SchedulerTestSuite) TestChangedChunkCount() {
	t := suite.T()
	cl, mock := suite.newMock()

	mock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            "/",
		RangeBegin:      0,
		RangeEnd:        0,
		CacheStatus:     map[string]*ccmsg.ContentRequest_ClientCacheStatus{},
	}).Return(suite.newContentResponse(CROptions{bundles: 1}), nil).Once()
	mock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            "/",
		RangeBegin:      256,
		RangeEnd:        0,
		CacheStatus: map[string]*ccmsg.ContentRequest_ClientCacheStatus{
			"\x00\x01\x02\x03\x04": &ccmsg.ContentRequest_ClientCacheStatus{
				BacklogDepth: 0x1,
			},
			"\x05\x06\x07\x08\x09": &ccmsg.ContentRequest_ClientCacheStatus{
				BacklogDepth: 0x1,
			},
		},
	}).Return(suite.newContentResponse(CROptions{bundles: 1, objectSize: 1024}), nil).Once()
	mock.makeNewCacheCall(cl.l, "192.0.2.1:1001", "\x00\x01\x02\x03\x04")
	mock.makeNewCacheCall(cl.l, "192.0.2.2:1002", "\x05\x06\x07\x08\x09")

	queue := make(chan *fetchGroup, 128)
	bundleCompletions := make(chan BundleOutcome, 128)
	cl.schedule(context.Background(), "/", queue, bundleCompletions)

	group := <-queue
	assert.Nil(t, group.err)
	assert.NotNil(t, group.bundle)

	group = <-queue
	assert.NotNil(t, group.err)
	assert.Nil(t, group.bundle)

	assert.Zero(t, len(queue))
}

func (suite *SchedulerTestSuite) TestChangedChunkSize() {
	t := suite.T()
	cl, mock := suite.newMock()

	mock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            "/",
		RangeBegin:      0,
		RangeEnd:        0,
		CacheStatus:     map[string]*ccmsg.ContentRequest_ClientCacheStatus{},
	}).Return(suite.newContentResponse(CROptions{bundles: 1}), nil).Once()
	mock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            "/",
		RangeBegin:      256,
		RangeEnd:        0,
		CacheStatus: map[string]*ccmsg.ContentRequest_ClientCacheStatus{
			"\x00\x01\x02\x03\x04": &ccmsg.ContentRequest_ClientCacheStatus{
				BacklogDepth: 0x1,
			},
			"\x05\x06\x07\x08\x09": &ccmsg.ContentRequest_ClientCacheStatus{
				BacklogDepth: 0x1,
			},
		},
	}).Return(suite.newContentResponse(CROptions{bundles: 1, chunkSize: 256, objectSize: 1024}), nil).Once()
	mock.makeNewCacheCall(cl.l, "192.0.2.1:1001", "\x00\x01\x02\x03\x04")
	mock.makeNewCacheCall(cl.l, "192.0.2.2:1002", "\x05\x06\x07\x08\x09")

	queue := make(chan *fetchGroup, 128)
	bundleCompletions := make(chan BundleOutcome, 128)
	cl.schedule(context.Background(), "/", queue, bundleCompletions)

	group := <-queue
	assert.Nil(t, group.err)
	assert.NotNil(t, group.bundle)

	group = <-queue
	assert.Equal(t, "object chunk size changed mid retrieval", group.err.Error())
	assert.Nil(t, group.bundle)

	assert.Zero(t, len(queue))
}

func (suite *SchedulerTestSuite) TestSchedulerClientRetriesOneBundle() {
	t := suite.T()
	cl, mock := suite.newMock()

	mock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            "/",
		RangeBegin:      0,
		RangeEnd:        0,
		CacheStatus:     map[string]*ccmsg.ContentRequest_ClientCacheStatus{},
	}).Return(suite.newContentResponse(CROptions{bundles: 1}), nil).Once()
	mock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            "/",
		RangeBegin:      256,
		RangeEnd:        0,
		CacheStatus: map[string]*ccmsg.ContentRequest_ClientCacheStatus{
			"\x00\x01\x02\x03\x04": &ccmsg.ContentRequest_ClientCacheStatus{
				BacklogDepth: 0x1,
			},
			"\x05\x06\x07\x08\x09": &ccmsg.ContentRequest_ClientCacheStatus{
				BacklogDepth: 0x1,
			},
		},
	}).Return(suite.newContentResponse(CROptions{bundles: 1, bundleOffset: 1}), nil).Once()
	// This is the retried bundle
	mock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            "/",
		RangeBegin:      0,
		RangeEnd:        256,
		CacheStatus: map[string]*ccmsg.ContentRequest_ClientCacheStatus{
			"\x00\x01\x02\x03\x04": &ccmsg.ContentRequest_ClientCacheStatus{
				BacklogDepth: 0x2,
			},
			"\x05\x06\x07\x08\x09": &ccmsg.ContentRequest_ClientCacheStatus{
				BacklogDepth: 0x2,
			},
		},
	}).Return(suite.newContentResponse(CROptions{bundles: 1}), nil).Once()
	mock.makeNewCacheCall(cl.l, "192.0.2.1:1001", "\x00\x01\x02\x03\x04")
	mock.makeNewCacheCall(cl.l, "192.0.2.2:1002", "\x05\x06\x07\x08\x09")

	queue := make(chan *fetchGroup, 128)
	okGroup := func(chunkIdx uint64) *fetchGroup {
		group := <-queue
		assert.Nil(t, group.err)
		assert.NotNil(t, group.bundle)
		assert.Equal(t, chunkIdx, group.bundle.TicketRequest[0].ChunkIdx)
		return group
	}

	bundleCompletions := make(chan BundleOutcome, 128)
	go cl.schedule(context.Background(), "/", queue, bundleCompletions)
	// micro sleep to allow the scheduler read-ahead to read the second bundle
	// otherwise we have non-determinism in the test
	time.Sleep(time.Millisecond * 50)
	// Request a retry of bundle at offset 0
	okGroup(0)
	bundleCompletions <- BundleOutcome{Outcome: Retry, ChunkOffset: 0, Chunks: 2}
	// Defer the read-ahead bundle at offset 2
	group := okGroup(2)
	bundleCompletions <- BundleOutcome{Outcome: Deferred, ChunkOffset: 2, Chunks: 2, Bundle: group}
	// Receive the retried bundle at offset 0
	okGroup(0)
	// Receive the deferred bundle at offset 2
	okGroup(2)
	// acknowledge the bundles
	bundleCompletions <- BundleOutcome{Outcome: Completed, ChunkOffset: 0, Chunks: 2}
	bundleCompletions <- BundleOutcome{Outcome: Completed, ChunkOffset: 2, Chunks: 2}

	assert.Zero(t, len(queue))
	mock.AssertExpectations(t)
}

func (suite *SchedulerTestSuite) TestSchedulerClientDefersOneBundleBadly() {
	t := suite.T()
	cl, mock := suite.newMock()

	mock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            "/",
		RangeBegin:      0,
		RangeEnd:        0,
		CacheStatus:     map[string]*ccmsg.ContentRequest_ClientCacheStatus{},
	}).Return(suite.newContentResponse(CROptions{bundles: 1}), nil).Once()
	mock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            "/",
		RangeBegin:      256,
		RangeEnd:        0,
		CacheStatus: map[string]*ccmsg.ContentRequest_ClientCacheStatus{
			"\x00\x01\x02\x03\x04": &ccmsg.ContentRequest_ClientCacheStatus{
				BacklogDepth: 0x1,
			},
			"\x05\x06\x07\x08\x09": &ccmsg.ContentRequest_ClientCacheStatus{
				BacklogDepth: 0x1,
			},
		},
	}).Return(suite.newContentResponse(CROptions{bundles: 1}), nil).Once()
	mock.makeNewCacheCall(cl.l, "192.0.2.1:1001", "\x00\x01\x02\x03\x04")
	mock.makeNewCacheCall(cl.l, "192.0.2.2:1002", "\x05\x06\x07\x08\x09")

	queue := make(chan *fetchGroup, 128)
	bundleCompletions := make(chan BundleOutcome, 128)
	// Deferring without providing the fetch group is an error
	bundleCompletions <- BundleOutcome{Outcome: Deferred, ChunkOffset: 0, Chunks: 2}
	cl.schedule(context.Background(), "/", queue, bundleCompletions)

	group := <-queue
	assert.NotNil(t, group.err)
	assert.Nil(t, group.bundle)

	assert.Zero(t, len(queue))
}

func (suite *SchedulerTestSuite) TestSchedulerClientDefersBundles() {
	t := suite.T()
	cl, mock := suite.newMock()

	mock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            "/",
		RangeBegin:      0,
		RangeEnd:        0,
		CacheStatus: map[string]*ccmsg.ContentRequest_ClientCacheStatus{
			"\x00\x01\x02\x03\x04": &ccmsg.ContentRequest_ClientCacheStatus{},
		},
	}).Return(suite.newContentResponse(CROptions{bundles: 1}), nil).Once()
	mock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            "/",
		RangeBegin:      256,
		RangeEnd:        0,
		CacheStatus: map[string]*ccmsg.ContentRequest_ClientCacheStatus{
			"\x00\x01\x02\x03\x04": &ccmsg.ContentRequest_ClientCacheStatus{
				BacklogDepth: 0x1,
			},
			"\x05\x06\x07\x08\x09": &ccmsg.ContentRequest_ClientCacheStatus{
				BacklogDepth: 0x1,
			},
		},
	}).Return(suite.newContentResponse(CROptions{bundles: 1}), nil).Once()
	mock.makeNewCacheCall(cl.l, "192.0.2.1:1001", "\x00\x01\x02\x03\x04")
	mock.makeNewCacheCall(cl.l, "192.0.2.2:1002", "\x05\x06\x07\x08\x09")

	queue := make(chan *fetchGroup, 128)
	bundleCompletions := make(chan BundleOutcome, 128)
	fg1 := fetchGroup{bundle: &ccmsg.TicketBundle{
		TicketRequest: []*ccmsg.TicketRequest{&ccmsg.TicketRequest{ChunkIdx: 0}},
		CacheInfo:     []*ccmsg.CacheInfo{&ccmsg.CacheInfo{Pubkey: &ccmsg.PublicKey{PublicKey: []byte("\x00\x01\x02\x03\x04")}}},
	}}
	fg2 := fetchGroup{bundle: &ccmsg.TicketBundle{
		TicketRequest: []*ccmsg.TicketRequest{&ccmsg.TicketRequest{ChunkIdx: 2}},
		CacheInfo:     []*ccmsg.CacheInfo{&ccmsg.CacheInfo{Pubkey: &ccmsg.PublicKey{PublicKey: []byte("\x00\x01\x02\x03\x04")}}},
	}}
	fgs := []*fetchGroup{&fg1, &fg2}
	// defer all bundles - readahead will read all the bundles for this sample object
	bundleCompletions <- BundleOutcome{Outcome: Deferred, ChunkOffset: 0, Chunks: 2, Bundle: &fg1}
	bundleCompletions <- BundleOutcome{Outcome: Deferred, ChunkOffset: 2, Chunks: 2, Bundle: &fg2}
	// now acknowledge
	bundleCompletions <- BundleOutcome{Outcome: Completed, ChunkOffset: 0, Chunks: 2}
	bundleCompletions <- BundleOutcome{Outcome: Completed, ChunkOffset: 2, Chunks: 2}
	cl.cacheConns["\x00\x01\x02\x03\x04"] = newCacheMock("", ed25519.PublicKey("\x00\x01\x02\x03\x04"), [][]byte{})
	cl.schedule(context.Background(), "/", queue, bundleCompletions)

	assert.Equal(t, 4, len(queue))

	// This test may seem counter-intuitive, but it is an artifact of being a
	// close-surface unit test. the scheduler processes client outcomes first,
	// and deferrals are handled by immediate submission back into the channel
	// so unless we have an active client in the test - which we don't need
	// - the test code sees the deferrals first.
	// The deferred deliveries
	for idx, fg := range fgs {
		group := <-queue
		assert.Nil(t, group.err)
		assert.NotNil(t, group.bundle)
		assert.Equalf(t, group, fg, "Bad group %d", idx)
	}

	// The initial deliveries
	for i := 0; i < 2; i++ {
		group := <-queue
		assert.Nil(t, group.err)
		assert.NotNil(t, group.bundle)
		assert.NotContains(t, fgs, group)
	}

	assert.Zero(t, len(queue))
}
