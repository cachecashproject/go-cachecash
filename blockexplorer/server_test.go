package blockexplorer

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	proto "github.com/golang/protobuf/proto"
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
	assert.Equal(t, rr.Code, http.StatusOK)
	body := rr.Body.String()
	assert.Contains(t, body, "<link rel=\"self\" href=\"http://localhost/\">")
	assert.Contains(t, body, "<a href=\"http://localhost/escrows\">Escrows</a>")
	assert.NotContains(t, body, "<a href=\"http://localhost/\">Self</a>")
}

func (suite *ServerTestSuite) TestRootDocJSON() {
	t := suite.T()

	conf := ConfigFile{Root: "http://localhost/", HTTPAddr: "address"}
	server, err := newBlockExplorerServer(suite.l, nil, &conf)
	assert.NotNil(t, server)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", gin.MIMEJSON)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	server.Handler.ServeHTTP(rr, req)
	assert.Equal(t, rr.Code, http.StatusOK)
	assert.Equal(t, "{\"_links\":{\"links\":{\"escrows\":{\"href\":\"http://localhost/escrows\",\"name\":\"Escrows\"},\"self\":{\"href\":\"http://localhost/\",\"name\":\"Self\"}}}}", rr.Body.String())
}

func (suite *ServerTestSuite) TestRootDocPB() {
	t := suite.T()

	conf := ConfigFile{Root: "http://localhost/", HTTPAddr: "address"}
	server, err := newBlockExplorerServer(suite.l, nil, &conf)
	assert.NotNil(t, server)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "application/protobuf")
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	server.Handler.ServeHTTP(rr, req)
	assert.Equal(t, rr.Code, http.StatusOK)
	newroot := &APIRoot{}
	body := rr.Body.Bytes()
	err = proto.Unmarshal(body, newroot)
	assert.Nil(t, err)
	assert.Len(t, newroot.XLinks.Links, 2)
	assert.Equal(t, "http://localhost/", newroot.XLinks.Links["self"].Href)
	assert.Equal(t, "http://localhost/escrows", newroot.XLinks.Links["escrows"].Href)
}
