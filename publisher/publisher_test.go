package publisher

import (
	"context"
	"net"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/cachecashproject/go-cachecash/catalog"
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/testutil"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/ed25519"
)

type PublisherTestSuite struct {
	suite.Suite

	l *logrus.Logger

	publisher *ContentPublisher

	chunkSize int

	clientPublic ed25519.PublicKey

	cachePublicKeys []ed25519.PublicKey
}

func TestPublisherTestSuite(t *testing.T) {
	suite.Run(t, new(PublisherTestSuite))
}

func (suite *PublisherTestSuite) SetupTest() {
	t := suite.T()

	l := logrus.New()
	suite.l = l

	var err error

	_, priv, err := ed25519.GenerateKey(nil) // TOOS: use faster, lower-quality entropy?
	if err != nil {
		t.Fatalf("failed to generate keypair: %v", err)
	}

	upstream, err := catalog.NewMockUpstream(l)
	if err != nil {
		t.Fatalf("failed to create mock upstream")
	}

	// Create a content object.
	suite.chunkSize = 128 * 1024
	upstream.Objects["/foo/bar"] = testutil.RandBytes(5 * suite.chunkSize)

	cat, err := catalog.NewCatalog(l, upstream)
	if err != nil {
		t.Fatalf("failed to create catalog")
	}

	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create stub database connection: %v", err)
	}

	innerMasterKeys := [][]byte{}
	cachePublicKeys := []ed25519.PublicKey{}

	for i := 0; i < 5; i++ {
		public, _, err := ed25519.GenerateKey(nil)
		if err != nil {
			t.Fatalf("failed to generate cache keypair: %v", err)
		}
		cachePublicKeys = append(cachePublicKeys, public)
		innerMasterKeys = append(innerMasterKeys, testutil.RandBytes(16))
	}

	escrows := sqlmock.NewRows([]string{"id"}).AddRow(0)
	sqlMock.ExpectQuery("^SELECT \\* FROM \"escrow\";").
		WithArgs().
		WillReturnRows(escrows)

	rows := sqlmock.NewRows([]string{"id", "escrow_id", "cache_id", "inner_master_key"}).
		AddRow(1, 0, 123, innerMasterKeys[0]).
		AddRow(2, 0, 124, innerMasterKeys[1]).
		AddRow(3, 0, 125, innerMasterKeys[2]).
		AddRow(4, 0, 126, innerMasterKeys[3]).
		AddRow(5, 0, 127, innerMasterKeys[4])
	sqlMock.ExpectQuery("^SELECT \"escrow_caches\"\\.\\* FROM \"escrow_caches\" WHERE \\(\"escrow_caches\"\\.\"escrow_id\"=\\$1\\) ORDER BY cache_id;").
		WithArgs(0).
		WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "public_key", "inetaddr", "port"}).
		AddRow(123, cachePublicKeys[0], net.ParseIP("127.0.0.1"), 9000).
		AddRow(124, cachePublicKeys[1], net.ParseIP("127.0.0.1"), 9001).
		AddRow(125, cachePublicKeys[2], net.ParseIP("127.0.0.1"), 9002).
		AddRow(126, cachePublicKeys[3], net.ParseIP("127.0.0.1"), 9003).
		AddRow(127, cachePublicKeys[3], net.ParseIP("127.0.0.1"), 9004)
	sqlMock.ExpectQuery("^SELECT \\* FROM \"cache\" WHERE \\(\"id\" IN \\(\\$1,\\$2,\\$3,\\$4\\,\\$5\\)\\);").
		WithArgs(123, 124, 125, 126, 127).
		WillReturnRows(rows)

	// XXX: Once we start using the catalog, passing nil is going to cause runtime panics.
	suite.publisher, err = NewContentPublisher(l, db, "", cat, priv)
	if err != nil {
		t.Fatalf("failed to construct publisher: %v", err)
	}

	_, err = suite.publisher.LoadFromDatabase(context.TODO())
	assert.Nil(t, err)
	suite.cachePublicKeys = cachePublicKeys
}

func (suite *PublisherTestSuite) TestContentRequest() {
	t := suite.T()

	_, err := suite.publisher.HandleContentRequest(context.TODO(), &ccmsg.ContentRequest{
		Path:            "/foo/bar",
		RangeBegin:      uint64(0),
		RangeEnd:        uint64(suite.chunkSize),
		ClientPublicKey: &ccmsg.PublicKey{PublicKey: suite.clientPublic},
	})
	assert.Nil(t, err, "failed to get bundle")
}

func (suite *PublisherTestSuite) TestContentRequestBeginBeyondObject() {
	t := suite.T()

	_, err := suite.publisher.HandleContentRequest(context.TODO(), &ccmsg.ContentRequest{
		Path:            "/foo/bar",
		RangeBegin:      uint64(300 * suite.chunkSize),
		RangeEnd:        uint64(301 * suite.chunkSize),
		ClientPublicKey: &ccmsg.PublicKey{PublicKey: suite.clientPublic},
	})
	assert.NotNil(t, err, "request should have caused an error but didn't")
}

