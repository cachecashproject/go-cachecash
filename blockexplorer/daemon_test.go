package blockexplorer

import (
	"context"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DaemonTestSuite struct {
	suite.Suite

	l *logrus.Logger
}

func (suite *DaemonTestSuite) SetupTest() {
	l := logrus.New()
	l.SetLevel(logrus.TraceLevel)
	suite.l = l
	gin.SetMode(gin.TestMode)
}

func TestDaemonTestSuite(t *testing.T) {
	suite.Run(t, new(DaemonTestSuite))
}

func (suite *DaemonTestSuite) TestNewDaemonErrors() {
	t := suite.T()
	conf := ConfigFile{
		HTTPAddr:   "127.0.0.1:0",
		Insecure:   false,
		LedgerAddr: "127.0.0.1:0",
		Root:       "%2",
		StatusAddr: "fred",
	}
	// Ledger failure path
	daemon, err := NewDaemon(suite.l, &conf)
	assert.Nil(t, daemon)
	assert.Regexp(t, "create ledger client", err)
	conf.Insecure = true
	// api server failure path
	daemon, err = NewDaemon(suite.l, &conf)
	assert.Nil(t, daemon)
	assert.Regexp(t, "create explorer server", err)
}

func (suite *DaemonTestSuite) TestNewDaemon() {
	t := suite.T()
	conf := ConfigFile{
		HTTPAddr:   "127.0.0.1:0",
		Insecure:   true,
		LedgerAddr: "127.0.0.1:0",
		Root:       "http://localhost/",
		StatusAddr: "127.0.0.1:0",
	}
	daemon, err := NewDaemon(suite.l, &conf)
	assert.NotNil(t, daemon)
	assert.Nil(t, err)
	err = daemon.Start()
	assert.Nil(t, err)
	gracefulCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = daemon.Shutdown(gracefulCtx)
	assert.Nil(t, err)
}
