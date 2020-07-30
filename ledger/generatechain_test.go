package ledger

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/ed25519"

	"github.com/cachecashproject/go-cachecash/keypair"
	"github.com/cachecashproject/go-cachecash/ledger/txscript"
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

	kpA, err := keypair.Generate()
	if err != nil {
		t.Fatalf("failed to generate keypair: %v", err)
	}
	outScrBufA, err := txscript.MakeInputScript(kpA.PublicKey)
	if err != nil {
		t.Fatalf("failed to generate output script: %v", err)
	}

	pubB, privB, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate keypair: %v", err)
	}
	outScrBufB, err := txscript.MakeInputScript(pubB)
	if err != nil {
		t.Fatalf("failed to generate output script: %v", err)
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

	// These commented lines would make this be the same as test_scriptsig - see issue #274
	//   inputScriptBytes, err := txscript.MakeOutputScript(kpA.PublicKey)
	//   assert.NoError(t, err)
	//  ... ScriptSig: inputScriptBytes // below

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
	prevOutputs := block0.Transactions.Transactions[0].Outputs()[0:1]
	err = txs[0].GenerateWitnesses(kpA, prevOutputs)
	assert.NoError(t, err)

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
	assert.Equal(t, 4, us.Length())

	// a block with a parent not in the chain

	badBlock, err := NewBlock(privCLA, MustDecodeBlockID("deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"), []*Transaction{})
	assert.NoError(t, err)

	// XXX: This should be moved to separate unit tests.
	suite.testChainDatabase([]*Block{block0, block1}, []*Block{badBlock})

	// Done.
	_ = privB
}

func (suite *GenerateChainTestSuite) testChainDatabase(blocks []*Block, badBlocks []*Block) {
	t := suite.T()
	ctx := context.Background()

	// Build chain database.
	cdb := NewDatabase(NewChainStorageMemory(blocks[0]))
	for i := 1; i < len(blocks); i++ {
		height, err := cdb.Height(ctx)
		assert.NoError(t, err)
		assert.Equal(t, uint64(i), height)
		newHeight, err := cdb.AddBlock(ctx, blocks[i])
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, uint64(i), newHeight)
	}

	tx := blocks[1].Transactions.Transactions[0]
	txid, err := tx.TXID()
	if !assert.Nil(t, err) {
		return
	}

	ccEnd := &Position{BlockID: blocks[1].BlockID(), TxIndex: 1}
	ccBeforeSpend := &Position{BlockID: blocks[1].BlockID(), TxIndex: 0}

	// The input to this transaction should be unspent when we give the correct chain.Position.
	unspent, err := cdb.Unspent(ctx, ccBeforeSpend, tx.Inpoints()[0])
	assert.Nil(t, err)
	assert.True(t, unspent)
	// See #274
	// assert.NoError(t, cdb.TransactionValid(ctx, ccBeforeSpend, tx))

	// ... but if we start one transaction later (as though we were examining a second transaction spending the same
	// input), we should find that the output has already been spent.
	unspent, err = cdb.Unspent(ctx, ccEnd, tx.Inpoints()[0])
	assert.Nil(t, err)
	assert.False(t, unspent)

	// Test that GetTransaction fetches transactions correctly.
	tx, err = cdb.GetTransaction(ctx, ccEnd, txid)
	assert.Nil(t, err)
	assert.NotNil(t, tx)

	// Test what happens when GetTransaction does not find the transaction.
	ccStart := &Position{BlockID: blocks[0].BlockID(), TxIndex: 1}
	tx, err = cdb.GetTransaction(ctx, ccStart, txid)
	assert.Nil(t, err)
	assert.Nil(t, tx)

	// bad blocks
	for _, block := range badBlocks {
		height, err := cdb.Height(ctx)
		assert.NoError(t, err)
		addHeight, err := cdb.AddBlock(ctx, block)
		assert.Error(t, err)
		assert.Equal(t, uint64(0), addHeight)
		newHeight, err := cdb.Height(ctx)
		assert.NoError(t, err)
		assert.Equal(t, height, newHeight)
	}

	// existing blocks should not succeed either
	for _, block := range blocks {
		height, err := cdb.Height(ctx)
		assert.NoError(t, err)
		addHeight, err := cdb.AddBlock(ctx, block)
		assert.Error(t, err)
		assert.Equal(t, uint64(0), addHeight)
		newHeight, err := cdb.Height(ctx)
		assert.NoError(t, err)
		assert.Equal(t, height, newHeight)
	}

}
