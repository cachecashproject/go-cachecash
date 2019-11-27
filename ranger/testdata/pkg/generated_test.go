package pkg

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func genRandom(n int) []byte {
	data := make([]byte, n)
	n2, err := rand.Read(data)
	if err != nil {
		panic(errors.Wrap(err, "rand.Read"))
	}

	if n != n2 {
		panic(errors.Wrap(err, "short read in rand.Read"))
	}

	return data[:n]
}

func TestEscrowOpenTransactionMarshalUnmarshalZeroValue(t *testing.T) {
	obj := &EscrowOpenTransaction{}

	obj2 := &EscrowOpenTransaction{}

	data, err := obj.Marshal()
	assert.Nil(t, err, "marshal failed for EscrowOpenTransaction")
	assert.Equal(t, len(data), obj.Size(), "EscrowOpenTransaction size check on zero value")
	assert.Nil(t, obj2.Unmarshal(data), "EscrowOpenTransaction zero value unmarshal test")
	assert.Equal(t, obj, obj2, "EscrowOpenTransaction unmarshal equality test")
	obj2 = &EscrowOpenTransaction{}
	l, err := obj2.UnmarshalFrom(data)
	assert.Nil(t, err, "EscrowOpenTransaction unmarshalfrom failed")
	assert.Equal(t, obj, obj2, "EscrowOpenTransaction unmarshalfrom equality test")
	assert.Equal(t, len(data), l, "EscrowOpenTransaction data length check")
	assert.Equal(t, obj.Size(), l, "EscrowOpenTransaction data size check")
}

func TestEscrowOpenTransactionMarshalUnmarshalRandomData(t *testing.T) {
	seed := time.Now().Unix()
	fmt.Printf("Seed is %v\n", seed)
	rand.Seed(seed)

	for i := 0; i < 100; i++ {
		obj := &EscrowOpenTransaction{}

		obj2 := &EscrowOpenTransaction{}

		data, err := obj.Marshal()
		assert.Nil(t, err, "marshal failed for EscrowOpenTransaction")
		assert.Equal(t, len(data), obj.Size(), "EscrowOpenTransaction size check on random values")
		assert.Nil(t, obj2.Unmarshal(data), "EscrowOpenTransaction random values unmarshal test")
		assert.Equal(t, obj, obj2, "EscrowOpenTransaction unmarshal equality test")

		obj2 = &EscrowOpenTransaction{}

		l, err := obj2.UnmarshalFrom(data)
		assert.Nil(t, err, fmt.Sprintf("EscrowOpenTransaction unmarshalfrom failed: %q", hex.EncodeToString(data)))
		assert.Equal(t, obj, obj2, "EscrowOpenTransaction unmarshalfrom equality test")
		assert.Equal(t, len(data), l, "EscrowOpenTransaction data length check")
		assert.Equal(t, obj.Size(), l, "EscrowOpenTransaction data size check")

		assert.True(t, reflect.DeepEqual(obj, obj2))
	}
}

func TestGenesisTransactionMarshalUnmarshalZeroValue(t *testing.T) {
	obj := &GenesisTransaction{

		Outputs: make([]*TransactionOutput, 0)}

	obj2 := &GenesisTransaction{

		Outputs: make([]*TransactionOutput, 0)}

	data, err := obj.Marshal()
	assert.Nil(t, err, "marshal failed for GenesisTransaction")
	assert.Equal(t, len(data), obj.Size(), "GenesisTransaction size check on zero value")
	assert.Nil(t, obj2.Unmarshal(data), "GenesisTransaction zero value unmarshal test")
	assert.Equal(t, obj, obj2, "GenesisTransaction unmarshal equality test")
	obj2 = &GenesisTransaction{

		Outputs: make([]*TransactionOutput, 0)}
	l, err := obj2.UnmarshalFrom(data)
	assert.Nil(t, err, "GenesisTransaction unmarshalfrom failed")
	assert.Equal(t, obj, obj2, "GenesisTransaction unmarshalfrom equality test")
	assert.Equal(t, len(data), l, "GenesisTransaction data length check")
	assert.Equal(t, obj.Size(), l, "GenesisTransaction data size check")
}

func TestGenesisTransactionMarshalUnmarshalRandomData(t *testing.T) {
	seed := time.Now().Unix()
	fmt.Printf("Seed is %v\n", seed)
	rand.Seed(seed)

	for i := 0; i < 100; i++ {
		obj := &GenesisTransaction{}

		obj.Outputs = []*TransactionOutput{&TransactionOutput{
			Value:        uint32(rand.Uint64() & math.MaxUint32),
			ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
		}, &TransactionOutput{
			Value:        uint32(rand.Uint64() & math.MaxUint32),
			ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
		}, &TransactionOutput{
			Value:        uint32(rand.Uint64() & math.MaxUint32),
			ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
		}, &TransactionOutput{
			Value:        uint32(rand.Uint64() & math.MaxUint32),
			ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
		}, &TransactionOutput{
			Value:        uint32(rand.Uint64() & math.MaxUint32),
			ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
		}, &TransactionOutput{
			Value:        uint32(rand.Uint64() & math.MaxUint32),
			ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
		}, &TransactionOutput{
			Value:        uint32(rand.Uint64() & math.MaxUint32),
			ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
		}, &TransactionOutput{
			Value:        uint32(rand.Uint64() & math.MaxUint32),
			ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
		}, &TransactionOutput{
			Value:        uint32(rand.Uint64() & math.MaxUint32),
			ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
		}, &TransactionOutput{
			Value:        uint32(rand.Uint64() & math.MaxUint32),
			ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
		}}

		obj2 := &GenesisTransaction{}

		data, err := obj.Marshal()
		assert.Nil(t, err, "marshal failed for GenesisTransaction")
		assert.Equal(t, len(data), obj.Size(), "GenesisTransaction size check on random values")
		assert.Nil(t, obj2.Unmarshal(data), "GenesisTransaction random values unmarshal test")
		assert.Equal(t, obj, obj2, "GenesisTransaction unmarshal equality test")

		obj2 = &GenesisTransaction{}

		l, err := obj2.UnmarshalFrom(data)
		assert.Nil(t, err, fmt.Sprintf("GenesisTransaction unmarshalfrom failed: %q", hex.EncodeToString(data)))
		assert.Equal(t, obj, obj2, "GenesisTransaction unmarshalfrom equality test")
		assert.Equal(t, len(data), l, "GenesisTransaction data length check")
		assert.Equal(t, obj.Size(), l, "GenesisTransaction data size check")

		assert.True(t, reflect.DeepEqual(obj, obj2))
	}
}

func TestGlobalConfigListInsertionMarshalUnmarshalZeroValue(t *testing.T) {
	obj := &GlobalConfigListInsertion{

		Index: 0,
		Value: make([]byte, 0)}

	obj2 := &GlobalConfigListInsertion{

		Index: 0,
		Value: make([]byte, 0)}

	data, err := obj.Marshal()
	assert.Nil(t, err, "marshal failed for GlobalConfigListInsertion")
	assert.Equal(t, len(data), obj.Size(), "GlobalConfigListInsertion size check on zero value")
	assert.Nil(t, obj2.Unmarshal(data), "GlobalConfigListInsertion zero value unmarshal test")
	assert.Equal(t, obj, obj2, "GlobalConfigListInsertion unmarshal equality test")
	obj2 = &GlobalConfigListInsertion{

		Index: 0,
		Value: make([]byte, 0)}
	l, err := obj2.UnmarshalFrom(data)
	assert.Nil(t, err, "GlobalConfigListInsertion unmarshalfrom failed")
	assert.Equal(t, obj, obj2, "GlobalConfigListInsertion unmarshalfrom equality test")
	assert.Equal(t, len(data), l, "GlobalConfigListInsertion data length check")
	assert.Equal(t, obj.Size(), l, "GlobalConfigListInsertion data size check")
}

