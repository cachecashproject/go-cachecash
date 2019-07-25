package ledger

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/ed25519"
)

type AddressTestSuite struct {
	suite.Suite

	l *logrus.Logger
}

func TestAddressTestSuite(t *testing.T) {
	suite.Run(t, new(AddressTestSuite))
}

func (suite *AddressTestSuite) SetupTest() {
	t := suite.T()

	l := logrus.New()
	suite.l = l

	_ = t
}

func (suite *AddressTestSuite) makeP2WPKHAddress() Address {
	t := suite.T()

	pub, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate keypair: %v", err)
	}

	return MakeP2WPKHAddress(pub)
}

func (suite *AddressTestSuite) TestBase58_RoundTrip() {
	t := suite.T()

	a := suite.makeP2WPKHAddress()
	s := a.Base58Check()

	a2, err := ParseAddress(s)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, a, a2)
}

func (suite *AddressTestSuite) TestBase58_Malformed() {
	t := suite.T()

	a, err := ParseAddress("123abc*123abc*123abc*123abc")
	assert.NotNil(t, err)
	assert.Nil(t, a)
}

func (suite *AddressTestSuite) TestBase58_Truncated() {
	t := suite.T()

	a := suite.makeP2WPKHAddress()
	s := a.Base58Check()
	a2, err := ParseAddress(s[:len(s)-1])
	assert.NotNil(t, err)
	assert.Nil(t, a2)
}
