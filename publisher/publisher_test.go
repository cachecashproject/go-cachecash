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
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/ed25519"
)

type PublisherTestSuite struct {
	suite.Suite

	l *logrus.Logger

	publisher *ContentPublisher
	escrow    *Escrow

	blockSize int

	clientPublic  ed25519.PublicKey
	clientPrivate ed25519.PrivateKey
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
	suite.blockSize = 128 * 1024
	upstream.Objects["/foo/bar"] = testutil.RandBytes(5 * suite.blockSize)

	cat, err := catalog.NewCatalog(l, upstream)
	if err != nil {
		t.Fatalf("failed to create catalog")
	}

	db, mock, err := sqlmock.New()
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
	mock.ExpectQuery("^SELECT \\* FROM \"escrow_caches\" WHERE \\(escrow_id = \\$1\\);").
		WithArgs(0).
		WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "public_key", "inetaddr", "port"}).
		AddRow(123, cachePublicKeys[0], net.ParseIP("127.0.0.1"), 9000).
		AddRow(124, cachePublicKeys[1], net.ParseIP("127.0.0.1"), 9001).
		AddRow(125, cachePublicKeys[2], net.ParseIP("127.0.0.1"), 9002).
		AddRow(126, cachePublicKeys[3], net.ParseIP("127.0.0.1"), 9003)
	mock.ExpectQuery("^SELECT \\* FROM \"cache\" WHERE \\(\"id\" IN \\(\\$1,\\$2,\\$3,\\$4\\)\\);").
		WithArgs(123, 124, 125, 126).
		WillReturnRows(rows)

	// XXX: Once we start using the catalog, passing nil is going to cause runtime panics.
	suite.publisher, err = NewContentPublisher(l, db, cat, priv)
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
		RangeEnd:        uint64(suite.blockSize),
		ClientPublicKey: &ccmsg.PublicKey{PublicKey: suite.clientPublic},
	})
	assert.Nil(t, err, "failed to get bundle")
}

func (suite *PublisherTestSuite) TestContentRequestBeginBeyondObject() {
	t := suite.T()

	_, err := suite.publisher.HandleContentRequest(context.TODO(), &ccmsg.ContentRequest{
		Path:            "/foo/bar",
		RangeBegin:      uint64(300 * suite.blockSize),
		RangeEnd:        uint64(301 * suite.blockSize),
		ClientPublicKey: &ccmsg.PublicKey{PublicKey: suite.clientPublic},
	})
	assert.NotNil(t, err, "request should have caused an error but didn't")
}

func (suite *PublisherTestSuite) TestContentRequestEndBeyondObject() {
	t := suite.T()

	_, err := suite.publisher.HandleContentRequest(context.TODO(), &ccmsg.ContentRequest{
		Path:            "/foo/bar",
		RangeBegin:      uint64(1 * suite.blockSize),
		RangeEnd:        uint64(100 * suite.blockSize),
		ClientPublicKey: &ccmsg.PublicKey{PublicKey: suite.clientPublic},
	})
	assert.Nil(t, err, "failed to get bundle")
}

func (suite *PublisherTestSuite) TestContentRequestEndBeforeBegin() {
	t := suite.T()

	_, err := suite.publisher.HandleContentRequest(context.TODO(), &ccmsg.ContentRequest{
		Path:            "/foo/bar",
		RangeBegin:      uint64(3 * suite.blockSize),
		RangeEnd:        uint64(1 * suite.blockSize),
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