func TestGlobalConfigListInsertionMarshalUnmarshalRandomData(t *testing.T) {
	seed := time.Now().Unix()
	fmt.Printf("Seed is %v\n", seed)
	rand.Seed(seed)

	for i := 0; i < 100; i++ {
		obj := &GlobalConfigListInsertion{}

		obj.Index = uint64(rand.Uint64() & math.MaxUint64)

		obj.Value = []byte(genRandom(rand.Int() % 20))

		obj2 := &GlobalConfigListInsertion{}

		data, err := obj.Marshal()
		assert.Nil(t, err, "marshal failed for GlobalConfigListInsertion")
		assert.Equal(t, len(data), obj.Size(), "GlobalConfigListInsertion size check on random values")
		assert.Nil(t, obj2.Unmarshal(data), "GlobalConfigListInsertion random values unmarshal test")
		assert.Equal(t, obj, obj2, "GlobalConfigListInsertion unmarshal equality test")

		obj2 = &GlobalConfigListInsertion{}

		l, err := obj2.UnmarshalFrom(data)
		assert.Nil(t, err, fmt.Sprintf("GlobalConfigListInsertion unmarshalfrom failed: %q", hex.EncodeToString(data)))
		assert.Equal(t, obj, obj2, "GlobalConfigListInsertion unmarshalfrom equality test")
		assert.Equal(t, len(data), l, "GlobalConfigListInsertion data length check")
		assert.Equal(t, obj.Size(), l, "GlobalConfigListInsertion data size check")

		assert.True(t, reflect.DeepEqual(obj, obj2))
	}
}

func TestGlobalConfigListUpdateMarshalUnmarshalZeroValue(t *testing.T) {
	obj := &GlobalConfigListUpdate{

		Key:        "",
		Deletions:  make([]uint64, 0),
		Insertions: make([]*GlobalConfigListInsertion, 0)}

	obj2 := &GlobalConfigListUpdate{

		Key:        "",
		Deletions:  make([]uint64, 0),
		Insertions: make([]*GlobalConfigListInsertion, 0)}

	data, err := obj.Marshal()
	assert.Nil(t, err, "marshal failed for GlobalConfigListUpdate")
	assert.Equal(t, len(data), obj.Size(), "GlobalConfigListUpdate size check on zero value")
	assert.Nil(t, obj2.Unmarshal(data), "GlobalConfigListUpdate zero value unmarshal test")
	assert.Equal(t, obj, obj2, "GlobalConfigListUpdate unmarshal equality test")
	obj2 = &GlobalConfigListUpdate{

		Key:        "",
		Deletions:  make([]uint64, 0),
		Insertions: make([]*GlobalConfigListInsertion, 0)}
	l, err := obj2.UnmarshalFrom(data)
	assert.Nil(t, err, "GlobalConfigListUpdate unmarshalfrom failed")
	assert.Equal(t, obj, obj2, "GlobalConfigListUpdate unmarshalfrom equality test")
	assert.Equal(t, len(data), l, "GlobalConfigListUpdate data length check")
	assert.Equal(t, obj.Size(), l, "GlobalConfigListUpdate data size check")
}

func TestGlobalConfigListUpdateMarshalUnmarshalRandomData(t *testing.T) {
	seed := time.Now().Unix()
	fmt.Printf("Seed is %v\n", seed)
	rand.Seed(seed)

	for i := 0; i < 100; i++ {
		obj := &GlobalConfigListUpdate{}

		obj.Key = string(genRandom(rand.Int() % 20))

		obj.Deletions = []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)}

		obj.Insertions = []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
			Index: uint64(rand.Uint64() & math.MaxUint64),
			Value: []byte(genRandom(rand.Int() % 20)),
		}, &GlobalConfigListInsertion{
			Index: uint64(rand.Uint64() & math.MaxUint64),
			Value: []byte(genRandom(rand.Int() % 20)),
		}, &GlobalConfigListInsertion{
			Index: uint64(rand.Uint64() & math.MaxUint64),
			Value: []byte(genRandom(rand.Int() % 20)),
		}, &GlobalConfigListInsertion{
			Index: uint64(rand.Uint64() & math.MaxUint64),
			Value: []byte(genRandom(rand.Int() % 20)),
		}, &GlobalConfigListInsertion{
			Index: uint64(rand.Uint64() & math.MaxUint64),
			Value: []byte(genRandom(rand.Int() % 20)),
		}, &GlobalConfigListInsertion{
			Index: uint64(rand.Uint64() & math.MaxUint64),
			Value: []byte(genRandom(rand.Int() % 20)),
		}, &GlobalConfigListInsertion{
			Index: uint64(rand.Uint64() & math.MaxUint64),
			Value: []byte(genRandom(rand.Int() % 20)),
		}, &GlobalConfigListInsertion{
			Index: uint64(rand.Uint64() & math.MaxUint64),
			Value: []byte(genRandom(rand.Int() % 20)),
		}, &GlobalConfigListInsertion{
			Index: uint64(rand.Uint64() & math.MaxUint64),
			Value: []byte(genRandom(rand.Int() % 20)),
		}, &GlobalConfigListInsertion{
			Index: uint64(rand.Uint64() & math.MaxUint64),
			Value: []byte(genRandom(rand.Int() % 20)),
		}}

		obj2 := &GlobalConfigListUpdate{}

		data, err := obj.Marshal()
		assert.Nil(t, err, "marshal failed for GlobalConfigListUpdate")
		assert.Equal(t, len(data), obj.Size(), "GlobalConfigListUpdate size check on random values")
		assert.Nil(t, obj2.Unmarshal(data), "GlobalConfigListUpdate random values unmarshal test")
		assert.Equal(t, obj, obj2, "GlobalConfigListUpdate unmarshal equality test")

		obj2 = &GlobalConfigListUpdate{}

		l, err := obj2.UnmarshalFrom(data)
		assert.Nil(t, err, fmt.Sprintf("GlobalConfigListUpdate unmarshalfrom failed: %q", hex.EncodeToString(data)))
		assert.Equal(t, obj, obj2, "GlobalConfigListUpdate unmarshalfrom equality test")
		assert.Equal(t, len(data), l, "GlobalConfigListUpdate data length check")
		assert.Equal(t, obj.Size(), l, "GlobalConfigListUpdate data size check")

		assert.True(t, reflect.DeepEqual(obj, obj2))
	}
}

func TestGlobalConfigScalarUpdateMarshalUnmarshalZeroValue(t *testing.T) {
	obj := &GlobalConfigScalarUpdate{

		Key:   "",
		Value: make([]byte, 0)}

	obj2 := &GlobalConfigScalarUpdate{

		Key:   "",
		Value: make([]byte, 0)}

	data, err := obj.Marshal()
	assert.Nil(t, err, "marshal failed for GlobalConfigScalarUpdate")
	assert.Equal(t, len(data), obj.Size(), "GlobalConfigScalarUpdate size check on zero value")
	assert.Nil(t, obj2.Unmarshal(data), "GlobalConfigScalarUpdate zero value unmarshal test")
	assert.Equal(t, obj, obj2, "GlobalConfigScalarUpdate unmarshal equality test")
	obj2 = &GlobalConfigScalarUpdate{

		Key:   "",
		Value: make([]byte, 0)}
	l, err := obj2.UnmarshalFrom(data)
	assert.Nil(t, err, "GlobalConfigScalarUpdate unmarshalfrom failed")
	assert.Equal(t, obj, obj2, "GlobalConfigScalarUpdate unmarshalfrom equality test")
	assert.Equal(t, len(data), l, "GlobalConfigScalarUpdate data length check")
	assert.Equal(t, obj.Size(), l, "GlobalConfigScalarUpdate data size check")
}

func TestGlobalConfigScalarUpdateMarshalUnmarshalRandomData(t *testing.T) {
	seed := time.Now().Unix()
	fmt.Printf("Seed is %v\n", seed)
	rand.Seed(seed)

	for i := 0; i < 100; i++ {
		obj := &GlobalConfigScalarUpdate{}

		obj.Key = string(genRandom(rand.Int() % 20))

		obj.Value = []byte(genRandom(rand.Int() % 20))

		obj2 := &GlobalConfigScalarUpdate{}

		data, err := obj.Marshal()
		assert.Nil(t, err, "marshal failed for GlobalConfigScalarUpdate")
		assert.Equal(t, len(data), obj.Size(), "GlobalConfigScalarUpdate size check on random values")
		assert.Nil(t, obj2.Unmarshal(data), "GlobalConfigScalarUpdate random values unmarshal test")
		assert.Equal(t, obj, obj2, "GlobalConfigScalarUpdate unmarshal equality test")

		obj2 = &GlobalConfigScalarUpdate{}

		l, err := obj2.UnmarshalFrom(data)
		assert.Nil(t, err, fmt.Sprintf("GlobalConfigScalarUpdate unmarshalfrom failed: %q", hex.EncodeToString(data)))
		assert.Equal(t, obj, obj2, "GlobalConfigScalarUpdate unmarshalfrom equality test")
		assert.Equal(t, len(data), l, "GlobalConfigScalarUpdate data length check")
		assert.Equal(t, obj.Size(), l, "GlobalConfigScalarUpdate data size check")

		assert.True(t, reflect.DeepEqual(obj, obj2))
	}
}

