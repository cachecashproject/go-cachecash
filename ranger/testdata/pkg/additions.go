package pkg

import (
	"github.com/cachecashproject/go-cachecash/ranger"
)

const (
	TxTypeTransfer     uint8 = iota
	TxTypeGenesis      uint8 = iota
	TxTypeGlobalConfig uint8 = iota
	TxTypeEscrowOpen   uint8 = iota
)

type TransactionBody interface {
	TxType() uint8
	ranger.Marshaler
}

func (obj *TransferTransaction) TxType() uint8     { return TxTypeTransfer }
func (obj *GenesisTransaction) TxType() uint8      { return TxTypeGenesis }
func (obj *EscrowOpenTransaction) TxType() uint8   { return TxTypeEscrowOpen }
func (obj *GlobalConfigTransaction) TxType() uint8 { return TxTypeGlobalConfig }
