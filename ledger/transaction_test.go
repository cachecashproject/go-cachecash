package ledger

import (
	"testing"

	"github.com/cachecashproject/go-cachecash/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	// An arbitrary size that is much larger than any individual thing we want to marshal.
	bufferSize = 4096
)

type TransactionTestSuite struct {
	suite.Suite

	l *logrus.Logger
}

func TestTransactionTestSuite(t *testing.T) {
	suite.Run(t, new(TransactionTestSuite))
}

func (suite *TransactionTestSuite) SetupTest() {
	t := suite.T()

	l := logrus.New()
	suite.l = l

	_ = t
}

func (suite *TransactionTestSuite) TestTransactionInput_RoundTrip() {
	t := suite.T()

	ti := TransactionInput{
		Outpoint: Outpoint{
			PreviousTx: MustDecodeTXID("deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
			Index:      0,
		},
		ScriptSig:  testutil.MustDecodeString("abc123"),
		SequenceNo: 0xFFFFFFFF,
	}

	data := make([]byte, bufferSize)
	n, err := ti.MarshalTo(data)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, ti.Size(), n, "MarshalTo() does not match Size()")

	var ti2 TransactionInput
	n2, err := ti2.UnmarshalFrom(data)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, ti.Size(), n2, "UnmarshalFrom() does not match Size()")

	assert.Equal(t, ti, ti2, "unmarshaled struct does not match original")
}

func (suite *TransactionTestSuite) TestTransactionOutput_RoundTrip() {
	t := suite.T()

	to := TransactionOutput{
		Value:        1234,
		ScriptPubKey: testutil.MustDecodeString("def456"),
	}

	data := make([]byte, bufferSize)
	n, err := to.MarshalTo(data)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, to.Size(), n, "MarshalTo() does not match Size()")

	var to2 TransactionOutput
	n2, err := to2.UnmarshalFrom(data)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, to.Size(), n2, "UnmarshalFrom() does not match Size()")

	assert.Equal(t, to, to2, "unmarshaled struct does not match original")
}

func (suite *TransactionTestSuite) TestTransactionWitness_RoundTrip() {
	t := suite.T()

	tw := TransactionWitness{
		Data: [][]byte{
			testutil.MustDecodeString("abc123"),
			testutil.MustDecodeString("cafebabecafebabe"),
			testutil.MustDecodeString("def456"),
		},
	}

	data := make([]byte, bufferSize)
	n, err := tw.MarshalTo(data)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, tw.Size(), n, "MarshalTo() does not match Size()")

	var tw2 TransactionWitness
	n2, err := tw2.UnmarshalFrom(data)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, tw.Size(), n2, "UnmarshalFrom() does not match Size()")

	assert.Equal(t, tw, tw2, "unmarshaled struct does not match original")
}

func (suite *TransactionTestSuite) makeTransferTransaction() *Transaction {
	tx := &Transaction{
		Version: 0x01,
		Flags:   0x0000,
		Body: &TransferTransaction{
			Inputs: []TransactionInput{
				{
					Outpoint: Outpoint{
						PreviousTx: MustDecodeTXID("deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
						Index:      0,
					},
					ScriptSig:  testutil.MustDecodeString("abc123"),
					SequenceNo: 0xFFFFFFFF,
				},
			},
			Witnesses: []TransactionWitness{
				{
					// A zero-item stack.
				},
			},
			Outputs: []TransactionOutput{
				{
					Value:        1234,
					ScriptPubKey: testutil.MustDecodeString("def456"),
				},
			},
		},
	}

	return tx
}

func (suite *TransactionTestSuite) TestTransferTransaction_RoundTrip() {
	t := suite.T()

	tx := suite.makeTransferTransaction()

	data := make([]byte, bufferSize)
	n, err := tx.MarshalTo(data)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, tx.Size(), n, "MarshalTo() does not match Size()")

	var tx2 Transaction
	n2, err := tx2.UnmarshalFrom(data)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, tx.Size(), n2, "UnmarshalFrom() does not match Size()")

	assert.Equal(t, *tx, tx2, "unmarshaled struct does not match original")
}

func (suite *TransactionTestSuite) TestTransferTransaction_InOutPoints() {
	t := suite.T()

	tx := suite.makeTransferTransaction()

	ips := tx.Inpoints()
	assert.Equal(t, []Outpoint{
		{
			PreviousTx: MustDecodeTXID("deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
			Index:      0,
		},
	}, ips)

	txid, err := tx.TXID()
	if err != nil {
		t.Fatalf("failed to compute TXID: %v", err) // XXX: modify TXID so that it doesn't return an error
	}

	ops := tx.Outpoints()
	assert.Equal(t, []Outpoint{
		{
			PreviousTx: txid,
			Index:      0,
		},
	}, ops)
}

func (suite *TransactionTestSuite) makeGenesisTransaction() *Transaction {
	tx := &Transaction{
		Version: 0x01,
		Flags:   0x0000,
		Body: &GenesisTransaction{
			Outputs: []TransactionOutput{
				{
					Value:        42,
					ScriptPubKey: testutil.MustDecodeString("abc123"),
				},
				{
					Value:        123,
					ScriptPubKey: testutil.MustDecodeString("def456"),
				},
			},
		},
	}

	return tx
}

func (suite *TransactionTestSuite) TestGenesisTransaction_RoundTrip() {
	t := suite.T()

	tx := suite.makeGenesisTransaction()

	data := make([]byte, bufferSize)
	n, err := tx.MarshalTo(data)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, tx.Size(), n, "MarshalTo() does not match Size()")

	var tx2 Transaction
	n2, err := tx2.UnmarshalFrom(data)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, tx.Size(), n2, "UnmarshalFrom() does not match Size()")

	assert.Equal(t, *tx, tx2, "unmarshaled struct does not match original")
}

