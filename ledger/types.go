package ledger

import (
	"bytes"
	"encoding/hex"

	"github.com/pkg/errors"
)

const (
	BlockIDSize = 32
)

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
