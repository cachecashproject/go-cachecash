package ledger

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/cachecashproject/go-cachecash/ledger/txscript"
)

// <---- Transactions ----

// Protobuf custom type glue for dealing with bug https://github.com/gogo/protobuf/issues/478

// Transactions wraps a slice of *Transaction
type Transactions struct {
	Transactions []*Transaction
}

func (transactions Transactions) Marshal() ([]byte, error) {
	s := transactions.Size()
	data := make([]byte, s)
	n, err := transactions.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	if n != len(data) {
		return nil, errors.New("unexpected data length in BlockHeader.Marshal()")
	}
	return data, nil
}

func (transactions *Transactions) MarshalTo(data []byte) (int, error) {
	var n int

	for _, tx := range transactions.Transactions {
		txBytes, err := tx.Marshal()
		if err != nil {
			return 0, err
		}
		binary.LittleEndian.PutUint32(data[n:], uint32(len(txBytes)))
		n += 4
		n += copy(data[n:], txBytes)
	}
	return n, nil
}

func (transactions *Transactions) Unmarshal(data []byte) error {
	_, err := transactions.UnmarshalFrom(data)
	return err
}

func (transactions *Transactions) UnmarshalFrom(data []byte) (int, error) {
	var n int
	transactions.Transactions = make([]*Transaction, 0)
	for len(data[n:]) > 0 {
		if len(data[n:]) < 4 {
			return 0, errors.New("incomplete tx length field")
		}
		b := int(binary.LittleEndian.Uint32(data[n:]))
		n += 4

		if len(data[n:]) < b {
			return 0, errors.New("transaction length field exceeds remaining data")
		}

		tx := Transaction{}
		err := tx.Unmarshal(data[n : n+b])
		if err != nil {
			return 0, errors.Wrap(err, "failed to unmarshal transaction")
		}
		n += b

		transactions.Transactions = append(transactions.Transactions, &tx)
	}

	return n, nil
}

func (transactions *Transactions) Size() int {
	var n int
	for _, tx := range transactions.Transactions {
		n += 4 + tx.Size()
	}
	return n
}

// ---- Transactions ---->

type TxType uint8

const (
	TxTypeUnknown    TxType = 0x00 // Not valid in serialized transactions.
	TxTypeTransfer   TxType = 0x01
	TxTypeGenesis    TxType = 0x02
	TxTypeEscrowOpen TxType = 0x03

	// XXX: this is an arbitrary limit to prevent a panic in make()
	MAX_INPUTS  = 512
	MAX_OUTPUTS = 512
	// this limit is identical with bitcoin
	MAX_FIELDLEN = 520
)

// In order to be used as a gogo/protobuf custom type, a struct must implement this interface...
// - removed JSON from this because its not actually needed: its there because protobuf invokes JSON serialisation
// on the types and the type thus has to be JSON serialisable. IFF the default behaviour is inappropriate do we need
// a custom implementation. If we do need to add it back in, use type aliases - http://choly.ca/post/go-json-marshalling/
type protobufCustomType interface {
	Marshal() ([]byte, error)
}

// ... and the pointer-to-struct type must implement this one.
type protobufCustomTypePtr interface {
	MarshalTo(data []byte) (n int, err error)
	Unmarshal(data []byte) error
	Size() int
}

type Transaction struct {
	Version uint8  // Must be 1.
	Flags   uint16 // Must be zero; no flags are defined for any transaction types.
	Body    TransactionBody
}

var _ protobufCustomType = (*Transaction)(nil)
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
	if len(data) < 4 {
		return 0, errors.New("incomplete transaction fields")
	}

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

func (tx *Transaction) TXID() (TXID, error) {
	data, err := tx.Marshal()
	if err != nil {
		return TXID{}, errors.Wrap(err, "failed to marshal transaction")
	}

	d := sha256.Sum256(data)
	d = sha256.Sum256(d[:])
	return d, nil
}

func (tx *Transaction) Inpoints() []Outpoint {
	return tx.Body.Inpoints()
}

func (tx *Transaction) Inputs() []TransactionInput {
	return tx.Body.TxInputs()
}

func (tx *Transaction) Outputs() []TransactionOutput {
	return tx.Body.TxOutputs()
}

