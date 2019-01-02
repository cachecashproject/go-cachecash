package provider

import (
	"crypto/aes"
	"net"
	"testing"

	cachecash "github.com/kelleyk/go-cachecash"
	"github.com/kelleyk/go-cachecash/batchsignature"
	"github.com/kelleyk/go-cachecash/ccmsg"
	"github.com/kelleyk/go-cachecash/testutil"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/ed25519"
)

type TicketBundleTestSuite struct {
	suite.Suite

	provider *ContentProvider
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

	var err error

	_, priv, err := ed25519.GenerateKey(nil) // TOOS: use faster, lower-quality entropy?
	if err != nil {
		panic(errors.Wrap(err, "failed to generate keypair"))
	}
	suite.provider, err = NewContentProvider(l, priv)
	if err != nil {
		panic(errors.Wrap(err, "failed to construct provider"))
	}

	ei := &ccmsg.EscrowInfo{
		StartBlock:      42,
		DrawDelay:       5,
		ExpirationDelay: 5,
		TicketsPerBlock: []*ccmsg.Segment{
			&ccmsg.Segment{Length: 3, Value: 100},
		},
	}
	suite.escrow, err = suite.provider.NewEscrow(ei)
	if err != nil {
		t.Fatalf("failed to construct escrow: %v", err)
	}

	suite.clientPublic, suite.clientPrivate, err = ed25519.GenerateKey(nil)
}

/*
func (suite *TicketBundleTestSuite) randomDigests(n uint) [][]byte {
	t := suite.T()

	dd := make([][]byte, n)
	for i := uint(0); i < n; i++ {
		dd[i] = make([]byte, sha512.Size384)
		if _, err := rand.Read(dd[i]); err != nil {
			t.Fatalf("failed to generate random digest: %v", err)
		}
	}

	return dd
}
*/

func (suite *TicketBundleTestSuite) generateCacheInfo() *ParticipatingCache {
	pub, _, err := ed25519.GenerateKey(nil) // TOOS: use faster, lower-quality entropy?
	if err != nil {
		panic(errors.Wrap(err, "failed to generate keypair"))
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

	bp := &BundleParams{
		ClientPublicKey:   suite.clientPublic,
		RequestSequenceNo: 123,
		Escrow:            suite.escrow,
		Object:            obj,
		ObjectID:          456, // XXX: Should this move into the ContentObject interface?
		Entries: []BundleEntryParams{
			BundleEntryParams{TicketNo: 0, BlockIdx: 0, Cache: caches[0]},
			BundleEntryParams{TicketNo: 1, BlockIdx: 1, Cache: caches[1]},
			// BundleEntryParams{TicketNo: 2, BlockIdx: 2, Cache: caches[2]},
			// BundleEntryParams{TicketNo: 3, BlockIdx: 3, Cache: caches[3]},
		},
	}

	batchSigner, err := batchsignature.NewTrivialBatchSigner(suite.escrow.PrivateKey.(ed25519.PrivateKey))
	if err != nil {
		t.Fatalf("failed to construct batch signer: %v", err)
	}

	gen := NewBundleGenerator(batchSigner)
	bundle, err := gen.GenerateTicketBundle(bp)
	assert.Nil(t, err, "failed to generate ticket bundle")

	// TODO: more!
	_ = bundle
}
