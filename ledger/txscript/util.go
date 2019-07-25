package txscript

import (
	"crypto/sha256"

	"golang.org/x/crypto/ripemd160"
)

func hash160Sum(b []byte) []byte {
	d := sha256.Sum256(b)
	return ripemd160Sum(d[:])
}

func ripemd160Sum(b []byte) []byte {
	h := ripemd160.New()
	h.Write(b)
	return h.Sum(nil)
}
