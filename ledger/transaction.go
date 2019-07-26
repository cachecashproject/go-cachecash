package ledger

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"

	"github.com/pkg/errors"
)

type TxType uint8

const (
	TxTypeUnknown    TxType = 0x00 // Not valid in serialized transactions.
	TxTypeTransfer   TxType = 0x01
	TxTypeGenesis    TxType = 0x02
	TxTypeEscrowOpen TxType = 0x03
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
	_, err := tx.UnmarshalFrom(data)
	return err
}

// N.B.: This is not strictly required for the protobuf interface, but it's useful for test code to be able to tell how
// many bytes were consumed.
func (tx *Transaction) UnmarshalFrom(data []byte) (int, error) {
	tx.Version = data[0]
	if tx.Version != 1 {
		return 0, errors.New("unexpected transaction version")
	}
	txType := (TxType)(data[1])
	tx.Flags = binary.LittleEndian.Uint16(data[2:])
	if tx.Flags != 0 {
		return 0, errors.New("unexpected transaction flags")
	}

	switch txType {
	case TxTypeTransfer:
		tx.Body = &TransferTransaction{}
	case TxTypeGenesis:
		tx.Body = &GenesisTransaction{}
	case TxTypeEscrowOpen:
		tx.Body = &EscrowOpenTransaction{}
	default:
		return 0, errors.New("unexpected transaction type")
	}

	ni, err := tx.Body.UnmarshalFrom(data[4:])
	if err != nil {
		return 0, err
	}
	return ni + 4, nil
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

func (tx *Transaction) TXID() ([]byte, error) {
	data, err := tx.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal transaction")
	}

	d := sha256.Sum256(data)
	d = sha256.Sum256(d[:])
	return d[:], nil
}

func (tx *Transaction) Inpoints() []Outpoint {
	return tx.Body.Inpoints()
}

func (tx *Transaction) Outpoints() []Outpoint {
	txid, err := tx.TXID()
	if err != nil {
		panic(err) // XXX: We should change TXID() so that it doesn't return an error.
	}

	var pp []Outpoint
	for i := uint8(0); i < tx.Body.OutputCount(); i++ {
		pp = append(pp, Outpoint{
			PreviousTx: txid,
			Index:      i,
		})
	}
	return pp
}

type TransactionBody interface {
	Size() int
	TxType() TxType
	MarshalTo(data []byte) (n int, err error)
	Unmarshal(data []byte) error
	UnmarshalFrom(data []byte) (n int, err error)
	Inpoints() []Outpoint
	OutputCount() uint8
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
	// We don't need to include len(tx.Witnesses) because it must be equal to len(tx.Inputs).
	n := UvarintSize(uint64(len(tx.Inputs))) + UvarintSize(uint64(len(tx.Outputs)))
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
	var n int

	if len(tx.Inputs) != len(tx.Witnesses) {
		return 0, errors.New("number of witnesses must match number of inputs")
	}

	n += binary.PutUvarint(data, uint64(len(tx.Inputs)))

	// XXX: There's quite a bit of repetition here.  Refactor?
	for _, o := range tx.Inputs {
		ni, err := o.MarshalTo(data[n:])
		if err != nil {
			return 0, errors.Wrap(err, "failed to marshal TransactionInput")
		}
		n += ni
	}

	n += binary.PutUvarint(data[n:], uint64(len(tx.Outputs)))

	for _, o := range tx.Outputs {
		ni, err := o.MarshalTo(data[n:])
		if err != nil {
			return 0, errors.Wrap(err, "failed to marshal TransactionOutput")
		}
		n += ni
	}

	for _, o := range tx.Witnesses {
		ni, err := o.MarshalTo(data[n:])
		if err != nil {
			return 0, errors.Wrap(err, "failed to marshal TransactionWitness")
		}
		n += ni
	}

	return n, nil
}

func (tx *TransferTransaction) Unmarshal(data []byte) error {
	_, err := tx.UnmarshalFrom(data)
	return err
}