func TestGlobalConfigTransactionMarshalUnmarshalZeroValue(t *testing.T) {
	obj := &GlobalConfigTransaction{

		ActivationBlockHeight: 0,
		ScalarUpdates:         make([]*GlobalConfigScalarUpdate, 0),
		ListUpdates:           make([]*GlobalConfigListUpdate, 0),
		SigPublicKey:          make([]byte, 0),
		Signature:             make([]byte, 0)}

	obj2 := &GlobalConfigTransaction{

		ActivationBlockHeight: 0,
		ScalarUpdates:         make([]*GlobalConfigScalarUpdate, 0),
		ListUpdates:           make([]*GlobalConfigListUpdate, 0),
		SigPublicKey:          make([]byte, 0),
		Signature:             make([]byte, 0)}

	data, err := obj.Marshal()
	assert.Nil(t, err, "marshal failed for GlobalConfigTransaction")
	assert.Equal(t, len(data), obj.Size(), "GlobalConfigTransaction size check on zero value")
	assert.Nil(t, obj2.Unmarshal(data), "GlobalConfigTransaction zero value unmarshal test")
	assert.Equal(t, obj, obj2, "GlobalConfigTransaction unmarshal equality test")
	obj2 = &GlobalConfigTransaction{

		ActivationBlockHeight: 0,
		ScalarUpdates:         make([]*GlobalConfigScalarUpdate, 0),
		ListUpdates:           make([]*GlobalConfigListUpdate, 0),
		SigPublicKey:          make([]byte, 0),
		Signature:             make([]byte, 0)}
	l, err := obj2.UnmarshalFrom(data)
	assert.Nil(t, err, "GlobalConfigTransaction unmarshalfrom failed")
	assert.Equal(t, obj, obj2, "GlobalConfigTransaction unmarshalfrom equality test")
	assert.Equal(t, len(data), l, "GlobalConfigTransaction data length check")
	assert.Equal(t, obj.Size(), l, "GlobalConfigTransaction data size check")
}

func TestGlobalConfigTransactionMarshalUnmarshalRandomData(t *testing.T) {
	seed := time.Now().Unix()
	fmt.Printf("Seed is %v\n", seed)
	rand.Seed(seed)

	for i := 0; i < 100; i++ {
		obj := &GlobalConfigTransaction{}

		obj.ActivationBlockHeight = uint64(rand.Uint64() & math.MaxUint64)

		obj.ScalarUpdates = []*GlobalConfigScalarUpdate{&GlobalConfigScalarUpdate{
			Key:   string(genRandom(rand.Int() % 20)),
			Value: []byte(genRandom(rand.Int() % 20)),
		}, &GlobalConfigScalarUpdate{
			Key:   string(genRandom(rand.Int() % 20)),
			Value: []byte(genRandom(rand.Int() % 20)),
		}, &GlobalConfigScalarUpdate{
			Key:   string(genRandom(rand.Int() % 20)),
			Value: []byte(genRandom(rand.Int() % 20)),
		}, &GlobalConfigScalarUpdate{
			Key:   string(genRandom(rand.Int() % 20)),
			Value: []byte(genRandom(rand.Int() % 20)),
		}, &GlobalConfigScalarUpdate{
			Key:   string(genRandom(rand.Int() % 20)),
			Value: []byte(genRandom(rand.Int() % 20)),
		}, &GlobalConfigScalarUpdate{
			Key:   string(genRandom(rand.Int() % 20)),
			Value: []byte(genRandom(rand.Int() % 20)),
		}, &GlobalConfigScalarUpdate{
			Key:   string(genRandom(rand.Int() % 20)),
			Value: []byte(genRandom(rand.Int() % 20)),
		}, &GlobalConfigScalarUpdate{
			Key:   string(genRandom(rand.Int() % 20)),
			Value: []byte(genRandom(rand.Int() % 20)),
		}, &GlobalConfigScalarUpdate{
			Key:   string(genRandom(rand.Int() % 20)),
			Value: []byte(genRandom(rand.Int() % 20)),
		}, &GlobalConfigScalarUpdate{
			Key:   string(genRandom(rand.Int() % 20)),
			Value: []byte(genRandom(rand.Int() % 20)),
		}}

		obj.ListUpdates = []*GlobalConfigListUpdate{&GlobalConfigListUpdate{
			Key:       string(genRandom(rand.Int() % 20)),
			Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
			Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}},
		}, &GlobalConfigListUpdate{
			Key:       string(genRandom(rand.Int() % 20)),
			Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
			Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}},
		}, &GlobalConfigListUpdate{
			Key:       string(genRandom(rand.Int() % 20)),
			Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
			Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}},
		}, &GlobalConfigListUpdate{
			Key:       string(genRandom(rand.Int() % 20)),
			Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
			Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}},
		}, &GlobalConfigListUpdate{
			Key:       string(genRandom(rand.Int() % 20)),
			Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
			Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}},
		}, &GlobalConfigListUpdate{
			Key:       string(genRandom(rand.Int() % 20)),
			Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
			Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}},
		}, &GlobalConfigListUpdate{
			Key:       string(genRandom(rand.Int() % 20)),
			Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
			Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}},
		}, &GlobalConfigListUpdate{
			Key:       string(genRandom(rand.Int() % 20)),
			Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
			Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}},
		}, &GlobalConfigListUpdate{
			Key:       string(genRandom(rand.Int() % 20)),
			Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
			Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}},
		}, &GlobalConfigListUpdate{
			Key:       string(genRandom(rand.Int() % 20)),
			Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
			Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigListInsertion{
				Index: uint64(rand.Uint64() & math.MaxUint64),
				Value: []byte(genRandom(rand.Int() % 20)),
			}},
		}}

		obj.SigPublicKey = []byte(genRandom(rand.Int() % 20))

		obj.Signature = []byte(genRandom(rand.Int() % 20))

		obj2 := &GlobalConfigTransaction{}

		data, err := obj.Marshal()
		assert.Nil(t, err, "marshal failed for GlobalConfigTransaction")
		assert.Equal(t, len(data), obj.Size(), "GlobalConfigTransaction size check on random values")
		assert.Nil(t, obj2.Unmarshal(data), "GlobalConfigTransaction random values unmarshal test")
		assert.Equal(t, obj, obj2, "GlobalConfigTransaction unmarshal equality test")

		obj2 = &GlobalConfigTransaction{}

		l, err := obj2.UnmarshalFrom(data)
		assert.Nil(t, err, fmt.Sprintf("GlobalConfigTransaction unmarshalfrom failed: %q", hex.EncodeToString(data)))
		assert.Equal(t, obj, obj2, "GlobalConfigTransaction unmarshalfrom equality test")
		assert.Equal(t, len(data), l, "GlobalConfigTransaction data length check")
		assert.Equal(t, obj.Size(), l, "GlobalConfigTransaction data size check")

		assert.True(t, reflect.DeepEqual(obj, obj2))
	}
}

func TestOutpointMarshalUnmarshalZeroValue(t *testing.T) {
	obj := &Outpoint{

		PreviousTx: [32]byte{},
		Index:      0}

	obj2 := &Outpoint{

		PreviousTx: [32]byte{},
		Index:      0}

	data, err := obj.Marshal()
	assert.Nil(t, err, "marshal failed for Outpoint")
	assert.Equal(t, len(data), obj.Size(), "Outpoint size check on zero value")
	assert.Nil(t, obj2.Unmarshal(data), "Outpoint zero value unmarshal test")
	assert.Equal(t, obj, obj2, "Outpoint unmarshal equality test")
	obj2 = &Outpoint{

		PreviousTx: [32]byte{},
		Index:      0}
	l, err := obj2.UnmarshalFrom(data)
	assert.Nil(t, err, "Outpoint unmarshalfrom failed")
	assert.Equal(t, obj, obj2, "Outpoint unmarshalfrom equality test")
	assert.Equal(t, len(data), l, "Outpoint data length check")
	assert.Equal(t, obj.Size(), l, "Outpoint data size check")
}

