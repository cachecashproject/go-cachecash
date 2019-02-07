package cachecash

import (
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"golang.org/x/crypto/ed25519"
)

func PublicKeyMessage(k ed25519.PublicKey) *ccmsg.PublicKey {
	// XXX: Should we encode these keys as described in e.g. https://stackoverflow.com/questions/21322182/how-to-store-ecdsa-private-key-in-go?
	return &ccmsg.PublicKey{PublicKey: []byte(k)}
}