func (tx *Transaction) Witnesses() []TransactionWitness {
	return tx.Body.TxWitnesses()
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

func (tx *Transaction) WellFormed() error {
	// XXX: Add validation logic here.
	return nil
}

func (tx *Transaction) Standard() error {
	// Check that input and output scripts are standard.
	for _, ti := range tx.Inputs() {
		scr, err := txscript.ParseScript(ti.ScriptSig)
		if err != nil {
			return errors.Wrap(err, "failed to parse script")
		}
		if err := scr.StandardInput(); err != nil {
			return errors.Wrap(err, "input script is not standard")
		}
	}
	// TODO: Do we also need to check that there are two witness values for each input?
	for _, to := range tx.Outputs() {
		scr, err := txscript.ParseScript(to.ScriptPubKey)
		if err != nil {
			return errors.Wrap(err, "failed to parse script")
		}
		if err := scr.StandardOutput(); err != nil {
			return errors.Wrap(err, "output script is not standard")
		}
	}

	return nil
}

type TransactionBody interface {
	Size() int
	TxType() TxType
	MarshalTo(data []byte) (n int, err error)
	Unmarshal(data []byte) error
	UnmarshalFrom(data []byte) (n int, err error)
	Inpoints() []Outpoint
	OutputCount() uint8
	TxInputs() []TransactionInput
	TxOutputs() []TransactionOutput
	TxWitnesses() []TransactionWitness
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
	if n <= 0 {
		return 0, errors.New("failed to read inputQty")
	}

	if inputQty > MAX_INPUTS {
		return 0, errors.New("exceeded maximum number of inputs")
	}
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
	if ni <= 0 {
		return 0, errors.New("failed to read outputQty")
	}
	n += ni

	if outputQty > MAX_OUTPUTS {
		return 0, errors.New("exceeded maximum number of outputs")
	}
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

func (tx *TransferTransaction) TxInputs() []TransactionInput {
	return tx.Inputs
}

func (tx *TransferTransaction) TxOutputs() []TransactionOutput {
	return tx.Outputs
}

func (tx *TransferTransaction) TxWitnesses() []TransactionWitness {
	return tx.Witnesses
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

func (tx *EscrowOpenTransaction) TxInputs() []TransactionInput {
	return nil
}

func (tx *EscrowOpenTransaction) TxOutputs() []TransactionOutput {
	return nil
}

func (tx *EscrowOpenTransaction) TxWitnesses() []TransactionWitness {
	return nil
}

// XXX: This is a bad name, because we use this struct to describe both inpoints and outpoints.
type Outpoint struct {
	PreviousTx TXID
	Index      uint8 // (of output in PreviousTx) // TODO: type
}

func (a Outpoint) Equal(b Outpoint) bool {
	return a.PreviousTx.Equal(b.PreviousTx) && a.Index == b.Index
}

func (o *Outpoint) Key() OutpointKey {
	var k OutpointKey
	copy(k[:], o.PreviousTx[:])
	k[32] = o.Index
	return k
}

type OutpointKey [33]byte

func NewOutpointKey(txid []byte, idx byte) (*OutpointKey, error) {
	if len(txid) != 32 {
		return nil, errors.New("txid has wrong length")
	}

	outpoint := OutpointKey{}
	copy(outpoint[:], txid)
	outpoint[32] = idx

	return &outpoint, nil
}

func (o *OutpointKey) TXID() []byte {
	return o[:32]
}

func (o *OutpointKey) Idx() byte {
	return o[32]
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
	n := copy(data, ti.PreviousTx[:])

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

	if len(data[n:]) < TransactionIDSize+1 {
		return 0, errors.New("TransactionInput is below minimum length")
	}

	n += copy(ti.PreviousTx[:], data[n:n+TransactionIDSize])

	ti.Index = uint8(data[n])
	n += 1

	fieldLen, ni := binary.Uvarint(data[n:])
	if ni <= 0 {
		return 0, errors.New("field to read fieldLen")
	}
	n += ni
	if fieldLen > MAX_FIELDLEN {
		return 0, errors.New("fieldLen exceeds MAX_FIELDLEN")
	}
	if len(data[n:]) < int(fieldLen) {
		return 0, errors.New("fieldLen exceeds data buffer")
	}
	ti.ScriptSig = data[n : n+int(fieldLen)]
	n += int(fieldLen)

	if len(data[n:]) < 4 {
		return 0, errors.New("failed to read SequenceNo")
	}
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
	if len(data) < 4 {
		return 0, errors.New("failed to read TransactionOutput value")
	}
	to.Value = binary.LittleEndian.Uint32(data)
	n := 4

	fieldLen, ni := binary.Uvarint(data[n:])
	if ni <= 0 {
		return 0, errors.New("field to read fieldLen")
	}
	n += ni
	if fieldLen > MAX_FIELDLEN {
		return 0, errors.New("fieldLen exceeds MAX_FIELDLEN")
	}
	if len(data[n:]) < int(fieldLen) {
		return 0, errors.New("fieldLen exceeds data buffer")
	}
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
	if n <= 0 {
		return 0, errors.New("field to read stackSize")
	}

	tw.Data = nil
	for i := uint64(0); i < stackSize; i++ {
		fieldLen, ni := binary.Uvarint(data[n:])
		if ni <= 0 {
			return 0, errors.New("field to read fieldLen")
		}
		n += ni
		if fieldLen > MAX_FIELDLEN {
			return 0, errors.New("fieldLen exceeds MAX_FIELDLEN")
		}
		if len(data[n:]) < int(fieldLen) {
			return 0, errors.New("failed to read witness data")
		}
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
	if ni <= 0 {
		return 0, errors.New("field to read outputQty")
	}
	n += ni

	if outputQty > MAX_OUTPUTS {
		return 0, errors.New("exceeded maximum number of outputs")
	}
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

func (tx *GenesisTransaction) TxInputs() []TransactionInput {
	return nil
}

func (tx *GenesisTransaction) TxOutputs() []TransactionOutput {
	return tx.Outputs
}

func (tx *GenesisTransaction) TxWitnesses() []TransactionWitness {
	return nil
}