func TestOutpointMarshalUnmarshalRandomData(t *testing.T) {
	seed := time.Now().Unix()
	fmt.Printf("Seed is %v\n", seed)
	rand.Seed(seed)

	for i := 0; i < 100; i++ {
		obj := &Outpoint{}

		obj.PreviousTx = [32]byte{}

		obj.Index = uint8(rand.Uint64() & math.MaxUint8)

		obj2 := &Outpoint{}

		data, err := obj.Marshal()
		assert.Nil(t, err, "marshal failed for Outpoint")
		assert.Equal(t, len(data), obj.Size(), "Outpoint size check on random values")
		assert.Nil(t, obj2.Unmarshal(data), "Outpoint random values unmarshal test")
		assert.Equal(t, obj, obj2, "Outpoint unmarshal equality test")

		obj2 = &Outpoint{}

		l, err := obj2.UnmarshalFrom(data)
		assert.Nil(t, err, fmt.Sprintf("Outpoint unmarshalfrom failed: %q", hex.EncodeToString(data)))
		assert.Equal(t, obj, obj2, "Outpoint unmarshalfrom equality test")
		assert.Equal(t, len(data), l, "Outpoint data length check")
		assert.Equal(t, obj.Size(), l, "Outpoint data size check")

		assert.True(t, reflect.DeepEqual(obj, obj2))
	}
}

func TestTransactionMarshalUnmarshalZeroValue(t *testing.T) {
	obj := &Transaction{

		Version: 0,
		Body: []TransactionBody{&TransferTransaction{
			Inputs:    make([]*TransactionInput, 0),
			Outputs:   make([]*TransactionOutput, 0),
			Witnesses: make([]*TransactionWitness, 0),
			LockTime:  0}, &GenesisTransaction{
			Outputs: make([]*TransactionOutput, 0)}, &GlobalConfigTransaction{
			ActivationBlockHeight: 0,
			ScalarUpdates:         make([]*GlobalConfigScalarUpdate, 0),
			ListUpdates:           make([]*GlobalConfigListUpdate, 0),
			SigPublicKey:          make([]byte, 0),
			Signature:             make([]byte, 0)}, &EscrowOpenTransaction{}}[rand.Int()%4],
		Flags: 0}

	obj2 := &Transaction{

		Version: 0,
		Body: []TransactionBody{&TransferTransaction{
			Inputs:    make([]*TransactionInput, 0),
			Outputs:   make([]*TransactionOutput, 0),
			Witnesses: make([]*TransactionWitness, 0),
			LockTime:  0}, &GenesisTransaction{
			Outputs: make([]*TransactionOutput, 0)}, &GlobalConfigTransaction{
			ActivationBlockHeight: 0,
			ScalarUpdates:         make([]*GlobalConfigScalarUpdate, 0),
			ListUpdates:           make([]*GlobalConfigListUpdate, 0),
			SigPublicKey:          make([]byte, 0),
			Signature:             make([]byte, 0)}, &EscrowOpenTransaction{}}[rand.Int()%4],
		Flags: 0}

	data, err := obj.Marshal()
	assert.Nil(t, err, "marshal failed for Transaction")
	assert.Equal(t, len(data), obj.Size(), "Transaction size check on zero value")
	assert.Nil(t, obj2.Unmarshal(data), "Transaction zero value unmarshal test")
	assert.Equal(t, obj, obj2, "Transaction unmarshal equality test")
	obj2 = &Transaction{

		Version: 0,
		Body: []TransactionBody{&TransferTransaction{
			Inputs:    make([]*TransactionInput, 0),
			Outputs:   make([]*TransactionOutput, 0),
			Witnesses: make([]*TransactionWitness, 0),
			LockTime:  0}, &GenesisTransaction{
			Outputs: make([]*TransactionOutput, 0)}, &GlobalConfigTransaction{
			ActivationBlockHeight: 0,
			ScalarUpdates:         make([]*GlobalConfigScalarUpdate, 0),
			ListUpdates:           make([]*GlobalConfigListUpdate, 0),
			SigPublicKey:          make([]byte, 0),
			Signature:             make([]byte, 0)}, &EscrowOpenTransaction{}}[rand.Int()%4],
		Flags: 0}
	l, err := obj2.UnmarshalFrom(data)
	assert.Nil(t, err, "Transaction unmarshalfrom failed")
	assert.Equal(t, obj, obj2, "Transaction unmarshalfrom equality test")
	assert.Equal(t, len(data), l, "Transaction data length check")
	assert.Equal(t, obj.Size(), l, "Transaction data size check")
}

