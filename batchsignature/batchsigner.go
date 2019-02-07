package batchsignature

import (
	"crypto"

	cachecash "github.com/cachecashproject/go-cachecash"
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ed25519"
)

type BatchSigner interface {
	BatchSign(d []byte) (*ccmsg.BatchSignature, error)
}

type trivialBatchSigner struct {
	signer ed25519.PrivateKey
}

var _ BatchSigner = (*trivialBatchSigner)(nil)

// NewTrivialBatchSigner returns a BatchSigner that individually signs each message as a single-element batch.
func NewTrivialBatchSigner(signer ed25519.PrivateKey) (BatchSigner, error) {
	return &trivialBatchSigner{signer: signer}, nil
}

func (bs *trivialBatchSigner) BatchSign(d []byte) (*ccmsg.BatchSignature, error) {
	rootDigest, trees, err := NewDigestTree([][]byte{d})
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate digest-trees")
	}

	// N.B.: See the godoc for `crypto/ed25519` for a discussion of the parameters to this call.  Passing nil as the
	// first argument makes Sign use crypto/rand.Reader for entropy.
	rootSig, err := bs.signer.Sign(nil, rootDigest, crypto.Hash(0))
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign root digest")
	}

	return &ccmsg.BatchSignature{
		PathDirection: trees[0].pathDirections,
		PathDigest:    trees[0].pathDigests,
		RootSignature: rootSig,
		SigningKey:    cachecash.PublicKeyMessage(bs.signer.Public().(ed25519.PublicKey)),
	}, nil
}
