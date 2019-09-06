package kv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func makeMemory(member string, init MemoryTable) *Client {
	return NewClient(member, NewMemoryDriver(init))
}

func TestKVCreateDelete(t *testing.T) {
	c := makeMemory("basic", nil)

	nonce, err := c.CreateUint64("uint", 0)
	assert.Nil(t, err)

	_, err = c.CreateUint64("uint", 0)
	assert.Equal(t, err, ErrAlreadySet)

	assert.Nil(t, c.Delete("uint", nonce))
	assert.Equal(t, c.Delete("uint", nonce), ErrNotEqual)
	assert.Equal(t, c.Delete("uint", nil), ErrUnsetValue)

	_, err = c.CreateUint64("uint", 0)
	assert.Nil(t, err)
	assert.Nil(t, c.Delete("uint", nil))
}

func TestKVBasic(t *testing.T) {
	c := makeMemory("basic", nil)
	_, _, err := c.GetUint64("uint")
	assert.Equal(t, err, ErrUnsetValue)

	nonce, err := c.SetUint64("uint", 1)
	assert.Nil(t, err)
	uintOut, _, err := c.GetUint64("uint")
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), uintOut)
	_, err = c.CASUint64("uint", nonce, 0, 2)
	assert.Equal(t, err, ErrNotEqual)
	_, err = c.CASUint64("uint", []byte{1, 2, 3, 4}, 1, 2)
	assert.Equal(t, err, ErrNotEqual)
	newnonce, err := c.CASUint64("uint", nonce, 1, 2)
	assert.Nil(t, err)
	assert.NotEqual(t, newnonce, nonce)
	uintOut, _, err = c.GetUint64("uint")
	assert.Nil(t, err)
	assert.Equal(t, uint64(2), uintOut)

	_, _, err = c.GetInt64("int")
	assert.Equal(t, err, ErrUnsetValue)
	nonce, err = c.SetInt64("int", -1)
	assert.Nil(t, err)
	intOut, _, err := c.GetInt64("int")
	assert.Nil(t, err)
	assert.Equal(t, int64(-1), intOut)
	_, err = c.CASInt64("int", nonce, 0, -2)
	assert.Equal(t, err, ErrNotEqual)
	_, err = c.CASInt64("int", []byte{1, 2, 3, 4}, -1, -2)
	assert.Equal(t, err, ErrNotEqual)
	_, err = c.CASInt64("int", nonce, -1, -2)
	assert.Nil(t, err)
	intOut, _, err = c.GetInt64("int")
	assert.Nil(t, err)
	assert.Equal(t, int64(-2), intOut)

	_, _, err = c.GetFloat64("float")
	assert.Equal(t, err, ErrUnsetValue)
	nonce, err = c.SetFloat64("float", -1.2)
	assert.Nil(t, err)
	floatOut, _, err := c.GetFloat64("float")
	assert.Nil(t, err)
	assert.Equal(t, float64(-1.2), floatOut)
	_, err = c.CASFloat64("float", nonce, -1.1, -2.4)
	assert.Equal(t, err, ErrNotEqual)
	_, err = c.CASFloat64("float", []byte{1, 2, 3, 4}, -1.2, -2.4)
	assert.Equal(t, err, ErrNotEqual)
	_, err = c.CASFloat64("float", nonce, -1.2, -2.4)
	assert.Nil(t, err)
	floatOut, _, err = c.GetFloat64("float")
	assert.Nil(t, err)
	assert.Equal(t, float64(-2.4), floatOut)

	_, _, err = c.GetString("str")
	assert.Equal(t, err, ErrUnsetValue)
	nonce, err = c.SetString("str", "hello")
	assert.Nil(t, err)
	strOut, _, err := c.GetString("str")
	assert.Nil(t, err)
	assert.Equal(t, "hello", strOut)
	_, err = c.CASString("str", nonce, "nope", "world")
	assert.Equal(t, err, ErrNotEqual)
	_, err = c.CASString("str", []byte{1, 2, 3, 4}, "hello", "world")
	assert.Equal(t, err, ErrNotEqual)
	_, err = c.CASString("str", nonce, "hello", "world")
	assert.Nil(t, err)
	strOut, _, err = c.GetString("str")
	assert.Nil(t, err)
	assert.Equal(t, "world", strOut)
}

func TestKVMember(t *testing.T) {
	c1 := makeMemory("member1", nil)
	c2 := makeMemory("member2", nil)

	_, err := c1.SetUint64("one", 1)
	assert.Nil(t, err)

	_, _, err = c2.GetUint64("one")
	assert.Equal(t, err, ErrUnsetValue)
	_, err = c2.SetUint64("one", 2)
	assert.Nil(t, err)

	out, _, err := c1.GetUint64("one")
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), out)
}
