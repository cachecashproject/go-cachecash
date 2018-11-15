package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type BlockCipherTestSuite struct {
	suite.Suite

	key   []byte
	iv    []byte
	block cipher.Block
}

const (
	BlockQty = 5
)

func TestBlockCipherTestSuite(t *testing.T) {
	suite.Run(t, new(BlockCipherTestSuite))
}

func (suite *BlockCipherTestSuite) SetupTest() {
	t := suite.T()

	suite.key = make([]byte, 16)
	if _, err := rand.Read(suite.key); err != nil {
		t.Fatal(errors.Wrap(err, "failed to generate random key"))
	}

	suite.iv = make([]byte, aes.BlockSize)
	if _, err := rand.Read(suite.iv); err != nil {
		t.Fatal(errors.Wrap(err, "failed to generate random IV"))
	}

	var err error
	suite.block, err = aes.NewCipher(suite.key)
	if err != nil {
		t.Fatal(errors.Wrap(err, "failed to create block cipher"))
	}
}

func (suite *BlockCipherTestSuite) TestSingleBlockEncryption() {
	t := suite.T()

	plaintext := make([]byte, aes.BlockSize*BlockQty)

	if _, err := rand.Read(plaintext); err != nil {
		t.Fatal(errors.Wrap(err, "failed to generate random plaintext"))
	}

	ciphertext := make([]byte, len(plaintext))
	stream := cipher.NewCTR(suite.block, suite.iv)
	stream.XORKeyStream(ciphertext, plaintext)

	for i := uint32(0); i < BlockQty; i++ {
		expected := ciphertext[i*aes.BlockSize : (i+1)*aes.BlockSize]
		plaintextBlock := plaintext[i*aes.BlockSize : (i+1)*aes.BlockSize]

		result, err := EncryptBlock(plaintextBlock, suite.key, suite.iv, i)
		if !assert.Nil(t, err) {
			return
		}
		if !assert.Equal(t, expected, result, fmt.Sprintf("unexpected result at block %v", i)) {
			return
		}
	}
}
