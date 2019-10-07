package pkg

import (
	"testing"

	"github.com/cachecashproject/go-cachecash/ranger"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func Test01Compiles(t *testing.T) {
	assert.Nil(t, nil)
}

func TestRandomInputs(t *testing.T) {
}

// all fuzz inputs take the first byte to select the objects to marshal to/from
// in the fuzz test. this is why you see [1:] everywhere the inputs are used.
func TestFuzzInputs(t *testing.T) {
	to := &TransactionOutput{}
	n, err := to.UnmarshalFrom([]byte("0�\x01\x00")[1:])
	assert.Equal(t, errors.Cause(err), ranger.ErrTooLarge)

	gt := &GenesisTransaction{}
	_, err = gt.UnmarshalFrom([]byte("\\\x97\x00\x80\x00\x00")[1:])
	assert.Equal(t, errors.Cause(err), ranger.ErrTooMany)

	ts := &Transactions{}
	_, err = ts.UnmarshalFrom([]byte("\xb40\x82")[1:])
	assert.Equal(t, errors.Cause(err), ranger.ErrTooMany)

	gclu := &GlobalConfigListUpdate{}
	_, err = gclu.UnmarshalFrom([]byte("\xed\x00\x000\xff\xff\xff")[1:])
	assert.Equal(t, errors.Cause(err), ranger.ErrTooMany)

	tt := &Transaction{}
	_, err = tt.UnmarshalFrom([]byte("U0\x00\xd20")[1:])
	assert.Equal(t, errors.Cause(err), ranger.ErrShortRead)

	t2 := &TransferTransaction{}
	_, err = t2.UnmarshalFrom([]byte("@\x00\x00\x02\x00\x00\x00")[1:])
	assert.Equal(t, errors.Cause(err), ranger.ErrShortRead)

	gcli := &GlobalConfigListInsertion{}
	n, err = gcli.UnmarshalFrom([]byte("\xd2\n\n0000000000")[1:])
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

	to = &TransactionOutput{}
	n, err = to.UnmarshalFrom([]byte("0��0\x00")[1:])
	assert.Equal(t, errors.Cause(err), ranger.ErrTooLarge)

	ti := &TransactionInput{}
	n, err = ti.UnmarshalFrom([]byte("<\b\x05\x00\x10{H00\x000")[1:])
	assert.Equal(t, errors.Cause(err), ranger.ErrLengthMismatch)

	tt = &Transaction{}
	_, err = tt.UnmarshalFrom([]byte("\xbd0\x03\x00")[1:])
	assert.Equal(t, errors.Cause(err), ranger.ErrShortRead)

	ts = &Transactions{}
	_, err = ts.UnmarshalFrom([]byte("\xdb0\a0\x01\x02\x80\x0000")[1:])
	assert.Equal(t, errors.Cause(err), ranger.ErrTooMany)

	tt = &Transaction{}
	_, err = tt.UnmarshalFrom([]byte("H0\x03\x0000")[1:])
	assert.Nil(t, err)
	data, err = tt.Marshal()
	assert.Nil(t, err)
	_, err = tt.UnmarshalFrom(data)
	assert.Nil(t, err)

	gt = &GenesisTransaction{}
	_, err = gt.UnmarshalFrom([]byte("\x01\x01\x020\x00")[1:])
	assert.Nil(t, err)
	_, err = gt.Marshal()
	assert.Nil(t, err)

	tt = &Transaction{}
	_, err = tt.UnmarshalFrom([]byte("\xbd0\x01\x01\x0000")[1:])
	assert.Nil(t, err)
	_, err = tt.Marshal()
	assert.Nil(t, err)

	tt = &Transaction{}
	_, err = tt.UnmarshalFrom([]byte("\a0\x02\a0\x00\x00\x0200\x0000")[1:])
	assert.Nil(t, err)
	data, err = tt.Marshal()
	assert.Nil(t, err)
	tt2 := &Transaction{}
	_, err = tt2.UnmarshalFrom(data)
	assert.Nil(t, err)
}
