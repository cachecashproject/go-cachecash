package blockexplorer

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cachecashproject/go-cachecash/ledger"
	"github.com/cachecashproject/go-cachecash/testutil"

	"github.com/cachecashproject/go-cachecash/ccmsg"

	"github.com/stretchr/testify/mock"

	"github.com/gin-gonic/gin"
	proto "github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/cachecashproject/go-cachecash/ledgerservice"
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
	assert.Contains(t, body, "<a href=\"http://localhost/blocks\">Blocks</a>")
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
	assert.Equal(t, "{\"_links\":{\"links\":{\"blocks\":{\"href\":\"http://localhost/blocks\",\"name\":\"Blocks\"},\"self\":{\"href\":\"http://localhost/\",\"name\":\"Self\"}}}}", rr.Body.String())
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
	assert.Equal(t, "http://localhost/blocks", newroot.XLinks.Links["blocks"].Href)
}

func getBlocksClient() *ledgerservice.LedgerClient {
	client_mock := ledgerservice.NewLedgerClientMock()
	client := ledgerservice.LedgerClient{GrpcClient: client_mock}
	block := ledger.Block{
		Header: &ledger.BlockHeader{
			Version:       123,
			PreviousBlock: [32]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
			MerkleRoot:    []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
			Timestamp:     1234,
			Signature:     []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63},
		},
		Transactions: &ledger.Transactions{Transactions: []*ledger.Transaction{
			{
				Version: 0x01,
				Flags:   0x0000,
				Body: &ledger.TransferTransaction{
					Inputs: []ledger.TransactionInput{
						{
							Outpoint: ledger.Outpoint{
								PreviousTx: ledger.MustDecodeTXID("deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
								Index:      0,
							},
							ScriptSig:  testutil.MustDecodeString("abc123"),
							SequenceNo: 0xFFFFFFFF,
						},
						{
							Outpoint: ledger.Outpoint{
								PreviousTx: ledger.MustDecodeTXID("deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
								Index:      0,
							},
							ScriptSig:  testutil.MustDecodeString("abc123"),
							SequenceNo: 0xFFFFFFFF,
						},
					},
					Witnesses: []ledger.TransactionWitness{
						{
							// A zero-item stack.
						},
						{
							// A zero-item stack.
						},
					},
					Outputs: []ledger.TransactionOutput{
						{
							Value:        1234,
							ScriptPubKey: testutil.MustDecodeString("789abc"),
						},
						{
							Value:        5678,
							ScriptPubKey: testutil.MustDecodeString("def456"),
						},
						{
							Value:        9012,
							ScriptPubKey: testutil.MustDecodeString("012345"),
						},
					},
				},
			},
		},
		},
	}
	response := &ccmsg.GetBlocksResponse{Blocks: []*ledger.Block{&block}}
	client_mock.On("GetBlocks", mock.Anything, &ccmsg.GetBlocksRequest{
		StartDepth: -1,
	}, []grpc.CallOption(nil)).Return(response, nil)
	return &client
}

func (suite *ServerTestSuite) TestBlocksHTML() {
	t := suite.T()
	client := getBlocksClient()
	conf := ConfigFile{Root: "http://localhost/", HTTPAddr: "address"}
	server, err := newBlockExplorerServer(suite.l, client, &conf)
	assert.NotNil(t, server)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("GET", "/blocks", nil)
	req.Header.Set("Accept", gin.MIMEHTML)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	server.Handler.ServeHTTP(rr, req)
	assert.Equal(t, rr.Code, http.StatusOK)
	body := rr.Body.String()
	assert.Contains(t, body, "<link rel=\"self\" href=\"http://localhost/blocks\">")
	assert.Contains(t, body, "<a href=\"http://localhost/blocks/6c3e5bad22de33baefe7799cf84229a064f0caf64136931aa8066aead177dfbd\">6c3e5bad22de33baefe7799cf84229a064f0caf64136931aa8066aead177dfbd</a>")
	assert.Contains(t, body, "<a href=\"http://localhost/blocks/6c3e5bad22de33baefe7799cf84229a064f0caf64136931aa8066aead177dfbd/tx/517f15c2fa884c12086f08971b70aa1f34c6fb1034c8a54fa8a321c9d38e5341\">517f15c2fa884c12086f08971b70aa1f34c6fb1034c8a54fa8a321c9d38e5341</a>")
}

func (suite *ServerTestSuite) TestBlocksJSON() {
	t := suite.T()
	client := getBlocksClient()
	conf := ConfigFile{Root: "http://localhost/blocks", HTTPAddr: "address"}
	server, err := newBlockExplorerServer(suite.l, client, &conf)
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
	assert.Equal(t, "{\"_links\":{\"links\":{\"blocks\":{\"href\":\"http://localhost/blocks\",\"name\":\"Blocks\"},\"self\":{\"href\":\"http://localhost/\",\"name\":\"Self\"}}}}", rr.Body.String())
	assert.Equal(t, "{\"_links\":{\"links\":{\"blocks\":{\"href\":\"http://localhost/blocks\",\"name\":\"Blocks\"},\"self\":{\"href\":\"http://localhost/\",\"name\":\"Self\"}}}}", rr.Body.String())
	assert.Equal(t, "{\"_links\":{\"links\":{\"blocks\":{\"href\":\"http://localhost/blocks\",\"name\":\"Blocks\"},\"self\":{\"href\":\"http://localhost/\",\"name\":\"Self\"}}}}", rr.Body.String())
}

func (suite *ServerTestSuite) TestBlocksPB() {
	t := suite.T()
	client := getBlocksClient()
	conf := ConfigFile{Root: "http://localhost/blocks", HTTPAddr: "address"}
	server, err := newBlockExplorerServer(suite.l, client, &conf)
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
	assert.Equal(t, "http://localhost/blocks", newroot.XLinks.Links["blocks"].Href)
}