func TestTransactionMarshalUnmarshalRandomData(t *testing.T) {
	seed := time.Now().Unix()
	fmt.Printf("Seed is %v\n", seed)
	rand.Seed(seed)

	for i := 0; i < 100; i++ {
		obj := &Transaction{}

		obj.Version = uint8(rand.Uint64() & math.MaxUint8)

		obj.Body = []TransactionBody{&TransferTransaction{
			Inputs: []*TransactionInput{&TransactionInput{
				Outpoint: Outpoint{
					PreviousTx: [32]byte{},
					Index:      uint8(rand.Uint64() & math.MaxUint8),
				},
				ScriptSig:  []byte(genRandom(rand.Int() % 520)),
				SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
			}, &TransactionInput{
				Outpoint: Outpoint{
					PreviousTx: [32]byte{},
					Index:      uint8(rand.Uint64() & math.MaxUint8),
				},
				ScriptSig:  []byte(genRandom(rand.Int() % 520)),
				SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
			}, &TransactionInput{
				Outpoint: Outpoint{
					PreviousTx: [32]byte{},
					Index:      uint8(rand.Uint64() & math.MaxUint8),
				},
				ScriptSig:  []byte(genRandom(rand.Int() % 520)),
				SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
			}, &TransactionInput{
				Outpoint: Outpoint{
					PreviousTx: [32]byte{},
					Index:      uint8(rand.Uint64() & math.MaxUint8),
				},
				ScriptSig:  []byte(genRandom(rand.Int() % 520)),
				SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
			}, &TransactionInput{
				Outpoint: Outpoint{
					PreviousTx: [32]byte{},
					Index:      uint8(rand.Uint64() & math.MaxUint8),
				},
				ScriptSig:  []byte(genRandom(rand.Int() % 520)),
				SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
			}, &TransactionInput{
				Outpoint: Outpoint{
					PreviousTx: [32]byte{},
					Index:      uint8(rand.Uint64() & math.MaxUint8),
				},
				ScriptSig:  []byte(genRandom(rand.Int() % 520)),
				SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
			}, &TransactionInput{
				Outpoint: Outpoint{
					PreviousTx: [32]byte{},
					Index:      uint8(rand.Uint64() & math.MaxUint8),
				},
				ScriptSig:  []byte(genRandom(rand.Int() % 520)),
				SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
			}, &TransactionInput{
				Outpoint: Outpoint{
					PreviousTx: [32]byte{},
					Index:      uint8(rand.Uint64() & math.MaxUint8),
				},
				ScriptSig:  []byte(genRandom(rand.Int() % 520)),
				SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
			}, &TransactionInput{
				Outpoint: Outpoint{
					PreviousTx: [32]byte{},
					Index:      uint8(rand.Uint64() & math.MaxUint8),
				},
				ScriptSig:  []byte(genRandom(rand.Int() % 520)),
				SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
			}, &TransactionInput{
				Outpoint: Outpoint{
					PreviousTx: [32]byte{},
					Index:      uint8(rand.Uint64() & math.MaxUint8),
				},
				ScriptSig:  []byte(genRandom(rand.Int() % 520)),
				SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
			}},
			Outputs: []*TransactionOutput{&TransactionOutput{
				Value:        uint32(rand.Uint64() & math.MaxUint32),
				ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
			}, &TransactionOutput{
				Value:        uint32(rand.Uint64() & math.MaxUint32),
				ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
			}, &TransactionOutput{
				Value:        uint32(rand.Uint64() & math.MaxUint32),
				ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
			}, &TransactionOutput{
				Value:        uint32(rand.Uint64() & math.MaxUint32),
				ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
			}, &TransactionOutput{
				Value:        uint32(rand.Uint64() & math.MaxUint32),
				ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
			}, &TransactionOutput{
				Value:        uint32(rand.Uint64() & math.MaxUint32),
				ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
			}, &TransactionOutput{
				Value:        uint32(rand.Uint64() & math.MaxUint32),
				ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
			}, &TransactionOutput{
				Value:        uint32(rand.Uint64() & math.MaxUint32),
				ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
			}, &TransactionOutput{
				Value:        uint32(rand.Uint64() & math.MaxUint32),
				ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
			}, &TransactionOutput{
				Value:        uint32(rand.Uint64() & math.MaxUint32),
				ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
			}},
			Witnesses: []*TransactionWitness{&TransactionWitness{
				Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
			}, &TransactionWitness{
				Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
			}, &TransactionWitness{
				Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
			}, &TransactionWitness{
				Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
			}, &TransactionWitness{
				Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
			}, &TransactionWitness{
				Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
			}, &TransactionWitness{
				Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
			}, &TransactionWitness{
				Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
			}, &TransactionWitness{
				Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
			}, &TransactionWitness{
				Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
			}},
			LockTime: 0}, &GenesisTransaction{
			Outputs: []*TransactionOutput{&TransactionOutput{
				Value:        uint32(rand.Uint64() & math.MaxUint32),
				ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
			}, &TransactionOutput{
				Value:        uint32(rand.Uint64() & math.MaxUint32),
				ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
			}, &TransactionOutput{
				Value:        uint32(rand.Uint64() & math.MaxUint32),
				ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
			}, &TransactionOutput{
				Value:        uint32(rand.Uint64() & math.MaxUint32),
				ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
			}, &TransactionOutput{
				Value:        uint32(rand.Uint64() & math.MaxUint32),
				ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
			}, &TransactionOutput{
				Value:        uint32(rand.Uint64() & math.MaxUint32),
				ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
			}, &TransactionOutput{
				Value:        uint32(rand.Uint64() & math.MaxUint32),
				ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
			}, &TransactionOutput{
				Value:        uint32(rand.Uint64() & math.MaxUint32),
				ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
			}, &TransactionOutput{
				Value:        uint32(rand.Uint64() & math.MaxUint32),
				ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
			}, &TransactionOutput{
				Value:        uint32(rand.Uint64() & math.MaxUint32),
				ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
			}}}, &GlobalConfigTransaction{
			ActivationBlockHeight: uint64(rand.Uint64() & math.MaxUint64),
			ScalarUpdates: []*GlobalConfigScalarUpdate{&GlobalConfigScalarUpdate{
				Key:   string(genRandom(rand.Int() % 20)),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigScalarUpdate{
				Key:   string(genRandom(rand.Int() % 20)),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigScalarUpdate{
				Key:   string(genRandom(rand.Int() % 20)),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigScalarUpdate{
				Key:   string(genRandom(rand.Int() % 20)),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigScalarUpdate{
				Key:   string(genRandom(rand.Int() % 20)),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigScalarUpdate{
				Key:   string(genRandom(rand.Int() % 20)),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigScalarUpdate{
				Key:   string(genRandom(rand.Int() % 20)),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigScalarUpdate{
				Key:   string(genRandom(rand.Int() % 20)),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigScalarUpdate{
				Key:   string(genRandom(rand.Int() % 20)),
				Value: []byte(genRandom(rand.Int() % 20)),
			}, &GlobalConfigScalarUpdate{
				Key:   string(genRandom(rand.Int() % 20)),
				Value: []byte(genRandom(rand.Int() % 20)),
			}},
			ListUpdates: []*GlobalConfigListUpdate{&GlobalConfigListUpdate{
				Key:       string(genRandom(rand.Int() % 20)),
				Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
				Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}},
			}, &GlobalConfigListUpdate{
				Key:       string(genRandom(rand.Int() % 20)),
				Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
				Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}},
			}, &GlobalConfigListUpdate{
				Key:       string(genRandom(rand.Int() % 20)),
				Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
				Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}},
			}, &GlobalConfigListUpdate{
				Key:       string(genRandom(rand.Int() % 20)),
				Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
				Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}},
			}, &GlobalConfigListUpdate{
				Key:       string(genRandom(rand.Int() % 20)),
				Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
				Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}},
			}, &GlobalConfigListUpdate{
				Key:       string(genRandom(rand.Int() % 20)),
				Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
				Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}},
			}, &GlobalConfigListUpdate{
				Key:       string(genRandom(rand.Int() % 20)),
				Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
				Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}},
			}, &GlobalConfigListUpdate{
				Key:       string(genRandom(rand.Int() % 20)),
				Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
				Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}},
			}, &GlobalConfigListUpdate{
				Key:       string(genRandom(rand.Int() % 20)),
				Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
				Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}},
			}, &GlobalConfigListUpdate{
				Key:       string(genRandom(rand.Int() % 20)),
				Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
				Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigListInsertion{
					Index: uint64(rand.Uint64() & math.MaxUint64),
					Value: []byte(genRandom(rand.Int() % 20)),
				}},
			}},
			SigPublicKey: []byte(genRandom(rand.Int() % 20)),
			Signature:    []byte(genRandom(rand.Int() % 20))}, &EscrowOpenTransaction{}}[rand.Int()%4]

		obj.Flags = uint16(rand.Uint64() & math.MaxUint16)

		obj2 := &Transaction{}

		data, err := obj.Marshal()
		assert.Nil(t, err, "marshal failed for Transaction")
		assert.Equal(t, len(data), obj.Size(), "Transaction size check on random values")
		assert.Nil(t, obj2.Unmarshal(data), "Transaction random values unmarshal test")
		assert.Equal(t, obj, obj2, "Transaction unmarshal equality test")

		obj2 = &Transaction{}

		l, err := obj2.UnmarshalFrom(data)
		assert.Nil(t, err, fmt.Sprintf("Transaction unmarshalfrom failed: %q", hex.EncodeToString(data)))
		assert.Equal(t, obj, obj2, "Transaction unmarshalfrom equality test")
		assert.Equal(t, len(data), l, "Transaction data length check")
		assert.Equal(t, obj.Size(), l, "Transaction data size check")

		assert.True(t, reflect.DeepEqual(obj, obj2))
	}
}

func TestTransactionInputMarshalUnmarshalZeroValue(t *testing.T) {
	obj := &TransactionInput{

		Outpoint: Outpoint{
			PreviousTx: [32]byte{},
			Index:      0,
		},
		ScriptSig:  make([]byte, 0),
		SequenceNo: 0}

	obj2 := &TransactionInput{

		Outpoint: Outpoint{
			PreviousTx: [32]byte{},
			Index:      0,
		},
		ScriptSig:  make([]byte, 0),
		SequenceNo: 0}

	data, err := obj.Marshal()
	assert.Nil(t, err, "marshal failed for TransactionInput")
	assert.Equal(t, len(data), obj.Size(), "TransactionInput size check on zero value")
	assert.Nil(t, obj2.Unmarshal(data), "TransactionInput zero value unmarshal test")
	assert.Equal(t, obj, obj2, "TransactionInput unmarshal equality test")
	obj2 = &TransactionInput{

		Outpoint: Outpoint{
			PreviousTx: [32]byte{},
			Index:      0,
		},
		ScriptSig:  make([]byte, 0),
		SequenceNo: 0}
	l, err := obj2.UnmarshalFrom(data)
	assert.Nil(t, err, "TransactionInput unmarshalfrom failed")
	assert.Equal(t, obj, obj2, "TransactionInput unmarshalfrom equality test")
	assert.Equal(t, len(data), l, "TransactionInput data length check")
	assert.Equal(t, obj.Size(), l, "TransactionInput data size check")
}

func TestTransactionInputMarshalUnmarshalRandomData(t *testing.T) {
	seed := time.Now().Unix()
	fmt.Printf("Seed is %v\n", seed)
	rand.Seed(seed)

	for i := 0; i < 100; i++ {
		obj := &TransactionInput{}

		obj.Outpoint = Outpoint{
			PreviousTx: [32]byte{},
			Index:      uint8(rand.Uint64() & math.MaxUint8),
		}

		obj.ScriptSig = []byte(genRandom(rand.Int() % 520))

		obj.SequenceNo = uint32(rand.Uint64() & math.MaxUint32)

		obj2 := &TransactionInput{}

		data, err := obj.Marshal()
		assert.Nil(t, err, "marshal failed for TransactionInput")
		assert.Equal(t, len(data), obj.Size(), "TransactionInput size check on random values")
		assert.Nil(t, obj2.Unmarshal(data), "TransactionInput random values unmarshal test")
		assert.Equal(t, obj, obj2, "TransactionInput unmarshal equality test")

		obj2 = &TransactionInput{}

		l, err := obj2.UnmarshalFrom(data)
		assert.Nil(t, err, fmt.Sprintf("TransactionInput unmarshalfrom failed: %q", hex.EncodeToString(data)))
		assert.Equal(t, obj, obj2, "TransactionInput unmarshalfrom equality test")
		assert.Equal(t, len(data), l, "TransactionInput data length check")
		assert.Equal(t, obj.Size(), l, "TransactionInput data size check")

		assert.True(t, reflect.DeepEqual(obj, obj2))
	}
}