func (suite *PublisherTestSuite) TestContentRequestEndBeyondObject() {
	t := suite.T()

	_, err := suite.publisher.HandleContentRequest(context.TODO(), &ccmsg.ContentRequest{
		Path:            "/foo/bar",
		RangeBegin:      uint64(1 * suite.chunkSize),
		RangeEnd:        uint64(100 * suite.chunkSize),
		ClientPublicKey: &ccmsg.PublicKey{PublicKey: suite.clientPublic},
	})
	assert.Nil(t, err, "failed to get bundle")
}

func (suite *PublisherTestSuite) TestContentRequestEndBeforeBegin() {
	t := suite.T()

	_, err := suite.publisher.HandleContentRequest(context.TODO(), &ccmsg.ContentRequest{
		Path:            "/foo/bar",
		RangeBegin:      uint64(3 * suite.chunkSize),
		RangeEnd:        uint64(1 * suite.chunkSize),
		ClientPublicKey: &ccmsg.PublicKey{PublicKey: suite.clientPublic},
	})
	assert.NotNil(t, err, "request should have caused an error but didn't")
}

func (suite *PublisherTestSuite) TestContentRequestEndWholeObject() {
	t := suite.T()

	_, err := suite.publisher.HandleContentRequest(context.TODO(), &ccmsg.ContentRequest{
		Path:            "/foo/bar",
		RangeBegin:      uint64(0),
		RangeEnd:        uint64(0),
		ClientPublicKey: &ccmsg.PublicKey{PublicKey: suite.clientPublic},
	})
	assert.Nil(t, err, "failed to get bundle")
}

func (suite *PublisherTestSuite) TestContentRequestTooManyFailedCaches() {
	t := suite.T()

	bundle, err := suite.publisher.HandleContentRequest(context.TODO(), &ccmsg.ContentRequest{
		Path:            "/foo/bar",
		RangeBegin:      uint64(0),
		RangeEnd:        uint64(suite.chunkSize),
		ClientPublicKey: &ccmsg.PublicKey{PublicKey: suite.clientPublic},
		CacheStatus: map[string]*ccmsg.ContentRequest_ClientCacheStatus{
			string(suite.cachePublicKeys[0]): &ccmsg.ContentRequest_ClientCacheStatus{
				Status: ccmsg.ContentRequest_ClientCacheStatus_UNUSABLE,
			},
			string(suite.cachePublicKeys[1]): &ccmsg.ContentRequest_ClientCacheStatus{
				Status: ccmsg.ContentRequest_ClientCacheStatus_UNUSABLE,
			},
		},
	})
	assert.NotNil(t, err, "got an impossible bundle")
	assert.Nil(t, bundle, "got an impossible bundle")
}

func (suite *PublisherTestSuite) TestContentRequestWithFailedCache() {
	t := suite.T()

	_, err := suite.publisher.HandleContentRequest(context.TODO(), &ccmsg.ContentRequest{
		Path:            "/foo/bar",
		RangeBegin:      uint64(0),
		RangeEnd:        uint64(suite.chunkSize),
		ClientPublicKey: &ccmsg.PublicKey{PublicKey: suite.clientPublic},
		CacheStatus: map[string]*ccmsg.ContentRequest_ClientCacheStatus{
			string(suite.cachePublicKeys[0]): &ccmsg.ContentRequest_ClientCacheStatus{
				Status: ccmsg.ContentRequest_ClientCacheStatus_UNUSABLE,
			},
		},
	})
	assert.Nil(t, err, "failed to get bundle")
}

// Unit test entirely mocked suite; PublisherTestSuite has larger tests that
// allow for actual ORM operation etc
type PublisherUnitTestSuite struct {
	suite.Suite

	l *logrus.Logger
}

func TestPublisherUnitTestSuite(t *testing.T) {
	suite.Run(t, new(PublisherUnitTestSuite))
}

func (suite *PublisherUnitTestSuite) SetupTest() {
	suite.l = logrus.New()
	suite.l.SetLevel(logrus.DebugLevel)
}

func (suite *PublisherUnitTestSuite) TestContentRequestFailedGetData() {
	t := suite.T()
	catalogMock := catalog.NewContentCatalogMock()
	p := &ContentPublisher{
		l:       suite.l,
		catalog: catalogMock,
	}
	ctx := context.TODO()
	request := &ccmsg.ContentRequest{
		Path:            "/foo/bar",
		RangeBegin:      uint64(0),
		RangeEnd:        uint64(51),
		ClientPublicKey: &ccmsg.PublicKey{PublicKey: []byte{}},
	}
	catalogMock.On("GetData", ctx, &ccmsg.ContentRequest{Path: request.Path}).Return((*catalog.ObjectMetadata)(nil), errors.New("no data"))

	_, err := p.HandleContentRequest(ctx, request)

	assert.Regexp(t, "failed to get metadata", err)
}

