package pkg

import (
	"github.com/cachecashproject/go-cachecash/ranger"
)

// XXX this is a shim to help with the TxType interface that lives alongside
// Transaction in our good.yml definition. The order of these must be preserved
// otherwise certain fuzz tests may not function the same.
const (
	TxTypeTransfer     uint8 = 0
	TxTypeGenesis            = 1
	TxTypeGlobalConfig       = 2
	TxTypeEscrowOpen         = 3
)

type TransactionBody interface {
	TxType() uint8
	ranger.Marshaler
}

func (obj *TransferTransaction) TxType() uint8     { return TxTypeTransfer }
func (obj *GenesisTransaction) TxType() uint8      { return TxTypeGenesis }
func (obj *EscrowOpenTransaction) TxType() uint8   { return TxTypeEscrowOpen }
func (obj *GlobalConfigTransaction) TxType() uint8 { return TxTypeGlobalConfig }
