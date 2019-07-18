package ledger

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	"github.com/btcsuite/btcutil/base58"
	"github.com/pkg/errors"
)

const (
	AddressHashSize = 20
)

type AddressVersion uint8

// These values are Cachecash-specific and do not line up with Bitcoin address version assignments.
const (
	AddressP2WPKHMainnet AddressVersion = 0x01
	AddressP2WPKHGoodnet                = 0x02
	AddressP2WPKHTestnet                = 0x03
)

type Address interface {
	Bytes() []byte
	Base58Check() string
}

type P2WPKHAddress struct {
	AddressVersion        AddressVersion
	WitnessProgramVersion uint8 // Must be 0, as in Bitcoin.
	PublicKeyHash         []byte
}

var _ Address = (*P2WPKHAddress)(nil)

func (a *P2WPKHAddress) Bytes() []byte {
	data := []byte{
		byte(a.AddressVersion),
		a.WitnessProgramVersion,
		0x00, // Why do Bitcoin P2WPKH addresses have this extra byte here?
	}
	data = append(data, a.PublicKeyHash...)
	return data
}

func (a *P2WPKHAddress) Base58Check() string {
	data := a.Bytes()

	ck := sha256.Sum256(data)
	ck = sha256.Sum256(ck[:])
	data = append(data, ck[:4]...)

	return base58.Encode(data)
}

func Base58CheckDecode(s string) ([]byte, error) {
	// N.B.: If s is not a base58-encoded string, dc will be 0-length.
	dc := base58.Decode(s)
	if len(dc) <= 4 {
		return nil, errors.New("check-encoded address malformed or too short")
	}

	data := dc[:len(dc)-4]
	actualCk := dc[len(dc)-4:]

	ck := sha256.Sum256(data)
	ck = sha256.Sum256(ck[:])

	if !bytes.Equal(actualCk, ck[:4]) {
		return nil, fmt.Errorf("checksum failed (got %#x, expected %#x)", ck[:4], actualCk)
	}

	return data, nil
}

func ParseAddress(s string) (Address, error) {
	d, err := Base58CheckDecode(s)
	if err != nil {
		return nil, err
	}
	return ParseAddressBytes(d)
}

func ParseAddressBytes(data []byte) (Address, error) {
	if len(data) < 1 {
		return nil, errors.New("address too short")
	}

	v := (AddressVersion)(data[0])
	switch v {
	case AddressP2WPKHMainnet:
	case AddressP2WPKHGoodnet:
	case AddressP2WPKHTestnet:
		a := &P2WPKHAddress{
			AddressVersion:        v,
			WitnessProgramVersion: data[1],
			PublicKeyHash:         data[3:],
		}
		if a.WitnessProgramVersion != 0 {
			return nil, errors.New("unexpected witness program version")
		}
		if len(a.PublicKeyHash) != AddressHashSize {
			return nil, errors.New("unexpected public key hash size")
		}
		return a, nil
	}

	return nil, errors.New("unexpected address version")
}
