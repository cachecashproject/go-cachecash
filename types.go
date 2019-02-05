package cachecash

import (
	"crypto"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"golang.org/x/crypto/ed25519"
)

func PublicKeyMessage(k crypto.PublicKey) *ccmsg.PublicKey {
	// XXX: Should we encode these keys as described in e.g. https://stackoverflow.com/questions/21322182/how-to-store-ecdsa-private-key-in-go?
	public := []byte(k.(ed25519.PublicKey))
	return &ccmsg.PublicKey{PublicKey: public}
}
