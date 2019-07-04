package publisher

import (
	"context"
	"net"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/cachecashproject/go-cachecash/catalog"
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/publisher/models"
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

	for i := 0; i < 4; i++ {
		public, _, err := ed25519.GenerateKey(nil)
		if err != nil {
			t.Fatalf("failed to generate cache keypair: %v", err)
		}
		cachePublicKeys = append(cachePublicKeys, public)
		innerMasterKeys = append(innerMasterKeys, testutil.RandBytes(16))
	}

	rows := sqlmock.NewRows([]string{"id", "escrow_id", "cache_id", "inner_master_key"}).
		AddRow(1, 0, 123, innerMasterKeys[0]).
		AddRow(2, 0, 124, innerMasterKeys[1]).
		AddRow(3, 0, 125, innerMasterKeys[2]).
		AddRow(4, 0, 126, innerMasterKeys[3])
	sqlMock.ExpectQuery("^SELECT \\* FROM \"escrow_caches\" WHERE \\(escrow_id = \\$1\\);").
		WithArgs(0).
		WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "public_key", "inetaddr", "port"}).
		AddRow(123, cachePublicKeys[0], net.ParseIP("127.0.0.1"), 9000).
		AddRow(124, cachePublicKeys[1], net.ParseIP("127.0.0.1"), 9001).
		AddRow(125, cachePublicKeys[2], net.ParseIP("127.0.0.1"), 9002).
		AddRow(126, cachePublicKeys[3], net.ParseIP("127.0.0.1"), 9003)
	sqlMock.ExpectQuery("^SELECT \\* FROM \"cache\" WHERE \\(\"id\" IN \\(\\$1,\\$2,\\$3,\\$4\\)\\);").
		WithArgs(123, 124, 125, 126).
		WillReturnRows(rows)

	// XXX: Once we start using the catalog, passing nil is going to cause runtime panics.
	suite.publisher, err = NewContentPublisher(l, db, "", cat, priv)
	if err != nil {
		t.Fatalf("failed to construct publisher: %v", err)
	}

	escrow := Escrow{
		Inner: models.Escrow{},
	}

	suite.publisher.escrows = append(suite.publisher.escrows, &escrow)
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