func (suite *PublisherUnitTestSuite) TestContentRequestBadRange() {
	t := suite.T()
	p := &ContentPublisher{
		l: suite.l,
	}
	ctx := context.TODO()
	request := &ccmsg.ContentRequest{
		Path:            "/foo/bar",
		RangeBegin:      uint64(51),
		RangeEnd:        uint64(50),
		ClientPublicKey: &ccmsg.PublicKey{PublicKey: []byte{}},
	}
	_, err := p.HandleContentRequest(ctx, request)

	assert.Regexp(t, "invalid range", err)
}

func (suite *PublisherUnitTestSuite) TestCachePermutations() {
	t := suite.T()
	p := &ContentPublisher{
		l:      suite.l,
		caches: make(map[string]*publisherCache),
	}

	getPermutation := func(key string, size uint) []uint64 {
		permutation, err := p.getCachePermutation(key, size)
		assert.Nil(t, err)
		return permutation
	}

	// Unknown caches result in errors in lookups
	_, err := p.getCachePermutation("cachepubkey", 1)
	assert.Regexp(t, "unknown cache", err)

	// Cached results are not caculated at all
	p.caches["cached"] = &publisherCache{permutations: map[int][]uint64{131: []uint64{0, 1, 2, 3, 4}}}
	assert.Equal(t, []uint64{0, 1, 2, 3, 4}, getPermutation("cached", 1))
	// Basic shape
	p.caches["cachepubkey"] = &publisherCache{permutations: map[int][]uint64{}}
	assert.Len(t, getPermutation("cachepubkey", 2), 257)
	// After doing a lookup, the permutation is cached
	assert.Len(t, p.caches["cachepubkey"].permutations[257], 257)
	// More caches require longer permutations
	assert.Len(t, getPermutation("cachepubkey", 4), 521)
	assert.Len(t, getPermutation("cachepubkey", 100), 8209)
	// Because we need a lookup table of primes, we cannot handle arbitrary length sizes.
	_, err = p.getCachePermutation("cachepubkey", 101)
	assert.Regexp(t, "more primes needed", err)
	// One literal for validation
	assert.Equal(t,
		[]uint64{
			0x10, 0x23, 0x36, 0x49, 0x5c, 0x6f, 0x82, 0x12, 0x25, 0x38, 0x4b,
			0x5e, 0x71, 0x1, 0x14, 0x27, 0x3a, 0x4d, 0x60, 0x73, 0x3, 0x16,
			0x29, 0x3c, 0x4f, 0x62, 0x75, 0x5, 0x18, 0x2b, 0x3e, 0x51, 0x64,
			0x77, 0x7, 0x1a, 0x2d, 0x40, 0x53, 0x66, 0x79, 0x9, 0x1c, 0x2f,
			0x42, 0x55, 0x68, 0x7b, 0xb, 0x1e, 0x31, 0x44, 0x57, 0x6a, 0x7d,
			0xd, 0x20, 0x33, 0x46, 0x59, 0x6c, 0x7f, 0xf, 0x22, 0x35, 0x48,
			0x5b, 0x6e, 0x81, 0x11, 0x24, 0x37, 0x4a, 0x5d, 0x70, 0x0, 0x13,
			0x26, 0x39, 0x4c, 0x5f, 0x72, 0x2, 0x15, 0x28, 0x3b, 0x4e, 0x61,
			0x74, 0x4, 0x17, 0x2a, 0x3d, 0x50, 0x63, 0x76, 0x6, 0x19, 0x2c,
			0x3f, 0x52, 0x65, 0x78, 0x8, 0x1b, 0x2e, 0x41, 0x54, 0x67, 0x7a,
			0xa, 0x1d, 0x30, 0x43, 0x56, 0x69, 0x7c, 0xc, 0x1f, 0x32, 0x45,
			0x58, 0x6b, 0x7e, 0xe, 0x21, 0x34, 0x47, 0x5a, 0x6d, 0x80},
		getPermutation("cachepubkey", 1))
	// Different public keys -> different permutations
	p.caches["othercache"] = &publisherCache{permutations: map[int][]uint64{}}
	assert.NotEqual(t,
		getPermutation("othercache", 2),
		getPermutation("cachepubkey", 2))
	// Different numbers of caches in the escrow within a size threshold
	// -> same permutation
	assert.Equal(t,
		getPermutation("cachepubkey", 2), // 150 -> 256
		getPermutation("cachepubkey", 3)) // 225 -> 256
	// Different numbers of caches in the escrow that cross a size threshold
	// -> different permutations
	assert.NotEqual(t,
		getPermutation("cachepubkey", 2), // 150 -> 256
		getPermutation("cachepubkey", 4)) // 300 -> 512
}
