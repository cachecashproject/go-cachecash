package util

import (
	"testing"

	"github.com/cachecashproject/go-cachecash/testutil"
	"github.com/stretchr/testify/assert"
)

func TestEncryptChunk(t *testing.T) {
	chunkIdx := uint64(43)
	reqSeqNo := uint64(9)
	original := testutil.RandBytes(1024)
	key := testutil.RandBytes(16)

	ciphertext, err := EncryptChunk(chunkIdx, reqSeqNo, key, original)
	assert.Nil(t, err)

	plaintext, err := EncryptChunk(chunkIdx, reqSeqNo, key, ciphertext)
	assert.Nil(t, err)

	assert.Equal(t, original, plaintext)
}
