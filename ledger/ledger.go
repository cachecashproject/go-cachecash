package ledger

import (
	"encoding/binary"
	"encoding/json"

	"github.com/pkg/errors"
)

type TxType uint8

const (
	TxTypeUnknown    TxType = 0x00 // Not valid in serialized transactions.
	TxTypeTransfer          = 0x01
	TxTypeEscrowOpen        = 0x02
)

// In order to be used as a gogo/protobuf custom type, a struct must implement this interface...
type protobufCustomType interface {
	Marshal() ([]byte, error)
	MarshalJSON() ([]byte, error)
}

// ... and the pointer-to-struct type must implement this one.
type protobufCustomTypePtr interface {
	MarshalTo(data []byte) (n int, err error)
	Unmarshal(data []byte) error
	Size() int
	UnmarshalJSON(data []byte) error
}

type Transaction struct {
	Version uint8  // Must be 1.
	Flags   uint16 // Must be zero; no flags are defined for any transaction types.
	Body    TransactionBody
}

var _ protobufCustomType = Transaction{}
var _ protobufCustomTypePtr = (*Transaction)(nil)

func (tx Transaction) Marshal() ([]byte, error) {
	data := make([]byte, tx.Size())
	n, err := tx.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	if n != len(data) {
		return nil, errors.New("unexpected data length in Transaction.Marshal()")
	}
	return data, nil
}

func (tx *Transaction) MarshalTo(data []byte) (int, error) {
	// TODO: Length checking.

	data[0] = (byte)(0x01)             // version
	data[1] = (byte)(tx.Body.TxType()) // transaction type
	binary.LittleEndian.PutUint16(data[2:4], tx.Flags)

	n, err := tx.Body.MarshalTo(data[4:])
	if err != nil {
		return 0, err
	}
	return n + 4, nil
}

func (tx *Transaction) Unmarshal(data []byte) error {
	return nil
}

func (tx *Transaction) Size() int {
	// The type-independent header is four bytes:
	// - Version  uint8
	// - TxType   uint8
	// - Flags    uint16
	return 4 + tx.Body.Size()
}

func (tx Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(tx)
}

func (tx *Transaction) UnmarshalJSON(data []byte) error {
	// XXX: Not sure this will work with Body which has an interface type.
	err := json.Unmarshal(data, tx)
	if err != nil {
		return err
	}
	return nil
}

type TransactionBody interface {
	Size() int
	TxType() TxType
	MarshalTo(data []byte) (n int, err error)
}

// A TransferTransaction is very similar to a Bitcoin transaction.  It consumes one or more unspent outputs of previous
// transactions (UTXOs) and produces one or more new outputs.  If the sum of the outputs is larger than the sum of the
// inputs, the difference is collected by the miner.  (For the time being, this fee MUST be zero.)
//
// These transactions *always* have segregated witness (segwit) data.  The "flags" field is mandatory (whereas it is
// optional in Bitcoin).
//
// Cachecash has different restrictions for standard transactions than Bitcoin does.  Currently, only pay-to-pubkey-hash
// (P2PKH) outputs are permitted.  Pay-to-script-hash (P2SH) outputs will be added later in development.
//
// Also, we use the same "uvarint" encoding everywhere, instead of the "compactSize" encoding that Bitcoin uses at a
// protocol level.
//
type TransferTransaction struct {
	Inputs    []TransactionInput
	Outputs   []TransactionOutput
	Witnesses []TransactionWitness
	LockTime  uint32
}

var _ TransactionBody = (*TransferTransaction)(nil)

func (tx *TransferTransaction) Size() int {
	// We don't need to include len(tx.Witnesses) because it must be equal to len(tx.Inputs)
	n := 4 + UvarintSize(uint(len(tx.Inputs))) + UvarintSize(uint(len(tx.Outputs)))
	for _, o := range tx.Inputs {
		n += o.Size()
	}
	for _, o := range tx.Outputs {
		n += o.Size()
	}
	for _, o := range tx.Witnesses {
		n += o.Size()
	}
	return n
}

func (tx *TransferTransaction) TxType() TxType {
	return TxTypeTransfer
}

