package ledger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockIDZero(t *testing.T) {
	var zero_bytes [BlockIDSize]byte
	id := BlockID(zero_bytes)
	assert.True(t, id.Zero())
	id[0] = 1
	assert.False(t, id.Zero())
}

func TestBlockIDEqual(t *testing.T) {
	var left_bytes [BlockIDSize]byte
	left := BlockID(left_bytes)
	left[0] = 1
	var mid_bytes [BlockIDSize]byte
	mid := BlockID(mid_bytes)
	var right_bytes [BlockIDSize]byte
	right := BlockID(right_bytes)
	right[31] = 1
	assert.Equal(t, left, left)
	assert.Equal(t, mid, mid)
	assert.Equal(t, right, right)
	assert.NotEqual(t, left, mid)
	assert.NotEqual(t, mid, right)
}

func TestBlockIDString(t *testing.T) {
	var zero_bytes [BlockIDSize]byte
	id := BlockID(zero_bytes)
	assert.Equal(t, "0000000000000000000000000000000000000000000000000000000000000000", id.String())
	id[0] = 1
	assert.Equal(t, "0100000000000000000000000000000000000000000000000000000000000000", id.String())
}

func TestTXIDEqual(t *testing.T) {
	var left_bytes [TransactionIDSize]byte
	left := TXID(left_bytes)
	left[0] = 1
	var mid_bytes [TransactionIDSize]byte
	mid := TXID(mid_bytes)
	var right_bytes [TransactionIDSize]byte
	right := TXID(right_bytes)
	right[31] = 1
	assert.Equal(t, left, left)
	assert.Equal(t, mid, mid)
	assert.Equal(t, right, right)
	assert.NotEqual(t, left, mid)
	assert.NotEqual(t, mid, right)
}

func TestTXIDString(t *testing.T) {
	var zero_bytes [TransactionIDSize]byte
	id := TXID(zero_bytes)
	assert.Equal(t, "0000000000000000000000000000000000000000000000000000000000000000", id.String())
	id[0] = 1
	assert.Equal(t, "0100000000000000000000000000000000000000000000000000000000000000", id.String())
}
