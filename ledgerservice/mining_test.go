package ledgerservice

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ed25519"

	"github.com/cachecashproject/go-cachecash/keypair"
	"github.com/cachecashproject/go-cachecash/ledger"
	ledger_models "github.com/cachecashproject/go-cachecash/ledger/models"
	"github.com/cachecashproject/go-cachecash/ledger/txscript"
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

type txo struct {
	txid   ledger_models.TXID
	idx    uint8
	amount uint32
	script []byte
}

type output struct {
	target ed25519.PublicKey
	amount uint32
}

func TestMineBlockWithEmptyMempool(t *testing.T) {
	ctx := context.Background()

	kp, err := keypair.Generate()
	assert.Nil(t, err)

	lm, block, err := SetupGenesis(kp)
	assert.Nil(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, 1, len(block.Transactions.Transactions))

	block, err = lm.GenerateBlock(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, block)
}

func block2txos(block *ledger.Block) ([]txo, error) {
	txos := []txo{}

	for _, tx := range block.Transactions.Transactions {
		outputs, err := tx2txos(tx)
		if err != nil {
			return nil, err
		}
		txos = append(txos, outputs...)
	}

	return txos, nil
}

func tx2txos(tx *ledger.Transaction) ([]txo, error) {
	txos := []txo{}

	txid, err := tx.TXID()
	if err != nil {
		return nil, err
	}

	for idx, out := range tx.Outputs() {
		txos = append(txos, txo{
			txid:   txid,
			idx:    uint8(idx),
			amount: out.Value,
			script: out.ScriptPubKey,
		})
	}

	return txos, nil
}

func makeInputs(kp *keypair.KeyPair, txos []txo) ([]ledger.TransactionInput, []ledger.TransactionOutput, error) {
	inputs := []ledger.TransactionInput{}
	prevOutputs := []ledger.TransactionOutput{}

	for _, txo := range txos {
		inputScriptBytes, err := txscript.MakeOutputScript(kp.PublicKey)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to create output script")
		}

		inputs = append(inputs, ledger.TransactionInput{
			Outpoint: ledger.Outpoint{
				PreviousTx: txo.txid,
				Index:      uint8(txo.idx),
			},
			ScriptSig:  inputScriptBytes,
			SequenceNo: 0xFFFFFFFF,
		})
		prevOutputs = append(prevOutputs, ledger.TransactionOutput{
			Value:        txo.amount,
			ScriptPubKey: txo.script,
		})
	}

	return inputs, prevOutputs, nil
}

func makeOutputs(outs []output) ([]ledger.TransactionOutput, error) {
	outputs := make([]ledger.TransactionOutput, len(outs))

	for i, out := range outs {
		outputScriptBytes, err := txscript.MakeInputScript(out.target)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create input script")
		}

		txo := ledger.TransactionOutput{
			Value:        out.amount,
			ScriptPubKey: outputScriptBytes,
		}
		outputs[i] = txo
	}

	return outputs, nil
}

func makeTx(kp *keypair.KeyPair, txos []txo, out []output) (*ledger.Transaction, error) {
	inputs, prevOutputs, err := makeInputs(kp, txos)
	if err != nil {
		return nil, err
	}
	outputs, err := makeOutputs(out)
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
	assert.Equal(t, 1, len(block.Transactions.Transactions))
	r1 := block.Header.Random

	txos, err := block2txos(block)
	assert.Nil(t, err)

	tx, err := makeTx(kp, txos, []output{
		{
			target: kp.PublicKey,
			amount: 420000000,
		},
	})
	assert.Nil(t, err)

	err = lm.QueueTX(ctx, tx)
	assert.Nil(t, err)

	block, err = lm.GenerateBlock(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, 1, len(block.Transactions.Transactions))
	assert.NotEqual(t, 0, block.Header.Random)
	assert.NotEqual(t, r1, block.Header.Random)
}

func TestMineMultipleOutputs(t *testing.T) {
	ctx := context.Background()

	kp, err := keypair.Generate()
	assert.Nil(t, err)

	lm, block, err := SetupGenesis(kp)
	assert.Nil(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, 1, len(block.Transactions.Transactions))

	txos, err := block2txos(block)
	assert.Nil(t, err)

	tx, err := makeTx(kp, txos, []output{
		{
			target: kp.PublicKey,
			amount: 105000000,
		},
		{
			target: kp.PublicKey,
			amount: 105000000,
		},
		{
			target: kp.PublicKey,
			amount: 105000000,
		},
		{
			target: kp.PublicKey,
			amount: 105000000,
		},
	})
	assert.Nil(t, err)

	err = lm.QueueTX(ctx, tx)
	assert.Nil(t, err)

	block, err = lm.GenerateBlock(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, 1, len(block.Transactions.Transactions))
}

