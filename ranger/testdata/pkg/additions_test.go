// +build rangertest

package pkg

import (
	"fmt"
	"testing"

	"github.com/cachecashproject/go-cachecash/ranger"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type typeWithError struct {
	typ interface{}
	err error
}

func Test01Compiles(t *testing.T) {
	assert.Nil(t, nil)
}

func TestFuzzInputsGreen(t *testing.T) {
	gcli := &GlobalConfigListInsertion{}
	n, err := gcli.UnmarshalFrom([]byte("\xd2\n\n0000000000")[1:])
	assert.Nil(t, err)
	assert.Equal(t, n, gcli.Size())
	data, err := gcli.Marshal()
	assert.Nil(t, err)
	assert.Equal(t, len(data), gcli.Size())
	assert.Equal(t, n, len(data))

	gct := &GlobalConfigTransaction{}
	n, err = gct.UnmarshalFrom([]byte("\xef\x00\x00\x00\x02\x02\x01\x01\x00")[1:])
	assert.Nil(t, err)
	assert.Equal(t, n, gct.Size())
	data, err = gct.Marshal()
	assert.Nil(t, err)
	assert.Equal(t, len(data), gct.Size())
	assert.Equal(t, n, len(data))
	assert.Equal(t, []byte("\xef\x00\x00\x00\x02\x02\x01\x01\x00")[1:], data)
}

// TestFuzzInputs tests fuzz inputs that have shown errors in the past.
// take the first byte to select the objects to marshal to/from in the fuzz
// test. this is why you see [1:] everywhere the inputs are used.
func TestFuzzInputsBasic(t *testing.T) {
	ios := map[string]typeWithError{
		"0\xef\xef\xef\xbf\xbd\x01\x00": {
			&TransactionOutput{},
			ranger.ErrTooLarge,
		},
		"\\\x97\x00\x80\x00\x00": {
			&GenesisTransaction{},
			ranger.ErrTooMany,
		},
		"\xb40\x82": {
			&Transactions{},
			ranger.ErrTooMany,
		},
		"\xed\x00\x000\xff\xff\xff": {
			&GlobalConfigListUpdate{},
			ranger.ErrTooMany,
		},
		"0��0\x00": {
			&TransactionOutput{},
			ranger.ErrTooLarge,
		},
		"<\b\x05\x00\x10{H00\x000": {
			&TransactionInput{},
			ranger.ErrLengthMismatch,
		},
		"\xbd0\x03\x00": {
			&Transaction{},
			ranger.ErrShortRead,
		},
		"\xdb0\a0\x01\x02\x80\x0000": {
			&Transactions{},
			ranger.ErrTooMany,
		},
	}

	for data, twe := range ios {
		_, err := twe.typ.(ranger.Marshaler).UnmarshalFrom([]byte(data)[1:])
		assert.Equal(t, errors.Cause(err), twe.err, fmt.Sprintf("%q", data))
	}
}

func TestFuzzInputsRoundTrip(t *testing.T) {
	tt := &Transaction{}
	_, err := tt.UnmarshalFrom([]byte("0\x03\x0000"))
	assert.Nil(t, err)
	data, err := tt.Marshal()
	assert.Nil(t, err)
	_, err = tt.UnmarshalFrom(data)
	assert.Nil(t, err)

	gt := &GenesisTransaction{}
	_, err = gt.UnmarshalFrom([]byte("\x010\x00"))
	assert.Nil(t, err)
	_, err = gt.Marshal()
	assert.Nil(t, err)

	tt = &Transaction{}
	_, err = tt.UnmarshalFrom([]byte("0\x01\x0000"))
	assert.Nil(t, err)
	_, err = tt.Marshal()
	assert.Nil(t, err)

	tt = &Transaction{}
	_, err = tt.UnmarshalFrom([]byte("0\x020\x00\x00\x0200\x0000"))
	assert.Nil(t, err)
	data, err = tt.Marshal()
	assert.Nil(t, err)
	tt2 := &Transaction{}
	_, err = tt2.UnmarshalFrom(data)
	assert.Nil(t, err)
}
