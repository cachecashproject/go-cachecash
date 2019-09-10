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

func mustDecodeTXID(s string) TXID {
	d := testutil.MustDecodeString(s)
	var txid TXID
	if len(d) != len(txid) {
		panic("bad length for TXID")
	}
	copy(txid[:], d)
	return txid
}

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
			PreviousTx: mustDecodeTXID("deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
			Index:      0,
		},
		ScriptSig:  testutil.MustDecodeString("abc123"),
		SequenceNo: 0xFFFFFFFF,
	}

	data := make([]byte, bufferSize)
	n, err := ti.MarshalTo(data)
	assert.Nil(t, err)
	assert.Equal(t, ti.Size(), n, "MarshalTo() does not match Size()")

	var ti2 TransactionInput
	n2, err := ti2.UnmarshalFrom(data)
	assert.Nil(t, err)
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
	assert.Nil(t, err)
	assert.Equal(t, to.Size(), n, "MarshalTo() does not match Size()")

	var to2 TransactionOutput
	n2, err := to2.UnmarshalFrom(data)
	assert.Nil(t, err)
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
	assert.Nil(t, err)
	assert.Equal(t, tw.Size(), n, "MarshalTo() does not match Size()")

	var tw2 TransactionWitness
	n2, err := tw2.UnmarshalFrom(data)
	assert.Nil(t, err)
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
						PreviousTx: mustDecodeTXID("deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
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
	assert.Nil(t, err)
	assert.Equal(t, tx.Size(), n, "MarshalTo() does not match Size()")

	var tx2 Transaction
	n2, err := tx2.UnmarshalFrom(data)
	assert.Nil(t, err)
	assert.Equal(t, tx.Size(), n2, "UnmarshalFrom() does not match Size()")

	assert.Equal(t, *tx, tx2, "unmarshaled struct does not match original")
}

func (suite *TransactionTestSuite) TestTransferTransaction_InOutPoints() {
	t := suite.T()

	tx := suite.makeTransferTransaction()

	ips := tx.Inpoints()
	assert.Equal(t, []Outpoint{
		{
			PreviousTx: mustDecodeTXID("deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
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
	assert.Nil(t, err)
	assert.Equal(t, tx.Size(), n, "MarshalTo() does not match Size()")

	var tx2 Transaction
	n2, err := tx2.UnmarshalFrom(data)
	assert.Nil(t, err)
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
