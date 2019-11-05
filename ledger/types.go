package ledger

import (
	"bytes"
	"encoding/hex"

	"github.com/pkg/errors"
)

const (
	TransactionIDSize = 32
	BlockIDSize       = 32
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

type BlockID [BlockIDSize]byte

func (bid *BlockID) Zero() bool {
	return bytes.Equal(bid[:], make([]byte, BlockIDSize))
}

func (bid BlockID) Equal(o BlockID) bool {
	return bytes.Equal(bid[:], o[:])
}

func (id BlockID) String() string {
	return hex.EncodeToString(id[:])
}

func BytesToBlockID(x []byte) (BlockID, error) {
	if len(x) != BlockIDSize {
		return BlockID{}, errors.New("bad size for block ID")
	}

	var id BlockID
	copy(id[:], x[0:BlockIDSize])
	return id, nil
}