func TestTransactionOutputMarshalUnmarshalZeroValue(t *testing.T) {
	obj := &TransactionOutput{

		Value:        0,
		ScriptPubKey: make([]byte, 0)}

	obj2 := &TransactionOutput{

		Value:        0,
		ScriptPubKey: make([]byte, 0)}

	data, err := obj.Marshal()
	assert.Nil(t, err, "marshal failed for TransactionOutput")
	assert.Equal(t, len(data), obj.Size(), "TransactionOutput size check on zero value")
	assert.Nil(t, obj2.Unmarshal(data), "TransactionOutput zero value unmarshal test")
	assert.Equal(t, obj, obj2, "TransactionOutput unmarshal equality test")
	obj2 = &TransactionOutput{

		Value:        0,
		ScriptPubKey: make([]byte, 0)}
	l, err := obj2.UnmarshalFrom(data)
	assert.Nil(t, err, "TransactionOutput unmarshalfrom failed")
	assert.Equal(t, obj, obj2, "TransactionOutput unmarshalfrom equality test")
	assert.Equal(t, len(data), l, "TransactionOutput data length check")
	assert.Equal(t, obj.Size(), l, "TransactionOutput data size check")
}

func TestTransactionOutputMarshalUnmarshalRandomData(t *testing.T) {
	seed := time.Now().Unix()
	fmt.Printf("Seed is %v\n", seed)
	rand.Seed(seed)

	for i := 0; i < 100; i++ {
		obj := &TransactionOutput{}

		obj.Value = uint32(rand.Uint64() & math.MaxUint32)

		obj.ScriptPubKey = []byte(genRandom(rand.Int() % 20))

		obj2 := &TransactionOutput{}

		data, err := obj.Marshal()
		assert.Nil(t, err, "marshal failed for TransactionOutput")
		assert.Equal(t, len(data), obj.Size(), "TransactionOutput size check on random values")
		assert.Nil(t, obj2.Unmarshal(data), "TransactionOutput random values unmarshal test")
		assert.Equal(t, obj, obj2, "TransactionOutput unmarshal equality test")

		obj2 = &TransactionOutput{}

		l, err := obj2.UnmarshalFrom(data)
		assert.Nil(t, err, fmt.Sprintf("TransactionOutput unmarshalfrom failed: %q", hex.EncodeToString(data)))
		assert.Equal(t, obj, obj2, "TransactionOutput unmarshalfrom equality test")
		assert.Equal(t, len(data), l, "TransactionOutput data length check")
		assert.Equal(t, obj.Size(), l, "TransactionOutput data size check")

		assert.True(t, reflect.DeepEqual(obj, obj2))
	}
}

func TestTransactionWitnessMarshalUnmarshalZeroValue(t *testing.T) {
	obj := &TransactionWitness{

		Data: make([][32]byte, 0)}

	obj2 := &TransactionWitness{

		Data: make([][32]byte, 0)}

	data, err := obj.Marshal()
	assert.Nil(t, err, "marshal failed for TransactionWitness")
	assert.Equal(t, len(data), obj.Size(), "TransactionWitness size check on zero value")
	assert.Nil(t, obj2.Unmarshal(data), "TransactionWitness zero value unmarshal test")
	assert.Equal(t, obj, obj2, "TransactionWitness unmarshal equality test")
	obj2 = &TransactionWitness{

		Data: make([][32]byte, 0)}
	l, err := obj2.UnmarshalFrom(data)
	assert.Nil(t, err, "TransactionWitness unmarshalfrom failed")
	assert.Equal(t, obj, obj2, "TransactionWitness unmarshalfrom equality test")
	assert.Equal(t, len(data), l, "TransactionWitness data length check")
	assert.Equal(t, obj.Size(), l, "TransactionWitness data size check")
}

func TestTransactionWitnessMarshalUnmarshalRandomData(t *testing.T) {
	seed := time.Now().Unix()
	fmt.Printf("Seed is %v\n", seed)
	rand.Seed(seed)

	for i := 0; i < 100; i++ {
		obj := &TransactionWitness{}

		obj.Data = [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}}

		obj2 := &TransactionWitness{}

		data, err := obj.Marshal()
		assert.Nil(t, err, "marshal failed for TransactionWitness")
		assert.Equal(t, len(data), obj.Size(), "TransactionWitness size check on random values")
		assert.Nil(t, obj2.Unmarshal(data), "TransactionWitness random values unmarshal test")
		assert.Equal(t, obj, obj2, "TransactionWitness unmarshal equality test")

		obj2 = &TransactionWitness{}

		l, err := obj2.UnmarshalFrom(data)
		assert.Nil(t, err, fmt.Sprintf("TransactionWitness unmarshalfrom failed: %q", hex.EncodeToString(data)))
		assert.Equal(t, obj, obj2, "TransactionWitness unmarshalfrom equality test")
		assert.Equal(t, len(data), l, "TransactionWitness data length check")
		assert.Equal(t, obj.Size(), l, "TransactionWitness data size check")

		assert.True(t, reflect.DeepEqual(obj, obj2))
	}
}

func TestTransactionsMarshalUnmarshalZeroValue(t *testing.T) {
	obj := &Transactions{

		Transactions: make([]*Transaction, 0)}

	obj2 := &Transactions{

		Transactions: make([]*Transaction, 0)}

	data, err := obj.Marshal()
	assert.Nil(t, err, "marshal failed for Transactions")
	assert.Equal(t, len(data), obj.Size(), "Transactions size check on zero value")
	assert.Nil(t, obj2.Unmarshal(data), "Transactions zero value unmarshal test")
	assert.Equal(t, obj, obj2, "Transactions unmarshal equality test")
	obj2 = &Transactions{

		Transactions: make([]*Transaction, 0)}
	l, err := obj2.UnmarshalFrom(data)
	assert.Nil(t, err, "Transactions unmarshalfrom failed")
	assert.Equal(t, obj, obj2, "Transactions unmarshalfrom equality test")
	assert.Equal(t, len(data), l, "Transactions data length check")
	assert.Equal(t, obj.Size(), l, "Transactions data size check")
}

