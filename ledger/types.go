package ledger

import "bytes"

const (
	TransactionIDSize = 32
	BlockIDSize       = 32
)

type TXID [TransactionIDSize]byte

func (txid TXID) Equal(o TXID) bool {
	return bytes.Equal(txid[:], o[:])
}

type BlockID [BlockIDSize]byte

func (bid *BlockID) Zero() bool {
	return bytes.Equal(bid[:], make([]byte, BlockIDSize))
}

func (bid BlockID) Equal(o BlockID) bool {
	return bytes.Equal(bid[:], o[:])
}
