package util

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/pkg/errors"
)

// EncryptBlock encrypts a single plaintext block using AES in the CTR mode.  Unusually, it lets the caller specify the
// counter's value; the counter is added to the IV.
func EncryptBlock(plaintext []byte, key []byte, iv []byte, counter uint32) ([]byte, error) {
	if len(plaintext) != aes.BlockSize {
		return nil, errors.New("cleartext must be exactly one block in length")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to construct block cipher")
	}

	// Duplicate iv so that we don't mutate the array backing the slice we were passed.
	ivCtr := make([]byte, len(iv))
	copy(ivCtr, iv)
	incrementIV(ivCtr, counter)

	ciphertext := make([]byte, aes.BlockSize)
	stream := cipher.NewCTR(block, ivCtr)
	stream.XORKeyStream(ciphertext, plaintext)

	return ciphertext, nil
}

// XXX: We can make this much more efficient; this code is just borrowed from `crypto/cipher/ctr.go`.
func incrementIV(iv []byte, counter uint32) {
	for j := uint32(0); j < counter; j++ {
		// Increment counter
		for i := len(iv) - 1; i >= 0; i-- {
			iv[i]++
			if iv[i] != 0 {
				break
			}
		}
	}
}
