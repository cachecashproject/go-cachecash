package txscript

import (
	"testing"

	"github.com/cachecashproject/go-cachecash/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ScriptTestSuite struct {
	suite.Suite

	l *logrus.Logger
}

func TestScriptTestSuite(t *testing.T) {
	suite.Run(t, new(ScriptTestSuite))
}

func (suite *ScriptTestSuite) SetupTest() {
	t := suite.T()

	l := logrus.New()
	suite.l = l

	_ = t
}

func (suite *ScriptTestSuite) TestP2WPKHOutput_StandardOutput() {
	t := suite.T()

	pubKeyHash := testutil.MustDecodeString("deadb33fdeadb33fdeadb33fdeadb33fdeadb33f")
	scr, err := MakeP2WPKHOutputScript(pubKeyHash)
	assert.Nil(t, err)

	assert.Nil(t, scr.StandardOutput(), "transaction should be standard")
}

func (suite *ScriptTestSuite) TestP2WPKHOutput_Marshal() {
	t := suite.T()

	pubKeyHash := testutil.MustDecodeString("deadb33fdeadb33fdeadb33fdeadb33fdeadb33f")
	scr, err := MakeP2WPKHOutputScript(pubKeyHash)
	assert.Nil(t, err)

	buf, err := scr.Marshal()
	assert.Nil(t, err)
	assert.Equal(t, testutil.MustDecodeString("0014deadb33fdeadb33fdeadb33fdeadb33fdeadb33f"), buf)
}

func (suite *ScriptTestSuite) TestP2WPKHOutput_ParseScript() {
	t := suite.T()

	buf := testutil.MustDecodeString("0014deadb33fdeadb33fdeadb33fdeadb33fdeadb33f")
	scr, err := ParseScript(buf)
	assert.Nil(t, err)

	assert.Nil(t, scr.StandardOutput())
}

func (suite *ScriptTestSuite) TestP2WPKHOutput_ParseScript_BadOpcode() {
	t := suite.T()

	buf := testutil.MustDecodeString("00ff")
	scr, err := ParseScript(buf)
	assert.NotNil(t, err)
	assert.Nil(t, scr)
}

func (suite *ScriptTestSuite) TestP2WPKHOutput_ParseScript_ImmediateUnderrun() {
	t := suite.T()

	buf := testutil.MustDecodeString("0014deadb33fdeadb33fdeadb33fdeadb33fdeadb3") // one byte missing
	scr, err := ParseScript(buf)
	assert.NotNil(t, err)
	assert.Nil(t, scr)
}

func (suite *ScriptTestSuite) TestP2WPKHOutput_PrettyPrint() {
	t := suite.T()

	pubKeyHash := testutil.MustDecodeString("deadb33fdeadb33fdeadb33fdeadb33fdeadb33f")
	scr, err := MakeP2WPKHOutputScript(pubKeyHash)
	assert.Nil(t, err)

	s, err := scr.PrettyPrint()
	assert.Nil(t, err)
	assert.Equal(t, "OP_0 OP_DATA_20 0xdeadb33fdeadb33fdeadb33fdeadb33fdeadb33f", s)
}

func (suite *ScriptTestSuite) TestP2WPKHInput_PrettyPrint() {
	t := suite.T()

	pubKeyHash := testutil.MustDecodeString("deadb33fdeadb33fdeadb33fdeadb33fdeadb33f")
	scr, err := MakeP2WPKHInputScript(pubKeyHash)
	assert.Nil(t, err)

	s, err := scr.PrettyPrint()
	assert.Nil(t, err)
	assert.Equal(t, "OP_DUP OP_HASH160 OP_DATA_20 0xdeadb33fdeadb33fdeadb33fdeadb33fdeadb33f OP_EQUALVERIFY OP_CHECKSIG", s)
}

// TODO:
// - Test that other scripts are not detected as being standard.
