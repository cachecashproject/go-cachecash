package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha512"
	"encoding/binary"

	"github.com/pkg/errors"
)

// TODO: Add tests for behavior on short inputs!  Should return an error, not panic.

// XXX: Document me!
func KeyedPRF(prfInput []byte, requestSeqNo uint32, key []byte) ([]byte, error) {
	h := sha512.New384()
	// XXX: Why is it important that we feed requestSeqNo in here?
	if err := binary.Write(h, binary.LittleEndian, requestSeqNo); err != nil {
		return nil, errors.Wrap(err, "failed to hash request sequence nubmer")
	}
	if _, err := h.Write(prfInput); err != nil {
		return nil, errors.Wrap(err, "failed to hash PRF input")
	}
	digest := h.Sum(nil)

	// We use the first portion of the digest as the IV, and the following part as the plaintext to be encrypted.
	iv := digest[:aes.BlockSize]
	plaintext := digest[aes.BlockSize : 3*aes.BlockSize]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to construct block cipher")
	}

	mode := cipher.NewCBCEncrypter(block, iv)

	ciphertext := make([]byte, len(plaintext))
	mode.CryptBlocks(ciphertext, plaintext)

	// Our result is the last block of the ciphertext.
	// XXX: Why do we encrypt two blocks if we are only going to use a single block of the ciphertext?
	return ciphertext[aes.BlockSize:], nil
}
