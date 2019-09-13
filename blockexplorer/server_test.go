package blockexplorer

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ServerTestSuite struct {
	suite.Suite

	l *logrus.Logger
}

func (suite *ServerTestSuite) SetupTest() {
	l := logrus.New()
	l.SetLevel(logrus.TraceLevel)
	suite.l = l
	gin.SetMode(gin.TestMode)
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}

func (suite *ServerTestSuite) TestBadRoot() {
	t := suite.T()

	conf := ConfigFile{Root: "/"}
	server, err := newBlockExplorerServer(suite.l, nil, &conf)
	assert.Nil(t, server)
	assert.Regexp(t, "absolute", err)
	conf.Root = "%2"
	server, err = newBlockExplorerServer(suite.l, nil, &conf)
	assert.Nil(t, server)
	assert.Regexp(t, "not parse", err)
}

func (suite *ServerTestSuite) TestNewBlockExplorerServer() {
	t := suite.T()

	conf := ConfigFile{Root: "http://localhost/", HTTPAddr: "address"}
	server, err := newBlockExplorerServer(suite.l, nil, &conf)
	assert.NotNil(t, server)
	assert.Nil(t, err)
	assert.Equal(t, server.Addr, conf.HTTPAddr)
}

func (suite *ServerTestSuite) TestRootDocHTML() {
	t := suite.T()

	conf := ConfigFile{Root: "http://localhost/", HTTPAddr: "address"}
	server, err := newBlockExplorerServer(suite.l, nil, &conf)
	assert.NotNil(t, server)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", gin.MIMEHTML)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	server.Handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}
