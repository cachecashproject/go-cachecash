package ledger

import (
	"testing"

	"github.com/cachecashproject/go-cachecash/keypair"
	"github.com/stretchr/testify/assert"
)

func TestSingleTx(t *testing.T) {
	s := NewSpendingState()

	kp, err := keypair.Generate()
	assert.Nil(t, err)

	prev, err := makePrevTx(kp.PublicKey, uint32(1234))
	assert.Nil(t, err)

	tx, _, err := makeTx(prev, kp, kp.PublicKey, uint32(1234))
	assert.Nil(t, err)

	err = s.AddTx(tx)
	assert.Nil(t, err)

	inpoints := tx.Inpoints()
	inpointKeys := make([]OutpointKey, len(inpoints))
	for i, inpoint := range inpoints {
		inpointKeys[i] = inpoint.Key()
	}

	for _, outpoint := range tx.Outpoints() {
		assert.NotNil(t, s.IsNewUnspent(outpoint.Key()))
	}
	assert.Nil(t, s.IsNewUnspent(OutpointKey{}))

	spent := s.SpentUTXOs()
	assert.Equal(t, inpointKeys, spent)
}

func TestTwoTx(t *testing.T) {
	s := NewSpendingState()

	kp, err := keypair.Generate()
	assert.Nil(t, err)

	prev, err := makePrevTx(kp.PublicKey, uint32(1234))
	assert.Nil(t, err)

	tx, _, err := makeTx(prev, kp, kp.PublicKey, uint32(1234))
	assert.Nil(t, err)

	err = s.AddTx(prev)
	assert.Nil(t, err)

	err = s.AddTx(tx)
	assert.Nil(t, err)
}

func TestAddTxTwice(t *testing.T) {
	s := NewSpendingState()

	kp, err := keypair.Generate()
	assert.Nil(t, err)

	prev, err := makePrevTx(kp.PublicKey, uint32(1234))
	assert.Nil(t, err)

	tx, _, err := makeTx(prev, kp, kp.PublicKey, uint32(1234))
	assert.Nil(t, err)

	err = s.AddTx(tx)
	assert.Nil(t, err)

	err = s.AddTx(tx)
	assert.NotNil(t, err)
}

func TestBlockLimit(t *testing.T) {
	s := NewSpendingState()

	kp, err := keypair.Generate()
	assert.Nil(t, err)

	prev, err := makePrevTx(kp.PublicKey, uint32(1234))
	assert.Nil(t, err)

	// insert regular transaction
	tx, _, err := makeTx(prev, kp, kp.PublicKey, uint32(1234))
	assert.Nil(t, err)
	err = s.AddTx(tx)
	assert.Nil(t, err)

	// fill block with transactions
	for i := 0; i < 10000; i++ {
		tx, _, err = makeTx(tx, kp, kp.PublicKey, uint32(1234))
		assert.Nil(t, err)
		_ = s.AddTx(tx)
	}

	// this fails, block is full
	tx, _, err = makeTx(tx, kp, kp.PublicKey, uint32(1234))
	assert.Nil(t, err)
	err = s.AddTx(tx)
	assert.NotNil(t, err)

	// ensure we have multiple accepted transactions
	assert.True(t, len(s.AcceptedTransactions()) > 10)
}
