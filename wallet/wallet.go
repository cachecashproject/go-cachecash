package wallet

import (
	"crypto/sha256"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ed25519"

	"github.com/cachecashproject/go-cachecash/ledger"
)

type Account struct {
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey // May be nil
}

func GenerateAccount() (*Account, error) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate keypair")
	}

	return &Account{
		PublicKey:  pub,
		PrivateKey: priv,
	}, nil
}

func (ac *Account) P2WPKHAddress(v ledger.AddressVersion) *ledger.P2WPKHAddress {
	pkh := sha256.Sum256(ac.PublicKey)

	return &ledger.P2WPKHAddress{
		AddressVersion:        v,
		WitnessProgramVersion: 0,
		PublicKeyHash:         pkh[:ledger.AddressHashSize],
	}
}
