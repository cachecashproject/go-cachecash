package metricsproxy

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"

	"github.com/stretchr/testify/assert"
)

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}

type ServerTestSuite struct {
	suite.Suite

	l *logrus.Logger
}

func (suite *ServerTestSuite) SetupTest() {
	l := logrus.New()
	l.SetLevel(logrus.DebugLevel)
	suite.l = l
}

func (suite *ServerTestSuite) TestExpireClient() {
	t := suite.T()

	server := newGRPCMetricsProxyServer(suite.l)
	server.generation = 11
	somecache := "somecache"
	server.metrics[somecache] = &scrapeStatus{generation: 0}
	mts, err := server.Gather()
	assert.Nil(t, err)
	assert.Empty(t, mts)
	assert.Empty(t, server.metrics)
}
