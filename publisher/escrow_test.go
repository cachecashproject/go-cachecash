package publisher

import (
	"crypto/aes"
	"net"
	"testing"

	"github.com/cachecashproject/go-cachecash/batchsignature"
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/publisher/models"
	"github.com/cachecashproject/go-cachecash/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/ed25519"
)

type EscrowTestSuite struct {
	suite.Suite

	l *logrus.Logger

	publisher *ContentPublisher
	escrow    *Escrow

	clientPublic  ed25519.PublicKey
	clientPrivate ed25519.PrivateKey
}

func TestTicketBundleTestSuite(t *testing.T) {
	suite.Run(t, new(EscrowTestSuite))
}

func (suite *EscrowTestSuite) SetupTest() {
	t := suite.T()

	l := logrus.New()
	suite.l = l

	var err error

	_, priv, err := ed25519.GenerateKey(nil) // TOOS: use faster, lower-quality entropy?
	if err != nil {
		t.Fatalf("failed to generate keypair: %v", err)
	}
	// XXX: Once we start using the catalog, passing nil is going to cause runtime panics.
	suite.publisher, err = NewContentPublisher(l, nil, "", nil, priv)
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

func (suite *EscrowTestSuite) generateCacheInfo() ParticipatingCache {
	t := suite.T()

	pub, _, err := ed25519.GenerateKey(nil) // TOOS: use faster, lower-quality entropy?
	if err != nil {
		t.Fatalf("failed to generate cache keypair: %v", err)
	}

	return ParticipatingCache{
		InnerMasterKey: testutil.RandBytes(16), // XXX: ??
		Cache: models.Cache{
			PublicKey: pub,
			Inetaddr:  net.ParseIP("10.0.0.1"),
			Port:      9999,
		},
	}
}

func (suite *EscrowTestSuite) TestGenerateTicketBundle() {
	t := suite.T()

	const chunkCount = 2

	plaintextChunks := make([][]byte, 0, chunkCount)
	caches := make([]ParticipatingCache, 0, chunkCount)
	for i := 0; i < chunkCount; i++ {
		plaintextChunks = append(plaintextChunks, testutil.RandBytes(aes.BlockSize*50))
		caches = append(caches, suite.generateCacheInfo())
	}

	objectID, err := common.BytesToObjectID(testutil.RandBytes(common.ObjectIDSize))
	if err != nil {
		t.Fatalf("failed to generate object ID: %v", err)
	}

	bp := &BundleParams{
		ClientPublicKey:   suite.clientPublic,
		RequestSequenceNo: 123,
		Escrow:            suite.escrow,
		ObjectID:          objectID,
		Entries: []BundleEntryParams{
			{TicketNo: 0, ChunkIdx: 0, Cache: caches[0]},
			{TicketNo: 1, ChunkIdx: 1, Cache: caches[1]},
		},
		PlaintextChunks: plaintextChunks,
	}

	batchSigner, err := batchsignature.NewTrivialBatchSigner(suite.escrow.Inner.PrivateKey)
	if err != nil {
		t.Fatalf("failed to construct batch signer: %v", err)
	}

	gen := NewBundleGenerator(suite.l, batchSigner)
	bundle, err := gen.GenerateTicketBundle(bp)
	assert.Nil(t, err, "failed to generate ticket bundle")

	// TODO: more!
	_ = bundle
}

// TODO: Need to add regression test specifically testing that chunk IDs are assigned correctly.  Had bug where all
//   chunks were given identical chunk IDs (that of the last chunk).  This was only caught by the integration tests.

func (suite *EscrowTestSuite) TestCalculateLookup() {
	t := suite.T()
	e := suite.escrow
	p := e.Publisher
	caches := make([]*ParticipatingCache, 0, 4)
	for i := 0; i < 4; i++ {
		cache := &ParticipatingCache{
			Cache: models.Cache{
				PublicKey: ed25519.PublicKey("key " + string(i)),
			},
		}
		caches = append(caches, cache)
		key := string(cache.PublicKey())
		p.caches[key] = &publisherCache{
			permutations:  make(map[int][]uint64),
			participation: cache}
	}

	e.Caches = caches
	assert.Nil(t, e.CalculateLookup())
	assert.Equal(t, []int{
		2, 3, 2, 3, 0, 0, 1, 3, 3, 0, 3, 0, 1, 1, 1, 2, 3, 2, 3, 2, 2, 2, 2,
		0, 3, 3, 1, 1, 1, 1, 0, 0, 0, 0, 2, 2, 2, 2, 0, 3, 0, 3, 0, 3, 1, 1,
		3, 2, 2, 2, 0, 0, 2, 1, 3, 0, 3, 0, 3, 0, 3, 0, 2, 2, 2, 2, 2, 2, 2,
		1, 0, 3, 1, 3, 1, 3, 1, 1, 0, 3, 2, 2, 2, 2, 1, 1, 3, 0, 0, 1, 3, 1,
		1, 2, 3, 2, 3, 2, 0, 1, 1, 3, 0, 3, 0, 0, 0, 0, 1, 3, 2, 3, 2, 3, 2,
		3, 1, 3, 3, 2, 3, 1, 3, 1, 1, 0, 0, 2, 2, 2, 3, 1, 3, 0, 0, 0, 2, 3,
		1, 1, 1, 2, 2, 2, 2, 3, 1, 3, 1, 3, 0, 3, 0, 0, 0, 1, 1, 2, 2, 2, 2,
		0, 0, 0, 1, 3, 3, 3, 3, 1, 1, 1, 0, 2, 2, 2, 1, 3, 1, 1, 0, 0, 0, 3,
		1, 3, 1, 3, 2, 2, 2, 0, 3, 1, 1, 1, 3, 0, 3, 0, 0, 1, 3, 1, 3, 2, 2,
		0, 0, 0, 1, 1, 2, 3, 2, 3, 0, 3, 0, 0, 2, 2, 3, 3, 1, 1, 1, 1, 0, 2,
		3, 2, 3, 2, 1, 1, 2, 0, 3, 1, 3, 1, 1, 3, 0, 0, 0, 0, 2, 2, 1, 2, 2,
		3, 0, 3, 0, 3, 1, 3, 3, 2, 2, 0, 0, 0, 1, 2, 3, 0, 3, 0, 3, 1, 1, 0,
		3, 2, 2, 2, 2, 2, 2, 0, 0, 3, 1, 3, 1, 1, 1, 3, 0, 3, 2, 2, 2, 2, 2,
		0, 0, 0, 0, 1, 3, 1, 1, 2, 3, 2, 3, 0, 3, 2, 1, 3, 0, 3, 0, 0, 0, 0,
		2, 3, 2, 3, 2, 2, 2, 3, 0, 0, 3, 1, 3, 1, 1, 1, 0, 0, 0, 2, 2, 2, 3,
		1, 3, 0, 3, 0, 1, 1, 1, 1, 2, 2, 2, 2, 0, 2, 1, 1, 3, 3, 0, 3, 0, 0,
		0, 1, 2, 2, 2, 2, 2, 0, 0, 1, 3, 3, 3, 1, 3, 1, 3, 1, 0, 2, 2, 2, 1,
		1, 1, 3, 0, 0, 2, 3, 1, 3, 1, 3, 2, 2, 2, 2, 1, 1, 1, 0, 3, 0, 3, 0,
		0, 1, 1, 2, 3, 2, 3, 0, 0, 0, 1, 3, 2, 3, 0, 3, 1, 1, 1, 0, 2, 2, 1,
		3, 1, 1, 1, 0, 0, 2, 3, 1, 1, 1, 1, 2, 2, 2, 3, 1, 3, 1, 1, 0, 0, 0,
		0, 0, 1, 1, 2, 2, 2, 3, 0, 0, 0, 3, 3, 3, 3, 2, 0, 0, 0, 0, 2, 2, 3,
		1, 3, 1, 3, 1, 0, 0, 3, 2, 3, 2, 1, 1, 2, 2, 0, 3, 1, 3, 1, 1, 0, 3,
		0, 0, 2, 3, 1, 1, 2, 0, 0, 0, 0, 1, 1, 1, 3}, *e.lookup)
}

func (suite *EscrowTestSuite) TestCalculateLookupNoCaches() {
	t := suite.T()
	e := suite.escrow
	caches := make([]*ParticipatingCache, 0)
	e.Caches = caches
	assert.NotNil(t, e.CalculateLookup())
}
