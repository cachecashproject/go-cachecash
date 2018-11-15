package util

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/pkg/errors"
)

// XXX: This function and EncryptBlock, which encrypts a signle *cipher* block, need names that are less likely to cause
// confusion.
// Also decrypts blocks, since we're using AES in the CTR mode.
func EncryptDataBlock(blockIdx uint64, reqSeqNo uint64, sessionKey, in []byte) ([]byte, error) {
	// Derive IV.
	iv, err := KeyedPRF(Uint64ToLE(blockIdx), uint32(reqSeqNo), sessionKey)
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
