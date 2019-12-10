package models

import (
	"bytes"
	"encoding/hex"

	"github.com/pkg/errors"
)

const (
	TransactionIDSize = 32
)

type TXID [TransactionIDSize]byte

func (txid TXID) Equal(o TXID) bool {
	return bytes.Equal(txid[:], o[:])
}

// String converts the TXID to string.
func (txid TXID) String() string {
	return hex.EncodeToString(txid[:])
}

func BytesToTXID(x []byte) (TXID, error) {
	if len(x) != TransactionIDSize {
		return TXID{}, errors.New("bad size for TXID")
	}

	var id TXID
	copy(id[:], x[0:TransactionIDSize])
	return id, nil
}
