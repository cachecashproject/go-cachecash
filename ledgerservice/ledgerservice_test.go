package ledgerservice

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type LedgerTestSuite struct {
	suite.Suite

	l *logrus.Logger

	// ledgerService *LedgerService
}

func TestLedgerTestSuite(t *testing.T) {
	suite.Run(t, new(LedgerTestSuite))
}

func (suite *LedgerTestSuite) SetupTest() {
	t := suite.T()

	l := logrus.New()
	suite.l = l

	// suite.ledgerService, err := NewLedgerService(l, db)

	_ = t
}

func (suite *LedgerTestSuite) TestNoop() {
	t := suite.T()

	// TODO: Test me!

	_ = t
}