func TestTransactionsMarshalUnmarshalRandomData(t *testing.T) {
	seed := time.Now().Unix()
	fmt.Printf("Seed is %v\n", seed)
	rand.Seed(seed)

	for i := 0; i < 100; i++ {
		obj := &Transactions{}

		obj.Transactions = []*Transaction{&Transaction{
			Version: uint8(rand.Uint64() & math.MaxUint8),
			Body: []TransactionBody{&TransferTransaction{
				Inputs: []*TransactionInput{&TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}},
				Outputs: []*TransactionOutput{&TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}},
				Witnesses: []*TransactionWitness{&TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}},
				LockTime: 0}, &GenesisTransaction{
				Outputs: []*TransactionOutput{&TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}}}, &GlobalConfigTransaction{
				ActivationBlockHeight: uint64(rand.Uint64() & math.MaxUint64),
				ScalarUpdates: []*GlobalConfigScalarUpdate{&GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}},
				ListUpdates: []*GlobalConfigListUpdate{&GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}},
				SigPublicKey: []byte(genRandom(rand.Int() % 20)),
				Signature:    []byte(genRandom(rand.Int() % 20))}, &EscrowOpenTransaction{}}[rand.Int()%4],
			Flags: uint16(rand.Uint64() & math.MaxUint16),
		}, &Transaction{
			Version: uint8(rand.Uint64() & math.MaxUint8),
			Body: []TransactionBody{&TransferTransaction{
				Inputs: []*TransactionInput{&TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}},
				Outputs: []*TransactionOutput{&TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}},
				Witnesses: []*TransactionWitness{&TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}},
				LockTime: 0}, &GenesisTransaction{
				Outputs: []*TransactionOutput{&TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}}}, &GlobalConfigTransaction{
				ActivationBlockHeight: uint64(rand.Uint64() & math.MaxUint64),
				ScalarUpdates: []*GlobalConfigScalarUpdate{&GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}},
				ListUpdates: []*GlobalConfigListUpdate{&GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}},
				SigPublicKey: []byte(genRandom(rand.Int() % 20)),
				Signature:    []byte(genRandom(rand.Int() % 20))}, &EscrowOpenTransaction{}}[rand.Int()%4],
			Flags: uint16(rand.Uint64() & math.MaxUint16),
		}, &Transaction{
			Version: uint8(rand.Uint64() & math.MaxUint8),
			Body: []TransactionBody{&TransferTransaction{
				Inputs: []*TransactionInput{&TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}},
				Outputs: []*TransactionOutput{&TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}},
				Witnesses: []*TransactionWitness{&TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}},
				LockTime: 0}, &GenesisTransaction{
				Outputs: []*TransactionOutput{&TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}}}, &GlobalConfigTransaction{
				ActivationBlockHeight: uint64(rand.Uint64() & math.MaxUint64),
				ScalarUpdates: []*GlobalConfigScalarUpdate{&GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}},
				ListUpdates: []*GlobalConfigListUpdate{&GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}},
				SigPublicKey: []byte(genRandom(rand.Int() % 20)),
				Signature:    []byte(genRandom(rand.Int() % 20))}, &EscrowOpenTransaction{}}[rand.Int()%4],
			Flags: uint16(rand.Uint64() & math.MaxUint16),
		}, &Transaction{
			Version: uint8(rand.Uint64() & math.MaxUint8),
			Body: []TransactionBody{&TransferTransaction{
				Inputs: []*TransactionInput{&TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}},
				Outputs: []*TransactionOutput{&TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}},
				Witnesses: []*TransactionWitness{&TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}},
				LockTime: 0}, &GenesisTransaction{
				Outputs: []*TransactionOutput{&TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}}}, &GlobalConfigTransaction{
				ActivationBlockHeight: uint64(rand.Uint64() & math.MaxUint64),
				ScalarUpdates: []*GlobalConfigScalarUpdate{&GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}},
				ListUpdates: []*GlobalConfigListUpdate{&GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}},
				SigPublicKey: []byte(genRandom(rand.Int() % 20)),
				Signature:    []byte(genRandom(rand.Int() % 20))}, &EscrowOpenTransaction{}}[rand.Int()%4],
			Flags: uint16(rand.Uint64() & math.MaxUint16),
		}, &Transaction{
			Version: uint8(rand.Uint64() & math.MaxUint8),
			Body: []TransactionBody{&TransferTransaction{
				Inputs: []*TransactionInput{&TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}},
				Outputs: []*TransactionOutput{&TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}},
				Witnesses: []*TransactionWitness{&TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}},
				LockTime: 0}, &GenesisTransaction{
				Outputs: []*TransactionOutput{&TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}}}, &GlobalConfigTransaction{
				ActivationBlockHeight: uint64(rand.Uint64() & math.MaxUint64),
				ScalarUpdates: []*GlobalConfigScalarUpdate{&GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}},
				ListUpdates: []*GlobalConfigListUpdate{&GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}},
				SigPublicKey: []byte(genRandom(rand.Int() % 20)),
				Signature:    []byte(genRandom(rand.Int() % 20))}, &EscrowOpenTransaction{}}[rand.Int()%4],
			Flags: uint16(rand.Uint64() & math.MaxUint16),
		}, &Transaction{
			Version: uint8(rand.Uint64() & math.MaxUint8),
			Body: []TransactionBody{&TransferTransaction{
				Inputs: []*TransactionInput{&TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}},
				Outputs: []*TransactionOutput{&TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}},
				Witnesses: []*TransactionWitness{&TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}},
				LockTime: 0}, &GenesisTransaction{
				Outputs: []*TransactionOutput{&TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}}}, &GlobalConfigTransaction{
				ActivationBlockHeight: uint64(rand.Uint64() & math.MaxUint64),
				ScalarUpdates: []*GlobalConfigScalarUpdate{&GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}},
				ListUpdates: []*GlobalConfigListUpdate{&GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}},
				SigPublicKey: []byte(genRandom(rand.Int() % 20)),
				Signature:    []byte(genRandom(rand.Int() % 20))}, &EscrowOpenTransaction{}}[rand.Int()%4],
			Flags: uint16(rand.Uint64() & math.MaxUint16),
		}, &Transaction{
			Version: uint8(rand.Uint64() & math.MaxUint8),
			Body: []TransactionBody{&TransferTransaction{
				Inputs: []*TransactionInput{&TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}},
				Outputs: []*TransactionOutput{&TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}},
				Witnesses: []*TransactionWitness{&TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}},
				LockTime: 0}, &GenesisTransaction{
				Outputs: []*TransactionOutput{&TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}}}, &GlobalConfigTransaction{
				ActivationBlockHeight: uint64(rand.Uint64() & math.MaxUint64),
				ScalarUpdates: []*GlobalConfigScalarUpdate{&GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}},
				ListUpdates: []*GlobalConfigListUpdate{&GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}},
				SigPublicKey: []byte(genRandom(rand.Int() % 20)),
				Signature:    []byte(genRandom(rand.Int() % 20))}, &EscrowOpenTransaction{}}[rand.Int()%4],
			Flags: uint16(rand.Uint64() & math.MaxUint16),
		}, &Transaction{
			Version: uint8(rand.Uint64() & math.MaxUint8),
			Body: []TransactionBody{&TransferTransaction{
				Inputs: []*TransactionInput{&TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}},
				Outputs: []*TransactionOutput{&TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}},
				Witnesses: []*TransactionWitness{&TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}},
				LockTime: 0}, &GenesisTransaction{
				Outputs: []*TransactionOutput{&TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}}}, &GlobalConfigTransaction{
				ActivationBlockHeight: uint64(rand.Uint64() & math.MaxUint64),
				ScalarUpdates: []*GlobalConfigScalarUpdate{&GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}},
				ListUpdates: []*GlobalConfigListUpdate{&GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}},
				SigPublicKey: []byte(genRandom(rand.Int() % 20)),
				Signature:    []byte(genRandom(rand.Int() % 20))}, &EscrowOpenTransaction{}}[rand.Int()%4],
			Flags: uint16(rand.Uint64() & math.MaxUint16),
		}, &Transaction{
			Version: uint8(rand.Uint64() & math.MaxUint8),
			Body: []TransactionBody{&TransferTransaction{
				Inputs: []*TransactionInput{&TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}},
				Outputs: []*TransactionOutput{&TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}},
				Witnesses: []*TransactionWitness{&TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}},
				LockTime: 0}, &GenesisTransaction{
				Outputs: []*TransactionOutput{&TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}}}, &GlobalConfigTransaction{
				ActivationBlockHeight: uint64(rand.Uint64() & math.MaxUint64),
				ScalarUpdates: []*GlobalConfigScalarUpdate{&GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}},
				ListUpdates: []*GlobalConfigListUpdate{&GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}},
				SigPublicKey: []byte(genRandom(rand.Int() % 20)),
				Signature:    []byte(genRandom(rand.Int() % 20))}, &EscrowOpenTransaction{}}[rand.Int()%4],
			Flags: uint16(rand.Uint64() & math.MaxUint16),
		}, &Transaction{
			Version: uint8(rand.Uint64() & math.MaxUint8),
			Body: []TransactionBody{&TransferTransaction{
				Inputs: []*TransactionInput{&TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}, &TransactionInput{
					Outpoint: Outpoint{
						PreviousTx: [32]byte{},
						Index:      uint8(rand.Uint64() & math.MaxUint8),
					},
					ScriptSig:  []byte(genRandom(rand.Int() % 520)),
					SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
				}},
				Outputs: []*TransactionOutput{&TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}},
				Witnesses: []*TransactionWitness{&TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}, &TransactionWitness{
					Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
				}},
				LockTime: 0}, &GenesisTransaction{
				Outputs: []*TransactionOutput{&TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}, &TransactionOutput{
					Value:        uint32(rand.Uint64() & math.MaxUint32),
					ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
				}}}, &GlobalConfigTransaction{
				ActivationBlockHeight: uint64(rand.Uint64() & math.MaxUint64),
				ScalarUpdates: []*GlobalConfigScalarUpdate{&GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}, &GlobalConfigScalarUpdate{
					Key:   string(genRandom(rand.Int() % 20)),
					Value: []byte(genRandom(rand.Int() % 20)),
				}},
				ListUpdates: []*GlobalConfigListUpdate{&GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}, &GlobalConfigListUpdate{
					Key:       string(genRandom(rand.Int() % 20)),
					Deletions: []uint64{uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64), uint64(rand.Uint64() & math.MaxUint64)},
					Insertions: []*GlobalConfigListInsertion{&GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}, &GlobalConfigListInsertion{
						Index: uint64(rand.Uint64() & math.MaxUint64),
						Value: []byte(genRandom(rand.Int() % 20)),
					}},
				}},
				SigPublicKey: []byte(genRandom(rand.Int() % 20)),
				Signature:    []byte(genRandom(rand.Int() % 20))}, &EscrowOpenTransaction{}}[rand.Int()%4],
			Flags: uint16(rand.Uint64() & math.MaxUint16),
		}}

		obj2 := &Transactions{}

		data, err := obj.Marshal()
		assert.Nil(t, err, "marshal failed for Transactions")
		assert.Equal(t, len(data), obj.Size(), "Transactions size check on random values")
		assert.Nil(t, obj2.Unmarshal(data), "Transactions random values unmarshal test")
		assert.Equal(t, obj, obj2, "Transactions unmarshal equality test")

		obj2 = &Transactions{}

		l, err := obj2.UnmarshalFrom(data)
		assert.Nil(t, err, fmt.Sprintf("Transactions unmarshalfrom failed: %q", hex.EncodeToString(data)))
		assert.Equal(t, obj, obj2, "Transactions unmarshalfrom equality test")
		assert.Equal(t, len(data), l, "Transactions data length check")
		assert.Equal(t, obj.Size(), l, "Transactions data size check")

		assert.True(t, reflect.DeepEqual(obj, obj2))
	}
}

