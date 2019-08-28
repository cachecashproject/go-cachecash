package txscript

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/ed25519"
)

type SigHashable interface {
	SigHash(script *Script, txIdx int, inputAmount int64) ([]byte, error)
}

func MakeOutputScript(pubkey ed25519.PublicKey) ([]byte, error) {
	pubKeyHash := Hash160Sum(pubkey)
	scriptPubKey, err := MakeP2WPKHOutputScript(pubKeyHash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create scriptPubKey")
	}

	scriptBytes, err := scriptPubKey.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal output script")
	}

	return scriptBytes, nil
}

func MakeInputScript(pubkey ed25519.PublicKey) ([]byte, error) {
	pubKeyHash := Hash160Sum(pubkey)
	scriptPubKey, err := MakeP2WPKHInputScript(pubKeyHash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create scriptPubKey")
	}

	scriptBytes, err := scriptPubKey.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal input script")
	}

	return scriptBytes, nil
}
