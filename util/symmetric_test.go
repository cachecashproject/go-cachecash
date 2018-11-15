package util

import (
	"testing"

	"github.com/kelleyk/go-cachecash/testutil"
	"github.com/stretchr/testify/assert"
)

func TestEncryptDataBlock(t *testing.T) {
	blockIdx := uint64(43)
	reqSeqNo := uint64(9)
	original := testutil.RandBytes(1024)
	key := testutil.RandBytes(16)

	ciphertext, err := EncryptDataBlock(blockIdx, reqSeqNo, key, original)
	assert.Nil(t, err)

	plaintext, err := EncryptDataBlock(blockIdx, reqSeqNo, key, ciphertext)
	assert.Nil(t, err)

	assert.Equal(t, original, plaintext)
}