func (tx *TransferTransaction) UnmarshalFrom(data []byte) (int, error) {
	inputQty, n := binary.Uvarint(data)
	tx.Inputs = make([]TransactionInput, inputQty)
	tx.Witnesses = make([]TransactionWitness, inputQty)

	for i := 0; i < len(tx.Inputs); i++ {
		ni, err := tx.Inputs[i].UnmarshalFrom(data[n:])
		if err != nil {
			return 0, errors.Wrap(err, "failed to unmarshal TransactionInput")
		}
		n += ni
	}

	outputQty, ni := binary.Uvarint(data[n:])
	n += ni
	tx.Outputs = make([]TransactionOutput, outputQty)

	for i := 0; i < len(tx.Outputs); i++ {
		ni, err := tx.Outputs[i].UnmarshalFrom(data[n:])
		if err != nil {
			return 0, errors.Wrap(err, "failed to unmarshal TransactionOutput")
		}
		n += ni
	}

	for i := 0; i < len(tx.Witnesses); i++ {
		ni, err := tx.Witnesses[i].UnmarshalFrom(data[n:])
		if err != nil {
			return 0, errors.Wrap(err, "failed to unmarshal TransactionWitness")
		}
		n += ni
	}

	return n, nil
}

func (tx *TransferTransaction) Inpoints() []Outpoint {
	var pp []Outpoint
	for _, ti := range tx.Inputs {
		pp = append(pp, Outpoint{
			PreviousTx: ti.PreviousTx,
			Index:      ti.Index,
		})
	}
	return pp
}

