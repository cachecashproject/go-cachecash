package ledgerservice

import (
	"context"
	"testing"

	"github.com/cachecashproject/go-cachecash/keypair"
	"github.com/cachecashproject/go-cachecash/ledger"
	"github.com/cachecashproject/go-cachecash/ledger/txscript"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ed25519"
)

func SetupGenesis(kp *keypair.KeyPair) (*LedgerMiner, *ledger.Block, error) {
	storage := NewLedgerMemory()
	lm, err := NewLedgerMiner(logrus.New(), storage, kp)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()
	block, err := lm.InitGenesisBlock(ctx, 420000000)
	if err != nil {
		return nil, nil, err
	}

	return lm, block, nil
}

func TestMineNoBlockWithEmptyMempool(t *testing.T) {
	ctx := context.Background()

	kp, err := keypair.Generate()
	assert.Nil(t, err)

	lm, block, err := SetupGenesis(kp)
	assert.Nil(t, err)
	assert.NotNil(t, block)

	block, err = lm.GenerateBlock(ctx)
	assert.Nil(t, err)
	assert.Nil(t, block)
}

func makeInputs(kp *keypair.KeyPair, block *ledger.Block) ([]ledger.TransactionInput, []ledger.TransactionOutput, error) {
	inputs := []ledger.TransactionInput{}
	prevOutputs := []ledger.TransactionOutput{}

	for _, tx := range block.Transactions {
		for idx, txo := range tx.Outputs() {
			inputScriptBytes, err := txscript.MakeOutputScript(kp.PublicKey)
			if err != nil {
				return nil, nil, errors.Wrap(err, "failed to create output script")
			}

			txid, err := tx.TXID()
			if err != nil {
				return nil, nil, err
			}

			inputs = append(inputs, ledger.TransactionInput{
				Outpoint: ledger.Outpoint{
					PreviousTx: txid,
					Index:      uint8(idx),
				},
				ScriptSig:  inputScriptBytes,
				SequenceNo: 0xFFFFFFFF,
			})
			prevOutputs = append(prevOutputs, txo)
		}
	}

	return inputs, prevOutputs, nil
}

func makeOutputs(target ed25519.PublicKey, amount uint32) ([]ledger.TransactionOutput, error) {
	outputs := []ledger.TransactionOutput{}

	outputScriptBytes, err := txscript.MakeInputScript(target)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create input script")
	}

	txo := ledger.TransactionOutput{
		Value:        amount,
		ScriptPubKey: outputScriptBytes,
	}
	outputs = append(outputs, txo)

	return outputs, nil
}

func makeTx(kp *keypair.KeyPair, block *ledger.Block, target ed25519.PublicKey, amount uint32) (*ledger.Transaction, error) {
	inputs, prevOutputs, err := makeInputs(kp, block)
	if err != nil {
		return nil, err
	}
	outputs, err := makeOutputs(target, amount)
	if err != nil {
		return nil, err
	}

	tx := &ledger.Transaction{
		Version: 1,
		Flags:   0,
		Body: &ledger.TransferTransaction{
			Inputs:   inputs,
			Outputs:  outputs,
			LockTime: 0,
		},
	}
	err = tx.GenerateWitnesses(kp, prevOutputs)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func TestMineOneTransaction(t *testing.T) {
	ctx := context.Background()

	kp, err := keypair.Generate()
	assert.Nil(t, err)

	lm, block, err := SetupGenesis(kp)
	assert.Nil(t, err)
	assert.NotNil(t, block)

	tx, err := makeTx(kp, block, kp.PublicKey, 420000000)
	assert.Nil(t, err)

	err = lm.QueueTX(ctx, tx)
	assert.Nil(t, err)

	block, err = lm.GenerateBlock(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, block)
}

func TestMineOutputTwice(t *testing.T) {
	ctx := context.Background()

	kp, err := keypair.Generate()
	assert.Nil(t, err)

	lm, block, err := SetupGenesis(kp)
	assert.Nil(t, err)
	assert.NotNil(t, block)

	tx, err := makeTx(kp, block, kp.PublicKey, 420000000)
	assert.Nil(t, err)

	err = lm.QueueTX(ctx, tx)
	assert.Nil(t, err)

	block, err = lm.GenerateBlock(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, block)

	// try to spend the same output twice
	err = lm.QueueTX(ctx, tx)
	assert.Nil(t, err)

	// verify this fails
	block, err = lm.GenerateBlock(ctx)
	assert.Nil(t, err)
	assert.Nil(t, block)
}

func TestMineConflictingTXs(t *testing.T) {
	ctx := context.Background()

	kp, err := keypair.Generate()
	assert.Nil(t, err)

	lm, block, err := SetupGenesis(kp)
	assert.Nil(t, err)
	assert.NotNil(t, block)

	tx, err := makeTx(kp, block, kp.PublicKey, 420000000)
	assert.Nil(t, err)

	// only one of them is going to get mined
	err = lm.QueueTX(ctx, tx)
	assert.Nil(t, err)
	err = lm.QueueTX(ctx, tx)
	assert.Nil(t, err)

	block, err = lm.GenerateBlock(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, block)

	assert.Equal(t, 1, len(block.Transactions))
}
