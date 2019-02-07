package publisher

import (
	"crypto/aes"
	"net"
	"testing"

	cachecash "github.com/cachecashproject/go-cachecash"
	"github.com/cachecashproject/go-cachecash/batchsignature"
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/ed25519"
)

type TicketBundleTestSuite struct {
	suite.Suite

	l *logrus.Logger

	publisher *ContentPublisher
	escrow   *Escrow

	clientPublic  ed25519.PublicKey
	clientPrivate ed25519.PrivateKey
}

func TestTicketBundleTestSuite(t *testing.T) {
	suite.Run(t, new(TicketBundleTestSuite))
}

func (suite *TicketBundleTestSuite) SetupTest() {
	t := suite.T()

	l := logrus.New()
	suite.l = l

	var err error

	_, priv, err := ed25519.GenerateKey(nil) // TOOS: use faster, lower-quality entropy?
	if err != nil {
		t.Fatalf("failed to generate keypair: %v", err)
	}
	// XXX: Once we start using the catalog, passing nil is going to cause runtime panics.
	suite.publisher, err = NewContentPublisher(l, nil, priv)
	if err != nil {
		t.Fatalf("failed to construct publisher: %v", err)
	}

	ei := &ccmsg.EscrowInfo{
		Id:              testutil.RandBytes(common.EscrowIDSize),
		StartBlock:      42,
		DrawDelay:       5,
		ExpirationDelay: 5,
		TicketsPerBlock: []*ccmsg.Segment{
			{Length: 3, Value: 100},
		},
	}
	suite.escrow, err = suite.publisher.NewEscrow(ei)
	if err != nil {
		t.Fatalf("failed to construct escrow: %v", err)
	}

	suite.clientPublic, suite.clientPrivate, err = ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate client keypair: %v", err)
	}
}

func (suite *TicketBundleTestSuite) generateCacheInfo() *ParticipatingCache {
	t := suite.T()

	pub, _, err := ed25519.GenerateKey(nil) // TOOS: use faster, lower-quality entropy?
	if err != nil {
		t.Fatalf("failed to generate cache keypair: %v", err)
	}

	return &ParticipatingCache{
		InnerMasterKey: testutil.RandBytes(16), // XXX: ??
		PublicKey:      pub,
		Inetaddr:       net.ParseIP("10.0.0.1"),
		Port:           9999,
	}
}

func (suite *TicketBundleTestSuite) TestGenerateTicketBundle() {
	t := suite.T()

	caches := []*ParticipatingCache{
		suite.generateCacheInfo(),
		suite.generateCacheInfo(),
	}

	obj, err := cachecash.RandomContentBuffer(4, aes.BlockSize*50)
	if err != nil {
		t.Fatalf("failed to construct random test data: %v", err)
	}

	objectID, err := common.BytesToObjectID(testutil.RandBytes(common.ObjectIDSize))
	if err != nil {
		t.Fatalf("failed to generate object ID: %v", err)
	}

	bp := &BundleParams{
		ClientPublicKey:   suite.clientPublic,
		RequestSequenceNo: 123,
		Escrow:            suite.escrow,
		Object:            obj,
		ObjectID:          objectID,
		Entries: []BundleEntryParams{
			{TicketNo: 0, BlockIdx: 0, Cache: caches[0]},
			{TicketNo: 1, BlockIdx: 1, Cache: caches[1]},
		},
	}

	batchSigner, err := batchsignature.NewTrivialBatchSigner(suite.escrow.PrivateKey.(ed25519.PrivateKey))
	if err != nil {
		t.Fatalf("failed to construct batch signer: %v", err)
	}

	gen := NewBundleGenerator(suite.l, batchSigner)
	bundle, err := gen.GenerateTicketBundle(bp)
	assert.Nil(t, err, "failed to generate ticket bundle")

	// TODO: more!
	_ = bundle
}

// TODO: Need to add regression test specifically testing that block IDs are assigned correctly.  Had bug where all
//   blocks were given identical block IDs (that of the last block).  This was only caught by the integration tests.
