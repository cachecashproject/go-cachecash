package ledger

import (
	"testing"

	"github.com/cachecashproject/go-cachecash/ledger/txscript"
	"github.com/cachecashproject/go-cachecash/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/ed25519"
)

type GenerateChainTestSuite struct {
	suite.Suite

	l *logrus.Logger
}

func TestGenerateChainTestSuite(t *testing.T) {
	suite.Run(t, new(GenerateChainTestSuite))
}

func (suite *GenerateChainTestSuite) SetupTest() {
	t := suite.T()

	l := logrus.New()
	suite.l = l

	_ = t
}

func (suite *GenerateChainTestSuite) TestGenerateChain() {
	t := suite.T()

	pubCLA, privCLA, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate CLA keypair: %v", err)
	}
	_ = pubCLA

	pubA, privA, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate keypair: %v", err)
	}
	addrA := MakeP2WPKHAddress(pubA)
	outScrA, err := txscript.MakeP2WPKHOutputScript(addrA.PublicKeyHash)
	if err != nil {
		t.Fatalf("failed to generate output script: %v", err)
	}
	outScrBufA, err := outScrA.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal output script: %v", err)
	}

	pubB, privB, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate keypair: %v", err)
	}
	addrB := MakeP2WPKHAddress(pubB)
	outScrB, err := txscript.MakeP2WPKHOutputScript(addrB.PublicKeyHash)
	if err != nil {
		t.Fatalf("failed to generate output script: %v", err)
	}
	outScrBufB, err := outScrB.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal output script: %v", err)
	}

	// Block 0

	txs := []*Transaction{
		{
			Version: 0x01,
			Flags:   0x0000,
			Body: &GenesisTransaction{
				Outputs: []TransactionOutput{
					{
						Value:        100,
						ScriptPubKey: outScrBufA,
					},
					{
						Value:        100,
						ScriptPubKey: outScrBufA,
					},
					{
						Value:        100,
						ScriptPubKey: outScrBufA,
					},
				},
			},
		},
	}

	// bidparent := BlockID{}
	// assert.True(t, bidparent.Zero())

	block0, err := NewBlock(privCLA, BlockID{}, txs)
	assert.Nil(t, err, "failed to create genesis block")
	_ = block0

	// Block 1

	prevTXID, err := block0.Transactions.Transactions[0].TXID()
	if err != nil {
		t.Fatalf("failed to get previous transaction ID: %v", err)
	}

	txs = []*Transaction{
		{
			Version: 0x01,
			Flags:   0x0000,
			Body: &TransferTransaction{
				Inputs: []TransactionInput{
					{
						Outpoint:   Outpoint{PreviousTx: prevTXID, Index: 0},
						ScriptSig:  nil, // N.B.: For P2WPKH transactions, this must always be empty.
						SequenceNo: 0xFFFFFFFF,
					},
				},
				Witnesses: []TransactionWitness{
					{
						Data: [][]byte{
							testutil.MustDecodeString("cafebabe"), // XXX: Replace once we have sighash.
							pubA,
						},
					},
				},
				Outputs: []TransactionOutput{
					{ // Send 10 coins to B.
						Value:        10,
						ScriptPubKey: outScrBufB,
					},
					{ // Change.
						Value:        90,
						ScriptPubKey: outScrBufA,
					},
				},
			},
		},
	}

	block1, err := NewBlock(privCLA, block0.BlockID(), txs)
	assert.Nil(t, err, "failed to create genesis block")
	_ = block1

	// Validate & compute UTXO set
	// TODO: Validate.
	us := NewUTXOSet()
	for _, blk := range []*Block{block0, block1} {
		for _, tx := range blk.Transactions.Transactions {
			if err := us.Update(tx); err != nil {
				t.Fatalf("failed to update UTXO set: %v", err)
			}
		}
	}
	assert.Equal(t, 4, len(us.utxos))

	// XXX: This should be moved to separate unit tests.
	suite.testChainDatabase([]*Block{block0, block1})

	// Done.
	_ = privA
	_ = privB
}

func (suite *GenerateChainTestSuite) testChainDatabase(blocks []*Block) {
	t := suite.T()

	// Build chain database.
	cdb, err := NewSimpleChainDatabase(blocks[0])
	if !assert.Nil(t, err) {
		return
	}
	for i := 1; i < len(blocks); i++ {
		if !assert.Nil(t, cdb.AddBlock(blocks[i])) {
			return
		}
	}

	tx := blocks[1].Transactions.Transactions[0]
	txid, err := tx.TXID()
	if !assert.Nil(t, err) {
		return
	}

	ccEnd := &ChainContext{BlockID: blocks[1].BlockID(), TxIndex: 1}

	// The input to this transaction should be unspent when we give the correct ChainContext.
	unspent, err := cdb.Unspent(&ChainContext{BlockID: blocks[1].BlockID(), TxIndex: 0}, tx.Inpoints()[0])
	assert.Nil(t, err)
	assert.True(t, unspent)

	// ... but if we start one transaction later (as though we were examining a second transaction spending the same
	// input), we should find that the output has already been spent.
	unspent, err = cdb.Unspent(ccEnd, tx.Inpoints()[0])
	assert.Nil(t, err)
	assert.False(t, unspent)

	// Test that GetTransaction fetches transactions correctly.
	tx, err = cdb.GetTransaction(ccEnd, txid)
	assert.Nil(t, err)
	assert.NotNil(t, tx)

	// Test what happens when GetTransaction does not find the transaction.
	tx, err = cdb.GetTransaction(ccEnd, MustDecodeTXID("deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"))
	assert.Nil(t, err)
	assert.Nil(t, tx)

	_ = cdb
}