func TestTransferTransactionMarshalUnmarshalZeroValue(t *testing.T) {
	obj := &TransferTransaction{

		Inputs:    make([]*TransactionInput, 0),
		Outputs:   make([]*TransactionOutput, 0),
		Witnesses: make([]*TransactionWitness, 0)}

	obj2 := &TransferTransaction{

		Inputs:    make([]*TransactionInput, 0),
		Outputs:   make([]*TransactionOutput, 0),
		Witnesses: make([]*TransactionWitness, 0)}

	data, err := obj.Marshal()
	assert.Nil(t, err, "marshal failed for TransferTransaction")
	assert.Equal(t, len(data), obj.Size(), "TransferTransaction size check on zero value")
	assert.Nil(t, obj2.Unmarshal(data), "TransferTransaction zero value unmarshal test")
	assert.Equal(t, obj, obj2, "TransferTransaction unmarshal equality test")
	obj2 = &TransferTransaction{

		Inputs:    make([]*TransactionInput, 0),
		Outputs:   make([]*TransactionOutput, 0),
		Witnesses: make([]*TransactionWitness, 0)}
	l, err := obj2.UnmarshalFrom(data)
	assert.Nil(t, err, "TransferTransaction unmarshalfrom failed")
	assert.Equal(t, obj, obj2, "TransferTransaction unmarshalfrom equality test")
	assert.Equal(t, len(data), l, "TransferTransaction data length check")
	assert.Equal(t, obj.Size(), l, "TransferTransaction data size check")
}

func TestTransferTransactionMarshalUnmarshalRandomData(t *testing.T) {
	seed := time.Now().Unix()
	fmt.Printf("Seed is %v\n", seed)
	rand.Seed(seed)

	for i := 0; i < 100; i++ {
		obj := &TransferTransaction{}

		obj.Inputs = []*TransactionInput{&TransactionInput{
			Outpoint: Outpoint{
				PreviousTx: [32]byte{},
				Index:      uint8(rand.Uint64() & math.MaxUint8),
			},
			ScriptSig:  []byte(genRandom(rand.Int() % 520)),
			SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
		}, &TransactionInput{
			Outpoint: Outpoint{
				PreviousTx: [32]byte{},
				Index:      uint8(rand.Uint64() & math.MaxUint8),
			},
			ScriptSig:  []byte(genRandom(rand.Int() % 520)),
			SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
		}, &TransactionInput{
			Outpoint: Outpoint{
				PreviousTx: [32]byte{},
				Index:      uint8(rand.Uint64() & math.MaxUint8),
			},
			ScriptSig:  []byte(genRandom(rand.Int() % 520)),
			SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
		}, &TransactionInput{
			Outpoint: Outpoint{
				PreviousTx: [32]byte{},
				Index:      uint8(rand.Uint64() & math.MaxUint8),
			},
			ScriptSig:  []byte(genRandom(rand.Int() % 520)),
			SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
		}, &TransactionInput{
			Outpoint: Outpoint{
				PreviousTx: [32]byte{},
				Index:      uint8(rand.Uint64() & math.MaxUint8),
			},
			ScriptSig:  []byte(genRandom(rand.Int() % 520)),
			SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
		}, &TransactionInput{
			Outpoint: Outpoint{
				PreviousTx: [32]byte{},
				Index:      uint8(rand.Uint64() & math.MaxUint8),
			},
			ScriptSig:  []byte(genRandom(rand.Int() % 520)),
			SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
		}, &TransactionInput{
			Outpoint: Outpoint{
				PreviousTx: [32]byte{},
				Index:      uint8(rand.Uint64() & math.MaxUint8),
			},
			ScriptSig:  []byte(genRandom(rand.Int() % 520)),
			SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
		}, &TransactionInput{
			Outpoint: Outpoint{
				PreviousTx: [32]byte{},
				Index:      uint8(rand.Uint64() & math.MaxUint8),
			},
			ScriptSig:  []byte(genRandom(rand.Int() % 520)),
			SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
		}, &TransactionInput{
			Outpoint: Outpoint{
				PreviousTx: [32]byte{},
				Index:      uint8(rand.Uint64() & math.MaxUint8),
			},
			ScriptSig:  []byte(genRandom(rand.Int() % 520)),
			SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
		}, &TransactionInput{
			Outpoint: Outpoint{
				PreviousTx: [32]byte{},
				Index:      uint8(rand.Uint64() & math.MaxUint8),
			},
			ScriptSig:  []byte(genRandom(rand.Int() % 520)),
			SequenceNo: uint32(rand.Uint64() & math.MaxUint32),
		}}

		obj.Outputs = []*TransactionOutput{&TransactionOutput{
			Value:        uint32(rand.Uint64() & math.MaxUint32),
			ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
		}, &TransactionOutput{
			Value:        uint32(rand.Uint64() & math.MaxUint32),
			ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
		}, &TransactionOutput{
			Value:        uint32(rand.Uint64() & math.MaxUint32),
			ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
		}, &TransactionOutput{
			Value:        uint32(rand.Uint64() & math.MaxUint32),
			ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
		}, &TransactionOutput{
			Value:        uint32(rand.Uint64() & math.MaxUint32),
			ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
		}, &TransactionOutput{
			Value:        uint32(rand.Uint64() & math.MaxUint32),
			ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
		}, &TransactionOutput{
			Value:        uint32(rand.Uint64() & math.MaxUint32),
			ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
		}, &TransactionOutput{
			Value:        uint32(rand.Uint64() & math.MaxUint32),
			ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
		}, &TransactionOutput{
			Value:        uint32(rand.Uint64() & math.MaxUint32),
			ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
		}, &TransactionOutput{
			Value:        uint32(rand.Uint64() & math.MaxUint32),
			ScriptPubKey: []byte(genRandom(rand.Int() % 20)),
		}}

		obj.Witnesses = []*TransactionWitness{&TransactionWitness{
			Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
		}, &TransactionWitness{
			Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
		}, &TransactionWitness{
			Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
		}, &TransactionWitness{
			Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
		}, &TransactionWitness{
			Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
		}, &TransactionWitness{
			Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
		}, &TransactionWitness{
			Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
		}, &TransactionWitness{
			Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
		}, &TransactionWitness{
			Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
		}, &TransactionWitness{
			Data: [][32]byte{[32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}, [32]byte{}},
		}}

		obj2 := &TransferTransaction{}

		data, err := obj.Marshal()
		assert.Nil(t, err, "marshal failed for TransferTransaction")
		assert.Equal(t, len(data), obj.Size(), "TransferTransaction size check on random values")
		assert.Nil(t, obj2.Unmarshal(data), "TransferTransaction random values unmarshal test")
		assert.Equal(t, obj, obj2, "TransferTransaction unmarshal equality test")

		obj2 = &TransferTransaction{}

		l, err := obj2.UnmarshalFrom(data)
		assert.Nil(t, err, fmt.Sprintf("TransferTransaction unmarshalfrom failed: %q", hex.EncodeToString(data)))
		assert.Equal(t, obj, obj2, "TransferTransaction unmarshalfrom equality test")
		assert.Equal(t, len(data), l, "TransferTransaction data length check")
		assert.Equal(t, obj.Size(), l, "TransferTransaction data size check")

		assert.True(t, reflect.DeepEqual(obj, obj2))
	}
}