func (suite *TransactionTestSuite) TestGenesisTransaction_InOutPoints() {
	t := suite.T()

	tx := suite.makeGenesisTransaction()

	ips := tx.Inpoints()
	assert.Equal(t, 0, len(ips))

	txid, err := tx.TXID()
	if err != nil {
		t.Fatalf("failed to compute TXID: %v", err) // XXX: modify TXID so that it doesn't return an error
	}

	ops := tx.Outpoints()
	assert.Equal(t, []Outpoint{
		{
			PreviousTx: txid,
			Index:      0,
		},
		{
			PreviousTx: txid,
			Index:      1,
		},
	}, ops)
}

func (suite *TransactionTestSuite) TestMarshal() {
	t := suite.T()

	transactions := &Transactions{Transactions: []*Transaction{
		{
			Version: 0x01,
			Flags:   0x0000,
			Body: &TransferTransaction{
				Inputs: []TransactionInput{
					{
						Outpoint: Outpoint{
							PreviousTx: MustDecodeTXID("deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
							Index:      0,
						},
						ScriptSig:  testutil.MustDecodeString("abc123"),
						SequenceNo: 0xFFFFFFFF,
					},
				},
				Witnesses: []TransactionWitness{
					{
						// A zero-item stack.
					},
				},
				Outputs: []TransactionOutput{
					{
						Value:        1234,
						ScriptPubKey: testutil.MustDecodeString("def456"),
					},
				},
			},
		},
	},
	}

	// These are the protobuf serialised bytes for /
	transactionsBytes := []byte{
		0x38, 0x0, 0x0, 0x0, 0x1, 0x1, 0x0, 0x0, 0x1, 0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe,
		0xef, 0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef, 0xde,
		0xad, 0xbe, 0xef, 0x0, 0x3, 0xab, 0xc1, 0x23, 0xff, 0xff, 0xff, 0xff, 0x1, 0xd2, 0x4, 0x0, 0x0, 0x3, 0xde, 0xf4,
		0x56, 0x0}

	bytes, err := transactions.Marshal()
	assert.Nil(t, err)
	assert.Equal(t, transactionsBytes, bytes)

	transactions2 := &Transactions{}
	err = transactions2.Unmarshal(transactionsBytes)
	assert.Nil(t, err)
	assert.Equal(t, transactions, transactions2)
}

func (suite *TransactionTestSuite) makeGlobalConfigTransaction() *Transaction {

	// XXX: Need to replace these with actually-valid values.
	sigPublicKey := make([]byte, 32)
	signature := make([]byte, 64)

	// N.b.: We include empty insertion/deletion lists here because our UnmarshalFrom implementations produce empty
	// lists (instead of leaving the slice nil), and `assert.Equal` does not consider nil and an empty slice to be equal
	// even though they behave similarly in most situations.
	tx := &Transaction{
		Version: 0x01,
		Flags:   0x0000,
		Body: &GlobalConfigTransaction{
			ActivationBlockHeight: 7890,
			ScalarUpdates: []GlobalConfigScalarUpdate{
				{Key: "ScalarA", Value: []byte("abc")},
				{Key: "ScalarB", Value: []byte("")},
				{Key: "ScalarC", Value: []byte("quick red fox")},
			},
			ListUpdates: []GlobalConfigListUpdate{
				{Key: "ListA", Deletions: []uint64{0, 1, 2}, Insertions: []GlobalConfigListInsertion{
					{0, []byte("foo")},
					{0, []byte("bar")},
					{1, []byte("baz")},
				}},
				{Key: "ListB", Deletions: []uint64{}, Insertions: []GlobalConfigListInsertion{
					{42, []byte("appended value")},
				}},
				{Key: "ListC", Deletions: []uint64{5, 7, 11}, Insertions: []GlobalConfigListInsertion{}},
			},
			SigPublicKey: sigPublicKey,
			Signature:    signature,
		},
	}

	return tx
}

func (suite *TransactionTestSuite) TestGlobalConfigTransaction_RoundTrip() {
	t := suite.T()

	tx := suite.makeGlobalConfigTransaction()

	data := make([]byte, bufferSize)
	n, err := tx.MarshalTo(data)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, tx.Size(), n, "MarshalTo() does not match Size()")

	var tx2 Transaction
	n2, err := tx2.UnmarshalFrom(data)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, tx.Size(), n2, "UnmarshalFrom() does not match Size()")

	assert.Equal(t, *tx, tx2, "unmarshaled struct does not match original")
}
