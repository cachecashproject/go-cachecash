package ledger

import (
	"testing"

	"github.com/cachecashproject/go-cachecash/keypair"
	"github.com/cachecashproject/go-cachecash/ledger/txscript"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ed25519"
)

func makeInput(pubkey ed25519.PublicKey, txid TXID) (*TransactionInput, error) {
	idx := uint8(0)

	inputScriptBytes, err := txscript.MakeOutputScript(pubkey)
	if err != nil {
		return nil, err
	}

	input := &TransactionInput{
		Outpoint: Outpoint{
			PreviousTx: txid,
			Index:      idx,
		},
		ScriptSig:  inputScriptBytes,
		SequenceNo: 0xFFFFFFFF,
	}

	return input, nil
}

func makeOutput(pubkey ed25519.PublicKey, amount uint32) (*TransactionOutput, error) {
	outputScriptBytes, err := txscript.MakeInputScript(pubkey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create input script")
	}

	txo := &TransactionOutput{
		Value:        amount,
		ScriptPubKey: outputScriptBytes,
	}

	return txo, nil
}

func makePrevTx(pubkey ed25519.PublicKey, amount uint32) (*Transaction, error) {
	output, err := makeOutput(pubkey, amount)
	if err != nil {
		return nil, err
	}

	return &Transaction{
		Version: 1,
		Flags:   0,
		Body: &GenesisTransaction{
			Outputs: []TransactionOutput{*output},
		},
	}, nil
}

func makeTx(prev *Transaction, kp *keypair.KeyPair, target ed25519.PublicKey, amount uint32) (*Transaction, []TransactionOutput, error) {
	txid, err := prev.TXID()
	if err != nil {
		return nil, nil, err
	}
	input, err := makeInput(kp.PublicKey, txid)
	if err != nil {
		return nil, nil, err
	}

	output, err := makeOutput(target, amount)
	if err != nil {
		return nil, nil, err
	}

	tx := &Transaction{
		Version: 1,
		Flags:   0,
		Body: &TransferTransaction{
			Inputs:   []TransactionInput{*input},
			Outputs:  []TransactionOutput{*output},
			LockTime: 0,
		},
	}

	prevOuts := prev.Outputs()
	err = tx.GenerateWitnesses(kp, prevOuts)
	if err != nil {
		return nil, nil, err
	}

	return tx, prevOuts, nil
}

func TestValidTransaction(t *testing.T) {
	amount := uint32(1234)

	kp1, err := keypair.Generate()
	assert.Nil(t, err)

	kp2, err := keypair.Generate()
	assert.Nil(t, err)

	prev, err := makePrevTx(kp1.PublicKey, amount)
	assert.Nil(t, err)

	tx, prevOuts, err := makeTx(prev, kp1, kp2.PublicKey, amount)
	assert.Nil(t, err)

	// Check that all script pairs execute correctly.
	witnesses := tx.Witnesses()
	for i, ti := range tx.Inputs() {
		inScr, err := txscript.ParseScript(ti.ScriptSig)
		assert.Nil(t, err)

		outScr, err := txscript.ParseScript(prevOuts[i].ScriptPubKey)
		assert.Nil(t, err)

		err = txscript.ExecuteVerify(inScr, outScr, witnesses[i].Data, tx, i, int64(prevOuts[i].Value))
		assert.Nil(t, err)
	}
}

func TestWrongWitnessSig(t *testing.T) {
	amount := uint32(1234)

	kp1, err := keypair.Generate()
	assert.Nil(t, err)

	kp2, err := keypair.Generate()
	assert.Nil(t, err)

	prev, err := makePrevTx(kp1.PublicKey, amount)
	assert.Nil(t, err)

	txid, err := prev.TXID()
	assert.Nil(t, err)
	input, err := makeInput(kp1.PublicKey, txid)
	assert.Nil(t, err)

	output, err := makeOutput(kp1.PublicKey, amount)
	assert.Nil(t, err)

	tx := &Transaction{
		Version: 1,
		Flags:   0,
		Body: &TransferTransaction{
			Inputs:   []TransactionInput{*input},
			Outputs:  []TransactionOutput{*output},
			LockTime: 0,
		},
	}

	prevOuts := prev.Outputs()
	err = tx.GenerateWitnesses(kp2, prevOuts)
	assert.Nil(t, err)

	// Check that all script pairs execute correctly.
	witnesses := tx.Witnesses()
	for i, ti := range tx.Inputs() {
		inScr, err := txscript.ParseScript(ti.ScriptSig)
		assert.Nil(t, err)

		outScr, err := txscript.ParseScript(prevOuts[i].ScriptPubKey)
		assert.Nil(t, err)

		err = txscript.ExecuteVerify(inScr, outScr, witnesses[i].Data, tx, i, int64(prevOuts[i].Value))
		assert.NotNil(t, err)
	}
}

func TestInvalidWrongInput(t *testing.T) {
	amount := uint32(1234)

	kp1, err := keypair.Generate()
	assert.Nil(t, err)

	kp2, err := keypair.Generate()
	assert.Nil(t, err)

	kp3, err := keypair.Generate()
	assert.Nil(t, err)

	prev, err := makePrevTx(kp1.PublicKey, amount)
	assert.Nil(t, err)

	txid, err := prev.TXID()
	assert.Nil(t, err)
	input, err := makeInput(kp3.PublicKey, txid)
	assert.Nil(t, err)

	output, err := makeOutput(kp2.PublicKey, amount)
	assert.Nil(t, err)

	tx := &Transaction{
		Version: 1,
		Flags:   0,
		Body: &TransferTransaction{
			Inputs:   []TransactionInput{*input},
			Outputs:  []TransactionOutput{*output},
			LockTime: 0,
		},
	}

	prevOuts := prev.Outputs()
	err = tx.GenerateWitnesses(kp1, prevOuts)
	assert.Nil(t, err)

	// Check that all script pairs execute correctly.
	witnesses := tx.Witnesses()
	for i, ti := range tx.Inputs() {
		inScr, err := txscript.ParseScript(ti.ScriptSig)
		assert.Nil(t, err)

		outScr, err := txscript.ParseScript(prevOuts[i].ScriptPubKey)
		assert.Nil(t, err)

		err = txscript.ExecuteVerify(inScr, outScr, witnesses[i].Data, tx, i, int64(prevOuts[i].Value))
		assert.NotNil(t, err)
	}
}