func (tx *TransferTransaction) MarshalTo(data []byte) (int, error) {
	var i int

	i = binary.PutUvarint(data, uint64(len(tx.Inputs)))
	data = data[i:]

	// XXX: There's quite a bit of repetition here.  Refactor?
	for _, o := range tx.Inputs {
		i, err := o.MarshalTo(data)
		if err != nil {
			return 0, errors.Wrap(err, "failed to marshal TransactionInput")
		}
		data = data[i:]
	}

	i = binary.PutUvarint(data, uint64(len(tx.Outputs)))
	data = data[i:]

	for _, o := range tx.Outputs {
		i, err := o.MarshalTo(data)
		if err != nil {
			return 0, errors.Wrap(err, "failed to marshal TransactionOutput")
		}
		data = data[i:]
	}

	for _, o := range tx.Witnesses {
		i, err := o.MarshalTo(data)
		if err != nil {
			return 0, errors.Wrap(err, "failed to marshal TransactionWitness")
		}
		data = data[i:]
	}

	return 0, nil
}

// An EscrowOpenTransaction ...
//
// This is a placeholder that allows us to test the code we've written to support multiple transaction types.  It will
// be replaced with a real implementation later.
//
type EscrowOpenTransaction struct {
}

var _ TransactionBody = (*EscrowOpenTransaction)(nil)

func (tx *EscrowOpenTransaction) Size() int {
	return 0
}

func (tx *EscrowOpenTransaction) TxType() TxType {
	return TxTypeEscrowOpen
}

func (tx *EscrowOpenTransaction) MarshalTo(data []byte) (n int, err error) {
	return 0, nil
}

type TransactionInput struct {
	PreviousTx []byte // TODO: type
	Index      uint8  // (of output in PreviousTx) // TODO: type
	ScriptSig  []byte // (first half of script) // TODO: type
	SequenceNo uint32 // Normally 0xFFFFFFFF; has no effect unless the transaction has LockTime > 0.
}

func (ti *TransactionInput) Size() int {
	return 5 + len(ti.PreviousTx) + len(ti.ScriptSig)
}

func (ti *TransactionInput) MarshalTo(data []byte) (int, error) {
	if len(ti.PreviousTx) != TransactionIDSize {
		return 0, errors.New("bad size for previous transaction ID")
	}
	i := copy(data, ti.PreviousTx)

	data[i] = (byte)(ti.Index)
	i += 1

	i += binary.PutUvarint(data[i:], uint64(len(ti.ScriptSig)))
	i += copy(data[i:], ti.ScriptSig)

	binary.LittleEndian.PutUint32(data[i:], ti.SequenceNo)
	i += 4

	return i, nil
}

type TransactionOutput struct {
	Value        uint32 // (number of tokens) // TODO: type
	ScriptPubKey []byte // (second half of script) // TODO: type
}

func (to *TransactionOutput) Size() int {
	return 4 + len(to.ScriptPubKey)
}

func (to *TransactionOutput) MarshalTo(data []byte) (int, error) {
	binary.LittleEndian.PutUint32(data, to.Value)
	i := 4

	data[i] = (byte)(len(to.ScriptPubKey))
	i += copy(data[i+1:], to.ScriptPubKey) + 1

	return i, nil
}

// In Bitcoin, this is serialized as a data stack.  The number of items in the stack (2, for witness data) is given as a
// compactSize-encoded uint.  Each of the items (the signature and then the pubkey) are given a single-byte length
// prefix.
type TransactionWitness struct {
	Signature []byte
	PubKey    []byte
}

func (tw *TransactionWitness) Size() int {
	return 3 + len(tw.Signature) + len(tw.PubKey)
}

func (tw *TransactionWitness) MarshalTo(data []byte) (int, error) {
	i := binary.PutUvarint(data, 2)

	data[i] = (byte)(len(tw.Signature))
	i += copy(data[i+1:], tw.Signature) + 1

	data[i] = (byte)(len(tw.PubKey))
	i += copy(data[i+1:], tw.PubKey) + 1

	return i, nil
}
