package ledgerservice

import (
	"encoding/json"
	"testing"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/ledger"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

type LedgerTestSuite struct {
	suite.Suite

	l *logrus.Logger
}

func TestLedgerTestSuite(t *testing.T) {
	suite.Run(t, new(LedgerTestSuite))
}

func (suite *LedgerTestSuite) SetupTest() {
	l := logrus.New()
	l.SetLevel(logrus.TraceLevel)
	suite.l = l
}

func (suite *LedgerTestSuite) TestDefaultQueries() {
	t := suite.T()
	l := suite.l
	// Replication friendly requests - start with height 0 and work up,
	// pagination for dealing with high fan out if it ever happens.
	// Default request: first 50 blocks
	req := &ccmsg.GetBlocksRequest{}
	query, err := requestToQuery(l, req)
	assert.Nil(t, err)
	expected := &blockquery{qm: []qm.QueryMod{qm.Limit(50), qm.OrderBy("height ASC, block_id ASC")}, flipped: false}
	assert.Equal(t, expected, query)
	// Get higher blocks
	req.StartDepth = 50
	query, err = requestToQuery(l, req)
	assert.Nil(t, err)
	expected = &blockquery{
		qm:      []qm.QueryMod{qm.Limit(50), qm.Where("height >= ?", int64(50)), qm.OrderBy("height ASC, block_id ASC")},
		flipped: false}
	assert.Equal(t, expected, query)
	// pagination of regular requests
	var blockid_bytes [ledger.BlockIDSize]byte
	blockid_bytes[0] = 31
	blockid_bytes[1] = 32
	blockid_bytes[2] = 33
	blockid := ledger.BlockID(blockid_bytes)
	token := pageToken{Forward: true, BlockID: blockid, Height: 52}
	req.PageToken, _ = json.Marshal(token)
	query, err = requestToQuery(l, req)
	assert.Nil(t, err)
	expected = &blockquery{
		qm: []qm.QueryMod{qm.Limit(50),
			qm.Where("height > ?", int64(52)), qm.Or("(height = ? AND block_id > ?)", int64(52), blockid_bytes[:]),
			qm.OrderBy("height ASC, block_id ASC"),
		},
		flipped: false,
		token:   &token}
	assert.Equal(t, expected, query)
	// and backwards
	token.Forward = false
	req.PageToken, _ = json.Marshal(token)
	query, err = requestToQuery(l, req)
	assert.Nil(t, err)
	expected = &blockquery{
		qm: []qm.QueryMod{qm.Limit(50),
			qm.Where("height < ?", int64(52)), qm.Or("(height = ? AND block_id < ?)", int64(52), blockid_bytes[:]),
			qm.OrderBy("height DESC, block_id DESC"),
		},
		flipped: true,
		token:   &token}
	assert.Equal(t, expected, query)

	// Browser friendly requests: latest 50 blocks
	req.StartDepth = -1
	req.PageToken = nil
	query, err = requestToQuery(l, req)
	assert.Nil(t, err)
	expected = &blockquery{
		qm:      []qm.QueryMod{qm.Limit(50), qm.OrderBy("height DESC, block_id DESC")},
		flipped: false}
	assert.Equal(t, expected, query)
	// Browser friendly request: page forward 50 more
	req.StartDepth = -1
	token.Forward = true
	req.PageToken, _ = json.Marshal(token)
	query, err = requestToQuery(l, req)
	assert.Nil(t, err)
	expected = &blockquery{
		qm: []qm.QueryMod{qm.Limit(50),
			qm.Where("height < ?", int64(52)), qm.Or("(height = ? AND block_id < ?)", int64(52), blockid_bytes[:]),
			qm.OrderBy("height DESC, block_id DESC"),
		},
		flipped: false,
		token:   &token}
	assert.Equal(t, expected, query)
	// Browser friendly request: page backwards 50 more
	req.StartDepth = -1
	token.Forward = false
	req.PageToken, _ = json.Marshal(token)
	query, err = requestToQuery(l, req)
	assert.Nil(t, err)
	expected = &blockquery{
		qm: []qm.QueryMod{qm.Limit(50),
			qm.Where("height > ?", int64(52)), qm.Or("(height = ? AND block_id > ?)", int64(52), blockid_bytes[:]),
			qm.OrderBy("height ASC, block_id ASC"),
		},
		flipped: true,
		token:   &token}
	assert.Equal(t, expected, query)
	// Custom limits
	req.StartDepth = 0
	req.Limit = 1
	req.PageToken = nil
	query, err = requestToQuery(l, req)
	assert.Nil(t, err)
	expected = &blockquery{
		qm:      []qm.QueryMod{qm.Limit(1), qm.OrderBy("height ASC, block_id ASC")},
		flipped: false}
	assert.Equal(t, expected, query)
}

func (suite *LedgerTestSuite) TestBadQueries() {
	t := suite.T()
	l := suite.l
	// More than -1 is invalid
	req := &ccmsg.GetBlocksRequest{StartDepth: -2}
	query, err := requestToQuery(l, req)
	assert.Regexp(t, "invalid", err)
	assert.Nil(t, query)
	// More than 100 items is invalid
	req = &ccmsg.GetBlocksRequest{Limit: 101}
	query, err = requestToQuery(l, req)
	assert.Regexp(t, "limit is too high", err)
	assert.Nil(t, query)
	// Bad page tokens cause errors
	req = &ccmsg.GetBlocksRequest{PageToken: []byte("123")}
	query, err = requestToQuery(l, req)
	assert.Regexp(t, "bad page token", err)
	assert.Nil(t, query)
}
