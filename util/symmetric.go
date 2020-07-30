package util

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/pkg/errors"
)

// Also decrypts chunk, since we're using AES in the CTR mode.
func EncryptChunk(chunkIdx uint64, reqSeqNo uint64, sessionKey, in []byte) ([]byte, error) {
	// Derive IV.
	iv, err := KeyedPRF(Uint64ToLE(chunkIdx), uint32(reqSeqNo), sessionKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate IV")
	}

	// Set up our cipher.
	block, err := aes.NewCipher(sessionKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to construct block cipher")
	}
	stream := cipher.NewCTR(block, iv)

	out := make([]byte, len(in))
	stream.XORKeyStream(out, in)
	return out, nil
}
