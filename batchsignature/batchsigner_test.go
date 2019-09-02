package batchsignature

import (
	"testing"

	"golang.org/x/crypto/ed25519"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type BatchSignerTestSuite struct {
	suite.Suite
}

func TestBatchSignerTestSuite(t *testing.T) {
	suite.Run(t, new(BatchSignerTestSuite))
}

func (suite *BatchSignerTestSuite) TestSmoke() {
	t := suite.T()

	_, priv, err := ed25519.GenerateKey(nil)
	assert.Nil(t, err)
	signer, err := NewTrivialBatchSigner(priv)
	assert.Nil(t, err)
	sig, err := signer.BatchSign([]byte("content"))
	assert.Nil(t, err)
	ok, err := Verify([]byte("content"), sig)
	assert.Nil(t, err)
	assert.True(t, ok)
}
