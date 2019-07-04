package client

import (
	"context"
	"net"
	"testing"

	"golang.org/x/crypto/ed25519"

	cachecash "github.com/cachecashproject/go-cachecash"
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/colocationpuzzle"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

type ClientTestSuite struct {
	suite.Suite

	l *logrus.Logger
}

func (suite *ClientTestSuite) SetupTest() {
	l := logrus.New()
	suite.l = l
}

func (suite *ClientTestSuite) newMock() (*client, *publisherMock) {
	mock := newPublisherMock()
	return &client{
		l:             suite.l,
		publisherConn: mock,
		cacheConns:    map[cacheID]cacheConnection{},
	}, mock
}

func (suite *ClientTestSuite) encryptTicketL2(secret []byte, l2 *ccmsg.TicketL2) []byte {
	encrypted, err := common.EncryptTicketL2(&colocationpuzzle.Puzzle{
		Secret: secret,
	}, l2)
	if err != nil {
		panic(err)
	}
	return encrypted
}

func (suite *ClientTestSuite) newContentResponse(n uint64, chunkOffset uint64, objectSize uint64, chunkSize uint64) *ccmsg.ContentResponse {
	bundles := []*ccmsg.TicketBundle{}

	for i := uint64(0); i < n; i++ {
		chunkIdx := uint64(i*4) + chunkOffset*4

		reqs := []*ccmsg.TicketRequest{}
		for j := uint64(0); j < 4; j++ {
			reqs = append(reqs, &ccmsg.TicketRequest{
				ChunkIdx: chunkIdx + j,
				InnerKey: &ccmsg.BlockKey{
					Key: []byte{},
				},
				CachePublicKey: &ccmsg.PublicKey{
					PublicKey: []byte{},
				},
			})
		}

		bundles = append(bundles, &ccmsg.TicketBundle{
			Remainder: &ccmsg.TicketBundleRemainder{
				PuzzleInfo: &ccmsg.ColocationPuzzleInfo{
					Goal:        []byte{0x11, 0x88, 0x53, 0xf9, 0x6d, 0xc6, 0x70, 0xe9, 0xd6, 0x6a, 0xab, 0xee, 0xf3, 0x4a, 0xed, 0x53, 0x5d, 0x2, 0xd2, 0xa9, 0x2b, 0xf0, 0xe0, 0x80, 0x9e, 0xc9, 0xb3, 0x12, 0xcd, 0xa0, 0x83, 0xfc, 0x5a, 0x5c, 0x94, 0x7c, 0xef, 0xba, 0xd7, 0x68, 0xe2, 0x3f, 0x64, 0xef, 0xd8, 0x8, 0x87, 0x20},
					Rounds:      2,
					StartOffset: 2,
				},
			},
			TicketRequest: reqs,
			EncryptedTicketL2: suite.encryptTicketL2(
				[]byte{0x58, 0x72, 0x17, 0xdd, 0x1e, 0xfd, 0x61, 0x12, 0xb2, 0xb5, 0xb6, 0x41, 0xd2, 0x7a, 0xa5, 0xfd, 0x47, 0x2f, 0x27, 0xb6, 0x8f, 0x19, 0x4b, 0x8c, 0x2f, 0x9, 0x2, 0x9e, 0xdb, 0x63, 0xca, 0x5f, 0x2b, 0xf4, 0xd0, 0x91, 0x6b, 0xbc, 0x26, 0xa2, 0x92, 0x92, 0xe3, 0x11, 0xae, 0x5a, 0xb5, 0x18},
				&ccmsg.TicketL2{
					InnerSessionKey: []*ccmsg.BlockKey{
						{
							Key: []byte{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18},
						},
						{
							Key: []byte{19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34},
						},
						{
							Key: []byte{35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50},
						},
						{
							Key: []byte{50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65},
						},
					},
				}),
			CacheInfo: []*ccmsg.CacheInfo{
				suite.newCache(net.ParseIP("192.0.2.1"), 1001, []byte{0, 1, 2, 3, 4}),
				suite.newCache(net.ParseIP("192.0.2.2"), 1002, []byte{5, 6, 7, 8, 9}),
				suite.newCache(net.ParseIP("192.0.2.3"), 1003, []byte{10, 11, 12, 13, 14}),
				suite.newCache(net.ParseIP("192.0.2.4"), 1004, []byte{15, 0, 1, 1, 3}),
			},
			Metadata: &ccmsg.ObjectMetadata{
				ObjectSize: objectSize,
				ChunkSize:  chunkSize,
			},
		})
	}
	return &ccmsg.ContentResponse{
		Bundles: bundles,
	}
}

func (suite *ClientTestSuite) newCache(ip net.IP, port uint32, pubkey []byte) *ccmsg.CacheInfo {
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

func (suite *ClientTestSuite) TestGetObject() {
	t := suite.T()
	cl, pubMock := suite.newMock()
	ctx := context.Background()
	path := "/"

	pubMock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            path,
		RangeBegin:      0,
		RangeEnd:        0,
		BacklogDepth:    map[string]uint64{},
	}).Return(suite.newContentResponse(1, 0, 512, 128), nil).Once()
	pubMock.makeNewCacheCall(cl.l, "192.0.2.1:1001", "\x00\x01\x02\x03\x04",
		[]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15})
	pubMock.makeNewCacheCall(cl.l, "192.0.2.2:1002", "\x05\x06\x07\x08\x09",
		[]byte{16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31})
	pubMock.makeNewCacheCall(cl.l, "192.0.2.3:1003", "\x0a\x0b\x0c\x0d\x0e",
		[]byte{32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47})
	pubMock.makeNewCacheCall(cl.l, "192.0.2.4:1004", "\x0f\x00\x01\x01\x03",
		[]byte{48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63})

	output := make(chan *OutputChunk, 128)
	cl.GetObject(ctx, path, output)

	assert.Equal(t, &OutputChunk{
		Data: []byte{0x16, 0xbd, 0xef, 0x2b, 0xdc, 0xf2, 0x40, 0xb2, 0x6d, 0x99, 0xda, 0x3d, 0xe8, 0x42, 0x61, 0x20},
		Err:  nil,
	}, <-output)
	assert.Equal(t, &OutputChunk{
		Data: []byte{0xd0, 0xd7, 0xe6, 0xd5, 0xa8, 0xe3, 0x78, 0xa7, 0xed, 0x9e, 0x89, 0x32, 0x91, 0xb2, 0xb3, 0x85},
		Err:  nil,
	}, <-output)
	assert.Equal(t, &OutputChunk{
		Data: []byte{0xb7, 0x7e, 0xe, 0xf8, 0x38, 0x38, 0x65, 0x73, 0xec, 0x8c, 0x1, 0xee, 0x9c, 0xfa, 0xb3, 0x1},
		Err:  nil,
	}, <-output)
	assert.Equal(t, &OutputChunk{
		Data: []byte{0xf6, 0x53, 0x15, 0x9f, 0xbc, 0xf2, 0x5c, 0xd1, 0xdc, 0xf3, 0x73, 0x33, 0x79, 0x6e, 0xf5, 0x1b},
		Err:  nil,
	}, <-output)
	assert.Equal(t, 0, len(output))

	err := cl.Close(ctx)
	assert.Nil(t, err)
	pubMock.AssertExpectations(t)

	path = "/path2"
	// Get a second object with different shape, demonstrating that clients can be reused.
	pubMock.On("GetContent", &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(cl.publicKey),
		Path:            path,
		RangeBegin:      0,
		RangeEnd:        0,
		BacklogDepth:    map[string]uint64{},
	}).Return(suite.newContentResponse(1, 0, 256, 32), nil).Once()
	pubMock.On("GetContent", mock.MatchedBy(func(request *ccmsg.ContentRequest) bool {
		return (path == request.Path && request.RangeBegin == 0x80 && request.RangeEnd == 0)
		// We ignore the backlog completely: for this functional test it is irrelevant.
	})).Return(suite.newContentResponse(1, 1, 256, 32), nil).Once()
	pubMock.makeNewCacheCall(cl.l, "192.0.2.1:1001", "\x00\x01\x02\x03\x04",
		[]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		[]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15})
	pubMock.makeNewCacheCall(cl.l, "192.0.2.2:1002", "\x05\x06\x07\x08\x09",
		[]byte{16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
		[]byte{16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31})
	pubMock.makeNewCacheCall(cl.l, "192.0.2.3:1003", "\x0a\x0b\x0c\x0d\x0e",
		[]byte{32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47},
		[]byte{32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47})
	pubMock.makeNewCacheCall(cl.l, "192.0.2.4:1004", "\x0f\x00\x01\x01\x03",
		[]byte{48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63},
		[]byte{48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63})

	output = make(chan *OutputChunk, 128)
	go cl.GetObject(ctx, "/path2", output)
	assert.Equal(t, &OutputChunk{
		Data: []byte{0x16, 0xbd, 0xef, 0x2b, 0xdc, 0xf2, 0x40, 0xb2, 0x6d, 0x99, 0xda, 0x3d, 0xe8, 0x42, 0x61, 0x20},
		Err:  nil,
	}, <-output)
	assert.Equal(t, &OutputChunk{
		Data: []byte{0xd0, 0xd7, 0xe6, 0xd5, 0xa8, 0xe3, 0x78, 0xa7, 0xed, 0x9e, 0x89, 0x32, 0x91, 0xb2, 0xb3, 0x85},
		Err:  nil,
	}, <-output)
	assert.Equal(t, &OutputChunk{
		Data: []byte{0xb7, 0x7e, 0xe, 0xf8, 0x38, 0x38, 0x65, 0x73, 0xec, 0x8c, 0x1, 0xee, 0x9c, 0xfa, 0xb3, 0x1},
		Err:  nil,
	}, <-output)
	assert.Equal(t, &OutputChunk{
		Data: []byte{0xf6, 0x53, 0x15, 0x9f, 0xbc, 0xf2, 0x5c, 0xd1, 0xdc, 0xf3, 0x73, 0x33, 0x79, 0x6e, 0xf5, 0x1b},
		Err:  nil,
	}, <-output)
	assert.Equal(t, &OutputChunk{
		Data: []byte{0x5a, 0x35, 0x78, 0x76, 0x46, 0xef, 0x0, 0x6f, 0x2c, 0xe8, 0x21, 0x54, 0x24, 0xe0, 0xd6, 0x86},
		Err:  nil,
	}, <-output)
	assert.Equal(t, &OutputChunk{
		Data: []byte{0xc8, 0xe4, 0x1e, 0xe9, 0x58, 0x8, 0x96, 0xde, 0x29, 0xd3, 0x18, 0xc0, 0x5d, 0xef, 0x9, 0xd5},
		Err:  nil,
	}, <-output)
	assert.Equal(t, &OutputChunk{
		Data: []byte{0x95, 0xc4, 0xcf, 0x26, 0x56, 0xe7, 0x76, 0x8a, 0x2b, 0xbf, 0x8f, 0x73, 0x89, 0x9a, 0x29, 0xff},
		Err:  nil,
	}, <-output)
	assert.Equal(t, &OutputChunk{
		Data: []byte{0x74, 0xbc, 0x7a, 0x4, 0x55, 0x81, 0xe3, 0x49, 0x31, 0xe, 0xa4, 0x8, 0xf4, 0xb8, 0x67, 0x5c},
		Err:  nil,
	}, <-output)
	assert.Equal(t, 0, len(output))

	err = cl.Close(ctx)
	assert.Nil(t, err)
	pubMock.AssertExpectations(t)
}

