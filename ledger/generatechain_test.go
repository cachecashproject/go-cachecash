package ledger

import (
	"testing"

	"github.com/cachecashproject/go-cachecash/ledger/txscript"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/ed25519"
)

type GenerateChainTestSuite struct {
	suite.Suite

	l *logrus.Logger
}

func TestGenerateChainTestSuite(t *testing.T) {
	suite.Run(t, new(GenerateChainTestSuite))
}

func (suite *GenerateChainTestSuite) SetupTest() {
	t := suite.T()

	l := logrus.New()
	suite.l = l

	_ = t
}

func (suite *GenerateChainTestSuite) TestGenerateChain() {
	t := suite.T()

	pubCLA, privCLA, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate CLA keypair: %v", err)
	}
	_ = pubCLA

	pubA, privA, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate keypair: %v", err)
	}
	addrA := MakeP2WPKHAddress(pubA)
	outScrA, err := txscript.MakeP2WPKHOutputScript(addrA.PublicKeyHash)
	if err != nil {
		t.Fatalf("failed to generate output script: %v", err)
	}
	outScrBufA, err := outScrA.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal output script: %v", err)
	}

	txs := []Transaction{
		{
			Version: 0x01,
			Flags:   0x0000,
			Body: &GenesisTransaction{
				Outputs: []TransactionOutput{
					{
						Value:        100,
						ScriptPubKey: outScrBufA,
					},
				},
			},
		},
	}

	block0, err := NewBlock(privCLA, make([]byte, 32), txs)
	assert.Nil(t, err, "failed to create genesis block")
	_ = block0
	// ...

	pubB, privB, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate keypair: %v", err)
	}

	_, _ = privA, pubA
	_, _ = privB, pubB
}