func (tx *TransferTransaction) OutputCount() uint8 {
	return uint8(len(tx.Outputs))
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

func (tx *EscrowOpenTransaction) Unmarshal(data []byte) error {
	return nil
}

func (tx *EscrowOpenTransaction) UnmarshalFrom(data []byte) (n int, err error) {
	return 0, nil
}

func (tx *EscrowOpenTransaction) Inpoints() []Outpoint {
	return nil
}

func (tx *EscrowOpenTransaction) OutputCount() uint8 {
	return 0
}

// XXX: This is a bad name, because we use this struct to describe both inpoints and outpoints.
type Outpoint struct {
	PreviousTx []byte // TODO: type
	Index      uint8  // (of output in PreviousTx) // TODO: type
}

type OutpointKey [33]byte

func (o *Outpoint) Key() OutpointKey {
	var k OutpointKey
	copy(k[:], o.PreviousTx[:])
	k[32] = o.Index
	return k
}

type TransactionInput struct {
	Outpoint
	ScriptSig  []byte // (first half of script) // TODO: type
	SequenceNo uint32 // Normally 0xFFFFFFFF; has no effect unless the transaction has LockTime > 0.
}

func (ti *TransactionInput) Size() int {
	// N.B.: PreviousTx is fixed-length
	return 5 + len(ti.PreviousTx) + int(UvarintSize(uint64(len(ti.ScriptSig)))) + len(ti.ScriptSig)
}

func (ti *TransactionInput) MarshalTo(data []byte) (int, error) {
	if len(ti.PreviousTx) != TransactionIDSize {
		return 0, errors.New("bad size for previous transaction ID")
	}
	n := copy(data, ti.PreviousTx)

	data[n] = (byte)(ti.Index)
	n += 1

	n += binary.PutUvarint(data[n:], uint64(len(ti.ScriptSig)))
	n += copy(data[n:], ti.ScriptSig)

	binary.LittleEndian.PutUint32(data[n:], ti.SequenceNo)
	n += 4

	return n, nil
}

func (ti *TransactionInput) UnmarshalFrom(data []byte) (int, error) {
	var n int

	ti.PreviousTx = data[n : n+TransactionIDSize]
	n += TransactionIDSize

	ti.Index = uint8(data[n])
	n += 1

	fieldLen, ni := binary.Uvarint(data[n:])
	n += ni
	ti.ScriptSig = data[n : n+int(fieldLen)]
	n += int(fieldLen)

	ti.SequenceNo = binary.LittleEndian.Uint32(data[n:])
	n += 4

	return n, nil
}

type TransactionOutput struct {
	Value        uint32 // (number of tokens) // TODO: type
	ScriptPubKey []byte // (second half of script) // TODO: type
}

func (to *TransactionOutput) Size() int {
	return 4 + int(UvarintSize(uint64(len(to.ScriptPubKey)))) + len(to.ScriptPubKey)
}

func (to *TransactionOutput) MarshalTo(data []byte) (int, error) {
	binary.LittleEndian.PutUint32(data, to.Value)
	i := 4

	i += binary.PutUvarint(data[i:], uint64(len(to.ScriptPubKey)))
	i += copy(data[i:], to.ScriptPubKey)

	return i, nil
}

func (to *TransactionOutput) UnmarshalFrom(data []byte) (int, error) {
	to.Value = binary.LittleEndian.Uint32(data)
	n := 4

	fieldLen, ni := binary.Uvarint(data[n:])
	n += ni
	to.ScriptPubKey = data[n : n+int(fieldLen)]
	n += int(fieldLen)

	return n, nil
}

// In Bitcoin, this is serialized as a data stack.  The number of items in the stack (2, for witness data) is given as a
// compactSize-encoded uint.  Each of the items (the signature and then the pubkey) are given a single-byte length
// prefix.
//
// In Cachecash, we use a uvarint for the number of items, and then each item has a uvarint length prefix.
//
type TransactionWitness struct {
	Data [][]byte
}

func (tw *TransactionWitness) Size() int {
	n := UvarintSize(uint64(len(tw.Data)))
	for _, d := range tw.Data {
		n += UvarintSize(uint64(len(d))) + len(d)
	}
	return n
}

func (tw *TransactionWitness) MarshalTo(data []byte) (int, error) {
	n := binary.PutUvarint(data, uint64(len(tw.Data)))

	for _, d := range tw.Data {
		n += binary.PutUvarint(data[n:], uint64(len(d)))
		n += copy(data[n:], d)
	}

	return n, nil
}

func (tw *TransactionWitness) UnmarshalFrom(data []byte) (int, error) {
	stackSize, n := binary.Uvarint(data)

	tw.Data = nil
	for i := uint64(0); i < stackSize; i++ {
		fieldLen, ni := binary.Uvarint(data[n:])
		n += ni
		tw.Data = append(tw.Data, data[n:n+int(fieldLen)])
		n += int(fieldLen)
	}

	return n, nil
}

// A GenesisTransaction creates coins from thin air.  They are only valid in the genesis block (block 0).  Because we do
// not have coinbase transactions, we need an explicit way to get coins into the system.
type GenesisTransaction struct {
	Outputs []TransactionOutput
}

var _ TransactionBody = (*GenesisTransaction)(nil)

func (tx *GenesisTransaction) Size() int {
	n := UvarintSize(uint64(len(tx.Outputs)))
	for _, o := range tx.Outputs {
		n += o.Size()
	}
	return n
}

func (tx *GenesisTransaction) TxType() TxType {
	return TxTypeGenesis
}

func (tx *GenesisTransaction) MarshalTo(data []byte) (int, error) {
	var n int

	n += binary.PutUvarint(data[n:], uint64(len(tx.Outputs)))

	for _, o := range tx.Outputs {
		ni, err := o.MarshalTo(data[n:])
		if err != nil {
			return 0, errors.Wrap(err, "failed to marshal TransactionOutput")
		}
		n += ni
	}

	return n, nil
}

func (tx *GenesisTransaction) Unmarshal(data []byte) error {
	_, err := tx.UnmarshalFrom(data)
	return err
}

func (tx *GenesisTransaction) UnmarshalFrom(data []byte) (int, error) {
	var n int

	outputQty, ni := binary.Uvarint(data[n:])
	n += ni
	tx.Outputs = make([]TransactionOutput, outputQty)

	for i := 0; i < len(tx.Outputs); i++ {
		ni, err := tx.Outputs[i].UnmarshalFrom(data[n:])
		if err != nil {
			return 0, errors.Wrap(err, "failed to unmarshal TransactionOutput")
		}
		n += ni
	}

	return n, nil
}

func (tx *GenesisTransaction) Inpoints() []Outpoint {
	return nil
}

func (tx *GenesisTransaction) OutputCount() uint8 {
	return uint8(len(tx.Outputs))
}