func TestMineMultipleInputs(t *testing.T) {
	ctx := context.Background()

	kp, err := keypair.Generate()
	assert.Nil(t, err)

	lm, block, err := SetupGenesis(kp)
	assert.Nil(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, 1, len(block.Transactions.Transactions))

	txos, err := block2txos(block)
	assert.Nil(t, err)

	tx, err := makeTx(kp, txos, []output{
		{
			target: kp.PublicKey,
			amount: 105000000,
		},
		{
			target: kp.PublicKey,
			amount: 105000000,
		},
		{
			target: kp.PublicKey,
			amount: 105000000,
		},
		{
			target: kp.PublicKey,
			amount: 105000000,
		},
	})
	assert.Nil(t, err)

	err = lm.QueueTX(ctx, tx)
	assert.Nil(t, err)

	block, err = lm.GenerateBlock(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, 1, len(block.Transactions.Transactions))

	txos, err = block2txos(block)
	assert.Nil(t, err)

	tx, err = makeTx(kp, txos, []output{
		{
			target: kp.PublicKey,
			amount: 420000000,
		},
	})
	assert.Nil(t, err)

	err = lm.QueueTX(ctx, tx)
	assert.Nil(t, err)

	block, err = lm.GenerateBlock(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, 1, len(block.Transactions.Transactions))
}

func TestMineOutputTwice(t *testing.T) {
	ctx := context.Background()

	kp, err := keypair.Generate()
	assert.Nil(t, err)

	lm, block, err := SetupGenesis(kp)
	assert.Nil(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, 1, len(block.Transactions.Transactions))

	txos, err := block2txos(block)
	assert.Nil(t, err)

	tx, err := makeTx(kp, txos, []output{
		{
			target: kp.PublicKey,
			amount: 420000000,
		},
	})
	assert.Nil(t, err)

	err = lm.QueueTX(ctx, tx)
	assert.Nil(t, err)

	block, err = lm.GenerateBlock(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, 1, len(block.Transactions.Transactions))

	// try to spend the same output twice
	err = lm.QueueTX(ctx, tx)
	assert.Nil(t, err)

	// The transaction won't be selected for mining
	block, err = lm.GenerateBlock(ctx)
	assert.Nil(t, err)
	assert.Len(t, block.Transactions.Transactions, 0)
}

func TestMineConflictingTXs(t *testing.T) {
	ctx := context.Background()

	kp, err := keypair.Generate()
	assert.Nil(t, err)

	lm, block, err := SetupGenesis(kp)
	assert.Nil(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, 1, len(block.Transactions.Transactions))

	txos, err := block2txos(block)
	assert.Nil(t, err)

	tx, err := makeTx(kp, txos, []output{
		{
			target: kp.PublicKey,
			amount: 420000000,
		},
	})
	assert.Nil(t, err)

	// only one of them is going to get mined
	err = lm.QueueTX(ctx, tx)
	assert.Nil(t, err)
	err = lm.QueueTX(ctx, tx)
	assert.Nil(t, err)

	block, err = lm.GenerateBlock(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, 1, len(block.Transactions.Transactions))

	assert.Equal(t, 1, len(block.Transactions.Transactions))
}

func TestMineDependingTXs(t *testing.T) {
	ctx := context.Background()

	kp, err := keypair.Generate()
	assert.Nil(t, err)

	lm, block, err := SetupGenesis(kp)
	assert.Nil(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, 1, len(block.Transactions.Transactions))

	txos, err := block2txos(block)
	assert.Nil(t, err)

	tx1, err := makeTx(kp, txos, []output{
		{
			target: kp.PublicKey,
			amount: 420000000,
		},
	})
	assert.Nil(t, err)

	txos, err = tx2txos(tx1)
	assert.Nil(t, err)

	tx2, err := makeTx(kp, txos, []output{
		{
			target: kp.PublicKey,
			amount: 420000000,
		},
	})
	assert.Nil(t, err)

	err = lm.QueueTX(ctx, tx1)
	assert.Nil(t, err)
	err = lm.QueueTX(ctx, tx2)
	assert.Nil(t, err)

	block, err = lm.GenerateBlock(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, block)
	assert.Len(t, block.Transactions.Transactions, 2)
}

