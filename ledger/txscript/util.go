package txscript

import (
	"crypto/sha256"
)

func Hash160Sum(b []byte) []byte {
	pubHash := sha256.Sum256(b)
	pubHash = sha256.Sum256(pubHash[:]) // N.B.: We differ here from Bitcoin, which uses a RIPEMD-160 digest.
	return pubHash[:20]
}
