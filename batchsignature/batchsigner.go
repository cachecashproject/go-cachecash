package batchsignature

import (
	"crypto"

	cachecash "github.com/cachecashproject/go-cachecash"
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ed25519"
)

// BatchSigner is a scalable message signing interface. Batch signatures include
// the path and adjacent values from a merkle tree plus the root of the tree -
// the batch residue. This permits that single message to be verified without
// the rest of the batch, and for signing to have performed a single asymmetric
// signing operation that signed all of the messages in the batch. Each signed
// message receives its own distinct batch signature, as it has a unique batch
// residue.
type BatchSigner interface {
	// BatchSign signs single message `d`. Implementations may block for
	// (slightly) longer than a regular signing in order to batch up multiple
	// signature requests.
	BatchSign(d []byte) (*ccmsg.BatchSignature, error)
}

type trivialBatchSigner struct {
	signer ed25519.PrivateKey
}

var _ BatchSigner = (*trivialBatchSigner)(nil)

// NewTrivialBatchSigner constructs a BatchSigner which signs messages using the
// given private key. TrivialBatchSigner always uses a batch size of 1 (and so
// never blocks to gather more items for the batch).
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
