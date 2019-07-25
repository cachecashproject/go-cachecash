package wallet

import (
	"testing"

	"github.com/cachecashproject/go-cachecash/ledger"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type WalletTestSuite struct {
	suite.Suite

	l *logrus.Logger
}

func TestWalletTestSuite(t *testing.T) {
	suite.Run(t, new(WalletTestSuite))
}

func (suite *WalletTestSuite) SetupTest() {
	t := suite.T()

	l := logrus.New()
	suite.l = l

	_ = t
}

func (suite *WalletTestSuite) TestFoo() {
	t := suite.T()

	ac, err := GenerateAccount()
	if !assert.Nil(t, err) {
		return
	}

	a := ac.P2WPKHAddress(ledger.AddressP2WPKHTestnet)

	_ = a
}