func TestMineMultipleWallets(t *testing.T) {
	ctx := context.Background()

	kp, err := keypair.Generate()
	assert.Nil(t, err)

	lm, block, err := SetupGenesis(kp)
	assert.Nil(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, 1, len(block.Transactions.Transactions))

	next, err := keypair.Generate()
	assert.Nil(t, err)

	for i := 0; i < 10; i++ {
		txos, err := block2txos(block)
		assert.Nil(t, err)

		tx, err := makeTx(kp, txos, []output{
			{
				target: next.PublicKey,
				amount: 420000000,
			},
		})
		assert.Nil(t, err)

		err = lm.QueueTX(ctx, tx)
		assert.Nil(t, err)

		block, err = lm.GenerateBlock(ctx)
		assert.Nil(t, err)
		assert.NotNil(t, block)
		assert.Equal(t, 1, len(block.Transactions.Transactions))

		kp = next
		next, err = keypair.Generate()
		assert.Nil(t, err)
	}
}

func TestMineWrongWallet(t *testing.T) {
	ctx := context.Background()

	kp1, err := keypair.Generate()
	assert.Nil(t, err)

	kp2, err := keypair.Generate()
	assert.Nil(t, err)

	lm, block, err := SetupGenesis(kp1)
	assert.Nil(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, 1, len(block.Transactions.Transactions))

	txos, err := block2txos(block)
	assert.Nil(t, err)

	tx, err := makeTx(kp2, txos, []output{
		{
			target: kp1.PublicKey,
			amount: 420000000,
		},
	})
	assert.Nil(t, err)

	err = lm.QueueTX(ctx, tx)
	assert.Nil(t, err)

	// no transaction is going to be mined
	block, err = lm.GenerateBlock(ctx)
	assert.Nil(t, err)
	assert.Len(t, block.Transactions.Transactions, 0)
}

func TestMultipleGenesisBlocks(t *testing.T) {
	ctx := context.Background()

	kp, err := keypair.Generate()
	assert.Nil(t, err)

	lm, block, err := SetupGenesis(kp)
	assert.Nil(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, 1, len(block.Transactions.Transactions))

	outputs, err := makeOutputs([]output{
		{
			target: kp.PublicKey,
			amount: 420000000,
		},
	})
	assert.Nil(t, err)

	tx := &ledger.Transaction{
		Version: 1,
		Flags:   0,
		Body: &ledger.GenesisTransaction{
			Outputs: outputs,
		},
	}

	err = lm.QueueTX(ctx, tx)
	assert.Nil(t, err)

	// transaction is going to get rejected
	block, err = lm.GenerateBlock(ctx)
	assert.Nil(t, err)
	assert.Len(t, block.Transactions.Transactions, 0)
}

func TestMineInflation(t *testing.T) {
	ctx := context.Background()

	kp, err := keypair.Generate()
	assert.Nil(t, err)

	lm, block, err := SetupGenesis(kp)
	assert.Nil(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, 1, len(block.Transactions.Transactions))

	txos, err := block2txos(block)
	assert.Nil(t, err)

	tx, err := makeTx(kp, txos, []output{
		{
			target: kp.PublicKey,
			amount: 210000000,
		},
		{
			target: kp.PublicKey,
			amount: 210000000,
		},
		{
			target: kp.PublicKey,
			amount: 210000000,
		},
	})
	assert.Nil(t, err)

	err = lm.QueueTX(ctx, tx)
	assert.Nil(t, err)

	// no transaction is going to be mined
	block, err = lm.GenerateBlock(ctx)
	assert.Nil(t, err)
	assert.Len(t, block.Transactions.Transactions, 0)
}

// fees are currently not required/supported
func TestMineDeflation(t *testing.T) {
	ctx := context.Background()

	kp, err := keypair.Generate()
	assert.Nil(t, err)

	lm, block, err := SetupGenesis(kp)
	assert.Nil(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, 1, len(block.Transactions.Transactions))

	txos, err := block2txos(block)
	assert.Nil(t, err)

	tx, err := makeTx(kp, txos, []output{
		{
			target: kp.PublicKey,
			amount: 210000000,
		},
	})
	assert.Nil(t, err)

	err = lm.QueueTX(ctx, tx)
	assert.Nil(t, err)

	// no transaction is going to be mined
	block, err = lm.GenerateBlock(ctx)
	assert.Nil(t, err)
	assert.Len(t, block.Transactions.Transactions, 0)
}
