package batchsignature

import (
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"golang.org/x/crypto/ed25519"
)

func Verify(d []byte, bs *ccmsg.BatchSignature) (bool, error) {
	tree := &BatchResidue{
		pathDirections: bs.PathDirection,
		pathDigests:    bs.PathDigest,
		leafDigest:     d,
	}
	ok := ed25519.Verify(bs.SigningKey.PublicKey, tree.RootDigest(), bs.RootSignature)
	return ok, nil
}