func (suite *ClientTestSuite) TestGetCacheConnection() {
	t := suite.T()
	cl, mock := suite.newMock()
	ctx := context.Background()

	// A cache can be connected to and looked up again on a different IP
	mock.makeNewCacheCall(cl.l, "192.0.2.1:1001", "\x00\x01\x02\x03\x04")
	k1 := ed25519.PublicKey(([]byte)("\x00\x01\x02\x03\x04"))
	con1, err := cl.GetCacheConnection(ctx, "192.0.2.1:1001", k1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(cl.cacheConns))
	con1prime, err := cl.GetCacheConnection(ctx, "ignored", k1)
	assert.Nil(t, err)
	assert.Equal(t, con1, con1prime)

	// A second cache with a different key is stored separately
	mock.makeNewCacheCall(cl.l, "192.0.2.2:1002", "\x05\x06\x07\x08\x09")
	k2 := ed25519.PublicKey(([]byte)("\x05\x06\x07\x08\x09"))
	con2, err := cl.GetCacheConnection(ctx, "192.0.2.2:1002", k2)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(cl.cacheConns))
	con1, err = cl.GetCacheConnection(ctx, "ignored", k1)
	assert.Nil(t, err)
	assert.Equal(t, con1, con1prime)
	assert.NotEqual(t, con1, con2)
}
