// Package pkg is for testing ranger
package pkg

import (
	"encoding/binary"
	"math"

	"github.com/cachecashproject/go-cachecash/ranger"
	"github.com/pkg/errors"
)

type EscrowOpenTransaction struct {
}

// Marshal returns a byte array containing the marshaled representation of EscrowOpenTransaction, or nil and an error.
func (obj *EscrowOpenTransaction) Marshal() ([]byte, error) {
	data := make([]byte, obj.Size())
	n, err := obj.MarshalTo(data)
	if err != nil {
		return nil, errors.Wrap(err, "EscrowOpenTransaction")
	}

	if n != len(data) {
		return nil, errors.Wrapf(ranger.ErrMarshalLength, "%s %d %d", "EscrowOpenTransaction", n, len(data))
	}

	return data, nil
}

// MarshalTo accepts a byte array with pre-allocated space (see Size()) for EscrowOpenTransaction.
// It returns how many bytes it wrote to the array, or 0 and an error.
func (obj *EscrowOpenTransaction) MarshalTo(data []byte) (int, error) {

	var n int

	return n, nil
}

// Size returns the computed size of EscrowOpenTransaction as would-be marshaled
// without actually marshaling it.
func (obj *EscrowOpenTransaction) Size() int {
	var n int
	return n
}

// Unmarshal accepts EscrowOpenTransaction's binary representation and transforms the
// EscrowOpenTransaction used as the object. It returns any error.
func (obj *EscrowOpenTransaction) Unmarshal(data []byte) error {
	_, err := obj.UnmarshalFrom(data)
	return err
}

// UnmarshalFrom is very similar to Unmarshal, but also returns the count of data it read.
func (obj *EscrowOpenTransaction) UnmarshalFrom(data []byte) (int, error) {
	var n int

	return n, nil
}

type GenesisTransaction struct {
	Outputs []*TransactionOutput
}

// Marshal returns a byte array containing the marshaled representation of GenesisTransaction, or nil and an error.
func (obj *GenesisTransaction) Marshal() ([]byte, error) {
	data := make([]byte, obj.Size())
	n, err := obj.MarshalTo(data)
	if err != nil {
		return nil, errors.Wrap(err, "GenesisTransaction")
	}

	if n != len(data) {
		return nil, errors.Wrapf(ranger.ErrMarshalLength, "%s %d %d", "GenesisTransaction", n, len(data))
	}

	return data, nil
}

// MarshalTo accepts a byte array with pre-allocated space (see Size()) for GenesisTransaction.
// It returns how many bytes it wrote to the array, or 0 and an error.
func (obj *GenesisTransaction) MarshalTo(data []byte) (int, error) {

	var n int

	if len(obj.Outputs) > 20 {
		return 0, errors.Wrap(ranger.ErrTooMany, "GenesisTransaction.Outputs")
	}

	n += binary.PutUvarint(data[n:], uint64(len(obj.Outputs)))
	for _, item := range obj.Outputs {

		if item == nil {
			return 0, errors.Wrap(ranger.ErrShortWrite, "GenesisTransaction.Outputs is nil, cannot serialise")
		}

		if len(data[n:]) < item.Size() {
			return 0, errors.Wrap(ranger.ErrShortWrite, "GenesisTransaction.Outputs")
		}

		{
			ni, err := item.MarshalTo(data[n : n+item.Size()])
			if err != nil {
				return 0, errors.Wrap(err, "GenesisTransaction.Outputs")
			}
			n += ni
		}

	}
	return n, nil
}

// Size returns the computed size of GenesisTransaction as would-be marshaled
// without actually marshaling it.
func (obj *GenesisTransaction) Size() int {
	var n int
	if obj.Outputs == nil {
		// Cannot calculate the value for missing fields
		return 0
	} else {
		n += ranger.UvarintSize(uint64(len(obj.Outputs)))
		for _, item := range obj.Outputs {
			n += item.Size()
		}
	}
	return n
}

// Unmarshal accepts GenesisTransaction's binary representation and transforms the
// GenesisTransaction used as the object. It returns any error.
func (obj *GenesisTransaction) Unmarshal(data []byte) error {
	_, err := obj.UnmarshalFrom(data)
	return err
}

// UnmarshalFrom is very similar to Unmarshal, but also returns the count of data it read.
func (obj *GenesisTransaction) UnmarshalFrom(data []byte) (int, error) {
	var n int

	{
		iLen, ni := binary.Uvarint(data[n:])
		if ni <= 0 {
			return 0, errors.Wrap(ranger.ErrShortRead, "Obtaining length of GenesisTransaction.Outputs")
		}
		n += ni
		if iLen > 20 {
			return 0, errors.Wrap(ranger.ErrTooMany, "GenesisTransaction.Outputs")
		}
		obj.Outputs = make([]*TransactionOutput, iLen)
		for i := uint64(0); i < iLen; i++ {
			obj.Outputs[i] = &TransactionOutput{}

			if len(data[n:]) < 2 {
				return 0, errors.Wrap(ranger.ErrShortRead, "GenesisTransaction.Outputs")
			}
			{
				ni, err := obj.Outputs[i].UnmarshalFrom(data[n:])
				if err != nil {
					return 0, errors.Wrap(err, "GenesisTransaction.Outputs")
				}
				n += ni
			}
		}
	}

	return n, nil
}

type GlobalConfigListInsertion struct {
	Index uint64
	Value []byte
}

// Marshal returns a byte array containing the marshaled representation of GlobalConfigListInsertion, or nil and an error.
func (obj *GlobalConfigListInsertion) Marshal() ([]byte, error) {
	data := make([]byte, obj.Size())
	n, err := obj.MarshalTo(data)
	if err != nil {
		return nil, errors.Wrap(err, "GlobalConfigListInsertion")
	}

	if n != len(data) {
		return nil, errors.Wrapf(ranger.ErrMarshalLength, "%s %d %d", "GlobalConfigListInsertion", n, len(data))
	}

	return data, nil
}

// MarshalTo accepts a byte array with pre-allocated space (see Size()) for GlobalConfigListInsertion.
// It returns how many bytes it wrote to the array, or 0 and an error.
func (obj *GlobalConfigListInsertion) MarshalTo(data []byte) (int, error) {

	var n int

	if len(data[n:]) < ranger.UvarintSize(uint64(obj.Index)) {
		return 0, errors.Wrap(ranger.ErrShortWrite, "GlobalConfigListInsertion.Index")
	}

	n += binary.PutUvarint(data[n:], uint64(obj.Index))

	if len(obj.Value) > 20 {
		return 0, errors.Wrap(ranger.ErrTooMany, "GlobalConfigListInsertion.Value")
	}

	if len(data[n:]) < ranger.UvarintSize(uint64(len(obj.Value)))+len(obj.Value) {
		return 0, errors.Wrap(ranger.ErrShortWrite, "GlobalConfigListInsertion.Value")
	}

	n += binary.PutUvarint(data[n:], uint64(len(obj.Value)))
	n += copy(data[n:n+len(obj.Value)], obj.Value)

	return n, nil
}

// Size returns the computed size of GlobalConfigListInsertion as would-be marshaled
// without actually marshaling it.
func (obj *GlobalConfigListInsertion) Size() int {
	var n int
	n += ranger.UvarintSize(uint64(obj.Index))
	n += ranger.UvarintSize(uint64(len(obj.Value))) + len(obj.Value)
	return n
}

// Unmarshal accepts GlobalConfigListInsertion's binary representation and transforms the
// GlobalConfigListInsertion used as the object. It returns any error.
func (obj *GlobalConfigListInsertion) Unmarshal(data []byte) error {
	_, err := obj.UnmarshalFrom(data)
	return err
}

// UnmarshalFrom is very similar to Unmarshal, but also returns the count of data it read.
func (obj *GlobalConfigListInsertion) UnmarshalFrom(data []byte) (int, error) {
	var n int

	if len(data[n:]) < 1 {
		return 0, errors.Wrap(ranger.ErrShortRead, "GlobalConfigListInsertion.Index")
	}
	{
		iL, ni := binary.Uvarint(data[n:])
		if ni <= 0 {
			return 0, errors.Wrap(ranger.ErrShortRead, "GlobalConfigListInsertion.Index")
		}
		if iL&math.MaxUint64 != iL {
			return 0, errors.Wrap(ranger.ErrTooLarge, "GlobalConfigListInsertion.Index")
		}
		obj.Index = uint64(iL)
		n += ni
	}

	if len(obj.Value) > 20 {
		return 0, errors.Wrap(ranger.ErrTooMany, "GlobalConfigListInsertion.Value")
	}

	if len(data[n:]) < 1 {
		return 0, errors.Wrap(ranger.ErrShortRead, "GlobalConfigListInsertion.Value")
	}
	{
		iL, ni := binary.Uvarint(data[n:])
		if ni <= 0 {
			return 0, errors.Wrap(ranger.ErrShortRead, "Obtaining length of GlobalConfigListInsertion.Value")
		}
		n += ni
		if iL > 20 {
			return 0, errors.Wrap(ranger.ErrTooMany, "GlobalConfigListInsertion.Value")
		}

		if iL > uint64(len(data[n:])) {
			return 0, errors.Wrap(ranger.ErrShortRead, "GlobalConfigListInsertion.Value")
		}
		byt := make([]byte, iL)
		n += copy(byt, data[n:uint64(n)+iL])
		obj.Value = (byt)
	}

	return n, nil
}

type GlobalConfigListUpdate struct {
	Key        string
	Deletions  []uint64
	Insertions []*GlobalConfigListInsertion
}

// Marshal returns a byte array containing the marshaled representation of GlobalConfigListUpdate, or nil and an error.
func (obj *GlobalConfigListUpdate) Marshal() ([]byte, error) {
	data := make([]byte, obj.Size())
	n, err := obj.MarshalTo(data)
	if err != nil {
		return nil, errors.Wrap(err, "GlobalConfigListUpdate")
	}

	if n != len(data) {
		return nil, errors.Wrapf(ranger.ErrMarshalLength, "%s %d %d", "GlobalConfigListUpdate", n, len(data))
	}

	return data, nil
}

// MarshalTo accepts a byte array with pre-allocated space (see Size()) for GlobalConfigListUpdate.
// It returns how many bytes it wrote to the array, or 0 and an error.
func (obj *GlobalConfigListUpdate) MarshalTo(data []byte) (int, error) {

	var n int

	if len(obj.Key) > 20 {
		return 0, errors.Wrap(ranger.ErrTooMany, "GlobalConfigListUpdate.Key")
	}

	if len(data[n:]) < ranger.UvarintSize(uint64(len(obj.Key)))+len(obj.Key) {
		return 0, errors.Wrap(ranger.ErrShortWrite, "GlobalConfigListUpdate.Key")
	}

	n += binary.PutUvarint(data[n:], uint64(len(obj.Key)))
	n += copy(data[n:n+len(obj.Key)], obj.Key)

	if len(obj.Deletions) > 20 {
		return 0, errors.Wrap(ranger.ErrTooMany, "GlobalConfigListUpdate.Deletions")
	}

	n += binary.PutUvarint(data[n:], uint64(len(obj.Deletions)))
	for _, item := range obj.Deletions {

		if len(data[n:]) < ranger.UvarintSize(uint64(item)) {
			return 0, errors.Wrap(ranger.ErrShortWrite, "GlobalConfigListUpdate.Deletions")
		}

		n += binary.PutUvarint(data[n:], uint64(item))
	}
	if len(obj.Insertions) > 20 {
		return 0, errors.Wrap(ranger.ErrTooMany, "GlobalConfigListUpdate.Insertions")
	}

	n += binary.PutUvarint(data[n:], uint64(len(obj.Insertions)))
	for _, item := range obj.Insertions {

		if item == nil {
			return 0, errors.Wrap(ranger.ErrShortWrite, "GlobalConfigListUpdate.Insertions is nil, cannot serialise")
		}

		if len(data[n:]) < item.Size() {
			return 0, errors.Wrap(ranger.ErrShortWrite, "GlobalConfigListUpdate.Insertions")
		}

		{
			ni, err := item.MarshalTo(data[n : n+item.Size()])
			if err != nil {
				return 0, errors.Wrap(err, "GlobalConfigListUpdate.Insertions")
			}
			n += ni
		}

	}
	return n, nil
}

// Size returns the computed size of GlobalConfigListUpdate as would-be marshaled
// without actually marshaling it.
func (obj *GlobalConfigListUpdate) Size() int {
	var n int
	n += ranger.UvarintSize(uint64(len(obj.Key))) + len(obj.Key)
	if obj.Deletions == nil {
		// Cannot calculate the value for missing fields
		return 0
	} else {
		n += ranger.UvarintSize(uint64(len(obj.Deletions)))
		for _, item := range obj.Deletions {
			n += ranger.UvarintSize(uint64(item))
		}
	}
	if obj.Insertions == nil {
		// Cannot calculate the value for missing fields
		return 0
	} else {
		n += ranger.UvarintSize(uint64(len(obj.Insertions)))
		for _, item := range obj.Insertions {
			n += item.Size()
		}
	}
	return n
}

// Unmarshal accepts GlobalConfigListUpdate's binary representation and transforms the
// GlobalConfigListUpdate used as the object. It returns any error.
func (obj *GlobalConfigListUpdate) Unmarshal(data []byte) error {
	_, err := obj.UnmarshalFrom(data)
	return err
}

// UnmarshalFrom is very similar to Unmarshal, but also returns the count of data it read.
func (obj *GlobalConfigListUpdate) UnmarshalFrom(data []byte) (int, error) {
	var n int

	if len(obj.Key) > 20 {
		return 0, errors.Wrap(ranger.ErrTooMany, "GlobalConfigListUpdate.Key")
	}

	if len(data[n:]) < 1 {
		return 0, errors.Wrap(ranger.ErrShortRead, "GlobalConfigListUpdate.Key")
	}
	{
		iL, ni := binary.Uvarint(data[n:])
		if ni <= 0 {
			return 0, errors.Wrap(ranger.ErrShortRead, "Obtaining length of GlobalConfigListUpdate.Key")
		}
		n += ni
		if iL > 20 {
			return 0, errors.Wrap(ranger.ErrTooMany, "GlobalConfigListUpdate.Key")
		}

		if iL > uint64(len(data[n:])) {
			return 0, errors.Wrap(ranger.ErrShortRead, "GlobalConfigListUpdate.Key")
		}
		byt := make([]byte, iL)
		n += copy(byt, data[n:uint64(n)+iL])
		obj.Key = string(byt)
	}

	{
		iLen, ni := binary.Uvarint(data[n:])
		if ni <= 0 {
			return 0, errors.Wrap(ranger.ErrShortRead, "Obtaining length of GlobalConfigListUpdate.Deletions")
		}
		n += ni
		if iLen > 20 {
			return 0, errors.Wrap(ranger.ErrTooMany, "GlobalConfigListUpdate.Deletions")
		}
		obj.Deletions = make([]uint64, iLen)
		for i := uint64(0); i < iLen; i++ {

			if len(data[n:]) < 1 {
				return 0, errors.Wrap(ranger.ErrShortRead, "GlobalConfigListUpdate.Deletions")
			}
			{
				iL, ni := binary.Uvarint(data[n:])
				if ni <= 0 {
					return 0, errors.Wrap(ranger.ErrShortRead, "GlobalConfigListUpdate.Deletions")
				}
				if iL&math.MaxUint64 != iL {
					return 0, errors.Wrap(ranger.ErrTooLarge, "GlobalConfigListUpdate.Deletions")
				}
				obj.Deletions[i] = uint64(iL)
				n += ni
			}
		}
	}

	{
		iLen, ni := binary.Uvarint(data[n:])
		if ni <= 0 {
			return 0, errors.Wrap(ranger.ErrShortRead, "Obtaining length of GlobalConfigListUpdate.Insertions")
		}
		n += ni
		if iLen > 20 {
			return 0, errors.Wrap(ranger.ErrTooMany, "GlobalConfigListUpdate.Insertions")
		}
		obj.Insertions = make([]*GlobalConfigListInsertion, iLen)
		for i := uint64(0); i < iLen; i++ {
			obj.Insertions[i] = &GlobalConfigListInsertion{}

			if len(data[n:]) < 2 {
				return 0, errors.Wrap(ranger.ErrShortRead, "GlobalConfigListUpdate.Insertions")
			}
			{
				ni, err := obj.Insertions[i].UnmarshalFrom(data[n:])
				if err != nil {
					return 0, errors.Wrap(err, "GlobalConfigListUpdate.Insertions")
				}
				n += ni
			}
		}
	}

	return n, nil
}

type GlobalConfigScalarUpdate struct {
	Key   string
	Value []byte
}

// Marshal returns a byte array containing the marshaled representation of GlobalConfigScalarUpdate, or nil and an error.
func (obj *GlobalConfigScalarUpdate) Marshal() ([]byte, error) {
	data := make([]byte, obj.Size())
	n, err := obj.MarshalTo(data)
	if err != nil {
		return nil, errors.Wrap(err, "GlobalConfigScalarUpdate")
	}

	if n != len(data) {
		return nil, errors.Wrapf(ranger.ErrMarshalLength, "%s %d %d", "GlobalConfigScalarUpdate", n, len(data))
	}

	return data, nil
}

// MarshalTo accepts a byte array with pre-allocated space (see Size()) for GlobalConfigScalarUpdate.
// It returns how many bytes it wrote to the array, or 0 and an error.
func (obj *GlobalConfigScalarUpdate) MarshalTo(data []byte) (int, error) {

	var n int

	if len(obj.Key) > 20 {
		return 0, errors.Wrap(ranger.ErrTooMany, "GlobalConfigScalarUpdate.Key")
	}

	if len(data[n:]) < ranger.UvarintSize(uint64(len(obj.Key)))+len(obj.Key) {
		return 0, errors.Wrap(ranger.ErrShortWrite, "GlobalConfigScalarUpdate.Key")
	}

	n += binary.PutUvarint(data[n:], uint64(len(obj.Key)))
	n += copy(data[n:n+len(obj.Key)], obj.Key)

	if len(obj.Value) > 20 {
		return 0, errors.Wrap(ranger.ErrTooMany, "GlobalConfigScalarUpdate.Value")
	}

	if len(data[n:]) < ranger.UvarintSize(uint64(len(obj.Value)))+len(obj.Value) {
		return 0, errors.Wrap(ranger.ErrShortWrite, "GlobalConfigScalarUpdate.Value")
	}

	n += binary.PutUvarint(data[n:], uint64(len(obj.Value)))
	n += copy(data[n:n+len(obj.Value)], obj.Value)

	return n, nil
}

// Size returns the computed size of GlobalConfigScalarUpdate as would-be marshaled
// without actually marshaling it.
func (obj *GlobalConfigScalarUpdate) Size() int {
	var n int
	n += ranger.UvarintSize(uint64(len(obj.Key))) + len(obj.Key)
	n += ranger.UvarintSize(uint64(len(obj.Value))) + len(obj.Value)
	return n
}

// Unmarshal accepts GlobalConfigScalarUpdate's binary representation and transforms the
// GlobalConfigScalarUpdate used as the object. It returns any error.
func (obj *GlobalConfigScalarUpdate) Unmarshal(data []byte) error {
	_, err := obj.UnmarshalFrom(data)
	return err
}

// UnmarshalFrom is very similar to Unmarshal, but also returns the count of data it read.
func (obj *GlobalConfigScalarUpdate) UnmarshalFrom(data []byte) (int, error) {
	var n int

	if len(obj.Key) > 20 {
		return 0, errors.Wrap(ranger.ErrTooMany, "GlobalConfigScalarUpdate.Key")
	}

	if len(data[n:]) < 1 {
		return 0, errors.Wrap(ranger.ErrShortRead, "GlobalConfigScalarUpdate.Key")
	}
	{
		iL, ni := binary.Uvarint(data[n:])
		if ni <= 0 {
			return 0, errors.Wrap(ranger.ErrShortRead, "Obtaining length of GlobalConfigScalarUpdate.Key")
		}
		n += ni
		if iL > 20 {
			return 0, errors.Wrap(ranger.ErrTooMany, "GlobalConfigScalarUpdate.Key")
		}

		if iL > uint64(len(data[n:])) {
			return 0, errors.Wrap(ranger.ErrShortRead, "GlobalConfigScalarUpdate.Key")
		}
		byt := make([]byte, iL)
		n += copy(byt, data[n:uint64(n)+iL])
		obj.Key = string(byt)
	}

	if len(obj.Value) > 20 {
		return 0, errors.Wrap(ranger.ErrTooMany, "GlobalConfigScalarUpdate.Value")
	}

	if len(data[n:]) < 1 {
		return 0, errors.Wrap(ranger.ErrShortRead, "GlobalConfigScalarUpdate.Value")
	}
	{
		iL, ni := binary.Uvarint(data[n:])
		if ni <= 0 {
			return 0, errors.Wrap(ranger.ErrShortRead, "Obtaining length of GlobalConfigScalarUpdate.Value")
		}
		n += ni
		if iL > 20 {
			return 0, errors.Wrap(ranger.ErrTooMany, "GlobalConfigScalarUpdate.Value")
		}

		if iL > uint64(len(data[n:])) {
			return 0, errors.Wrap(ranger.ErrShortRead, "GlobalConfigScalarUpdate.Value")
		}
		byt := make([]byte, iL)
		n += copy(byt, data[n:uint64(n)+iL])
		obj.Value = (byt)
	}

	return n, nil
}

type GlobalConfigTransaction struct {
	ActivationBlockHeight uint64
	ScalarUpdates         []*GlobalConfigScalarUpdate
	ListUpdates           []*GlobalConfigListUpdate
	SigPublicKey          []byte
	Signature             []byte
}

// Marshal returns a byte array containing the marshaled representation of GlobalConfigTransaction, or nil and an error.
func (obj *GlobalConfigTransaction) Marshal() ([]byte, error) {
	data := make([]byte, obj.Size())
	n, err := obj.MarshalTo(data)
	if err != nil {
		return nil, errors.Wrap(err, "GlobalConfigTransaction")
	}

	if n != len(data) {
		return nil, errors.Wrapf(ranger.ErrMarshalLength, "%s %d %d", "GlobalConfigTransaction", n, len(data))
	}

	return data, nil
}

// MarshalTo accepts a byte array with pre-allocated space (see Size()) for GlobalConfigTransaction.
// It returns how many bytes it wrote to the array, or 0 and an error.
func (obj *GlobalConfigTransaction) MarshalTo(data []byte) (int, error) {

	var n int

	if len(data[n:]) < ranger.UvarintSize(uint64(obj.ActivationBlockHeight)) {
		return 0, errors.Wrap(ranger.ErrShortWrite, "GlobalConfigTransaction.ActivationBlockHeight")
	}

	n += binary.PutUvarint(data[n:], uint64(obj.ActivationBlockHeight))

	if len(obj.ScalarUpdates) > 20 {
		return 0, errors.Wrap(ranger.ErrTooMany, "GlobalConfigTransaction.ScalarUpdates")
	}

	n += binary.PutUvarint(data[n:], uint64(len(obj.ScalarUpdates)))
	for _, item := range obj.ScalarUpdates {

		if item == nil {
			return 0, errors.Wrap(ranger.ErrShortWrite, "GlobalConfigTransaction.ScalarUpdates is nil, cannot serialise")
		}

		if len(data[n:]) < item.Size() {
			return 0, errors.Wrap(ranger.ErrShortWrite, "GlobalConfigTransaction.ScalarUpdates")
		}

		{
			ni, err := item.MarshalTo(data[n : n+item.Size()])
			if err != nil {
				return 0, errors.Wrap(err, "GlobalConfigTransaction.ScalarUpdates")
			}
			n += ni
		}

	}
	if len(obj.ListUpdates) > 20 {
		return 0, errors.Wrap(ranger.ErrTooMany, "GlobalConfigTransaction.ListUpdates")
	}

	n += binary.PutUvarint(data[n:], uint64(len(obj.ListUpdates)))
	for _, item := range obj.ListUpdates {

		if item == nil {
			return 0, errors.Wrap(ranger.ErrShortWrite, "GlobalConfigTransaction.ListUpdates is nil, cannot serialise")
		}

		if len(data[n:]) < item.Size() {
			return 0, errors.Wrap(ranger.ErrShortWrite, "GlobalConfigTransaction.ListUpdates")
		}

		{
			ni, err := item.MarshalTo(data[n : n+item.Size()])
			if err != nil {
				return 0, errors.Wrap(err, "GlobalConfigTransaction.ListUpdates")
			}
			n += ni
		}

	}
	if len(obj.SigPublicKey) > 20 {
		return 0, errors.Wrap(ranger.ErrTooMany, "GlobalConfigTransaction.SigPublicKey")
	}

	if len(data[n:]) < ranger.UvarintSize(uint64(len(obj.SigPublicKey)))+len(obj.SigPublicKey) {
		return 0, errors.Wrap(ranger.ErrShortWrite, "GlobalConfigTransaction.SigPublicKey")
	}

	n += binary.PutUvarint(data[n:], uint64(len(obj.SigPublicKey)))
	n += copy(data[n:n+len(obj.SigPublicKey)], obj.SigPublicKey)

	if len(obj.Signature) > 20 {
		return 0, errors.Wrap(ranger.ErrTooMany, "GlobalConfigTransaction.Signature")
	}

	if len(data[n:]) < ranger.UvarintSize(uint64(len(obj.Signature)))+len(obj.Signature) {
		return 0, errors.Wrap(ranger.ErrShortWrite, "GlobalConfigTransaction.Signature")
	}

	n += binary.PutUvarint(data[n:], uint64(len(obj.Signature)))
	n += copy(data[n:n+len(obj.Signature)], obj.Signature)

	return n, nil
}

// Size returns the computed size of GlobalConfigTransaction as would-be marshaled
// without actually marshaling it.
func (obj *GlobalConfigTransaction) Size() int {
	var n int
	n += ranger.UvarintSize(uint64(obj.ActivationBlockHeight))
	if obj.ScalarUpdates == nil {
		// Cannot calculate the value for missing fields
		return 0
	} else {
		n += ranger.UvarintSize(uint64(len(obj.ScalarUpdates)))
		for _, item := range obj.ScalarUpdates {
			n += item.Size()
		}
	}
	if obj.ListUpdates == nil {
		// Cannot calculate the value for missing fields
		return 0
	} else {
		n += ranger.UvarintSize(uint64(len(obj.ListUpdates)))
		for _, item := range obj.ListUpdates {
			n += item.Size()
		}
	}
	n += ranger.UvarintSize(uint64(len(obj.SigPublicKey))) + len(obj.SigPublicKey)
	n += ranger.UvarintSize(uint64(len(obj.Signature))) + len(obj.Signature)
	return n
}

// Unmarshal accepts GlobalConfigTransaction's binary representation and transforms the
// GlobalConfigTransaction used as the object. It returns any error.
func (obj *GlobalConfigTransaction) Unmarshal(data []byte) error {
	_, err := obj.UnmarshalFrom(data)
	return err
}

// UnmarshalFrom is very similar to Unmarshal, but also returns the count of data it read.
func (obj *GlobalConfigTransaction) UnmarshalFrom(data []byte) (int, error) {
	var n int

	if len(data[n:]) < 1 {
		return 0, errors.Wrap(ranger.ErrShortRead, "GlobalConfigTransaction.ActivationBlockHeight")
	}
	{
		iL, ni := binary.Uvarint(data[n:])
		if ni <= 0 {
			return 0, errors.Wrap(ranger.ErrShortRead, "GlobalConfigTransaction.ActivationBlockHeight")
		}
		if iL&math.MaxUint64 != iL {
			return 0, errors.Wrap(ranger.ErrTooLarge, "GlobalConfigTransaction.ActivationBlockHeight")
		}
		obj.ActivationBlockHeight = uint64(iL)
		n += ni
	}

	{
		iLen, ni := binary.Uvarint(data[n:])
		if ni <= 0 {
			return 0, errors.Wrap(ranger.ErrShortRead, "Obtaining length of GlobalConfigTransaction.ScalarUpdates")
		}
		n += ni
		if iLen > 20 {
			return 0, errors.Wrap(ranger.ErrTooMany, "GlobalConfigTransaction.ScalarUpdates")
		}
		obj.ScalarUpdates = make([]*GlobalConfigScalarUpdate, iLen)
		for i := uint64(0); i < iLen; i++ {
			obj.ScalarUpdates[i] = &GlobalConfigScalarUpdate{}

			if len(data[n:]) < 2 {
				return 0, errors.Wrap(ranger.ErrShortRead, "GlobalConfigTransaction.ScalarUpdates")
			}
			{
				ni, err := obj.ScalarUpdates[i].UnmarshalFrom(data[n:])
				if err != nil {
					return 0, errors.Wrap(err, "GlobalConfigTransaction.ScalarUpdates")
				}
				n += ni
			}
		}
	}

	{
		iLen, ni := binary.Uvarint(data[n:])
		if ni <= 0 {
			return 0, errors.Wrap(ranger.ErrShortRead, "Obtaining length of GlobalConfigTransaction.ListUpdates")
		}
		n += ni
		if iLen > 20 {
			return 0, errors.Wrap(ranger.ErrTooMany, "GlobalConfigTransaction.ListUpdates")
		}
		obj.ListUpdates = make([]*GlobalConfigListUpdate, iLen)
		for i := uint64(0); i < iLen; i++ {
			obj.ListUpdates[i] = &GlobalConfigListUpdate{}

			if len(data[n:]) < 3 {
				return 0, errors.Wrap(ranger.ErrShortRead, "GlobalConfigTransaction.ListUpdates")
			}
			{
				ni, err := obj.ListUpdates[i].UnmarshalFrom(data[n:])
				if err != nil {
					return 0, errors.Wrap(err, "GlobalConfigTransaction.ListUpdates")
				}
				n += ni
			}
		}
	}

	if len(obj.SigPublicKey) > 20 {
		return 0, errors.Wrap(ranger.ErrTooMany, "GlobalConfigTransaction.SigPublicKey")
	}

	if len(data[n:]) < 1 {
		return 0, errors.Wrap(ranger.ErrShortRead, "GlobalConfigTransaction.SigPublicKey")
	}
	{
		iL, ni := binary.Uvarint(data[n:])
		if ni <= 0 {
			return 0, errors.Wrap(ranger.ErrShortRead, "Obtaining length of GlobalConfigTransaction.SigPublicKey")
		}
		n += ni
		if iL > 20 {
			return 0, errors.Wrap(ranger.ErrTooMany, "GlobalConfigTransaction.SigPublicKey")
		}

		if iL > uint64(len(data[n:])) {
			return 0, errors.Wrap(ranger.ErrShortRead, "GlobalConfigTransaction.SigPublicKey")
		}
		byt := make([]byte, iL)
		n += copy(byt, data[n:uint64(n)+iL])
		obj.SigPublicKey = (byt)
	}

	if len(obj.Signature) > 20 {
		return 0, errors.Wrap(ranger.ErrTooMany, "GlobalConfigTransaction.Signature")
	}

	if len(data[n:]) < 1 {
		return 0, errors.Wrap(ranger.ErrShortRead, "GlobalConfigTransaction.Signature")
	}
	{
		iL, ni := binary.Uvarint(data[n:])
		if ni <= 0 {
			return 0, errors.Wrap(ranger.ErrShortRead, "Obtaining length of GlobalConfigTransaction.Signature")
		}
		n += ni
		if iL > 20 {
			return 0, errors.Wrap(ranger.ErrTooMany, "GlobalConfigTransaction.Signature")
		}

		if iL > uint64(len(data[n:])) {
			return 0, errors.Wrap(ranger.ErrShortRead, "GlobalConfigTransaction.Signature")
		}
		byt := make([]byte, iL)
		n += copy(byt, data[n:uint64(n)+iL])
		obj.Signature = (byt)
	}

	return n, nil
}

type Outpoint struct {
	PreviousTx [32]byte
	Index      uint8
}

// Marshal returns a byte array containing the marshaled representation of Outpoint, or nil and an error.
func (obj *Outpoint) Marshal() ([]byte, error) {
	data := make([]byte, obj.Size())
	n, err := obj.MarshalTo(data)
	if err != nil {
		return nil, errors.Wrap(err, "Outpoint")
	}

	if n != len(data) {
		return nil, errors.Wrapf(ranger.ErrMarshalLength, "%s %d %d", "Outpoint", n, len(data))
	}

	return data, nil
}

// MarshalTo accepts a byte array with pre-allocated space (see Size()) for Outpoint.
// It returns how many bytes it wrote to the array, or 0 and an error.
func (obj *Outpoint) MarshalTo(data []byte) (int, error) {

	var n int

	if len(data[n:]) < int(32) {
		return 0, errors.Wrap(ranger.ErrShortWrite, "Outpoint.PreviousTx")
	}

	n += copy(data[n:], obj.PreviousTx[:])

	if len(data[n:]) < 1 {
		return 0, errors.Wrap(ranger.ErrShortWrite, "Outpoint.Index")
	}

	data[n] = obj.Index
	n += 1

	return n, nil
}

// Size returns the computed size of Outpoint as would-be marshaled
// without actually marshaling it.
func (obj *Outpoint) Size() int {
	var n int
	n += int(32)
	n += 1
	return n
}

// Unmarshal accepts Outpoint's binary representation and transforms the
// Outpoint used as the object. It returns any error.
func (obj *Outpoint) Unmarshal(data []byte) error {
	_, err := obj.UnmarshalFrom(data)
	return err
}

// UnmarshalFrom is very similar to Unmarshal, but also returns the count of data it read.
func (obj *Outpoint) UnmarshalFrom(data []byte) (int, error) {
	var n int

	if len(data[n:]) < 1 {
		return 0, errors.Wrap(ranger.ErrShortRead, "Outpoint.PreviousTx")
	}
	{
		iL := uint64(32)

		if iL > uint64(len(data[n:])) {
			return 0, errors.Wrap(ranger.ErrShortRead, "Outpoint.PreviousTx")
		}
		n += copy(obj.PreviousTx[:], data[n:uint64(n)+iL])
	}

	if len(data[n:]) < 1 {
		return 0, errors.Wrap(ranger.ErrShortRead, "Outpoint.Index")
	}
	obj.Index = data[n]
	n += 1

	return n, nil
}

type Transaction struct {
	Version uint8
	Body    TransactionBody
	Flags   uint16
}

// Marshal returns a byte array containing the marshaled representation of Transaction, or nil and an error.
func (obj *Transaction) Marshal() ([]byte, error) {
	data := make([]byte, obj.Size())
	n, err := obj.MarshalTo(data)
	if err != nil {
		return nil, errors.Wrap(err, "Transaction")
	}

	if n != len(data) {
		return nil, errors.Wrapf(ranger.ErrMarshalLength, "%s %d %d", "Transaction", n, len(data))
	}

	return data, nil
}

// MarshalTo accepts a byte array with pre-allocated space (see Size()) for Transaction.
// It returns how many bytes it wrote to the array, or 0 and an error.
func (obj *Transaction) MarshalTo(data []byte) (int, error) {

	var n int

	if len(data[n:]) < 1 {
		return 0, errors.Wrap(ranger.ErrShortWrite, "Transaction.Version")
	}

	data[n] = obj.Version
	n += 1

	if obj.Body == nil {
		return 0, errors.Wrap(ranger.ErrShortWrite, "Transaction.Body is nil, cannot serialise")
	}

	if len(data[n:]) < 1+obj.Body.Size() {
		return 0, errors.Wrap(ranger.ErrShortWrite, "Transaction.Body")
	}

	data[n] = obj.Body.TxType()
	n += 1
	{
		ni, err := obj.Body.MarshalTo(data[n : n+obj.Body.Size()])
		if err != nil {
			return 0, errors.Wrap(err, "Transaction.Body")
		}
		n += ni
	}

	if len(data[n:]) < 2 {
		return 0, errors.Wrap(ranger.ErrShortWrite, "Transaction.Flags")
	}

	binary.LittleEndian.PutUint16(data[n:], obj.Flags)
	n += 2

	return n, nil
}

// Size returns the computed size of Transaction as would-be marshaled
// without actually marshaling it.
func (obj *Transaction) Size() int {
	var n int
	n += 1
	n += 1 + obj.Body.Size()
	n += 2
	return n
}

// Unmarshal accepts Transaction's binary representation and transforms the
// Transaction used as the object. It returns any error.
func (obj *Transaction) Unmarshal(data []byte) error {
	_, err := obj.UnmarshalFrom(data)
	return err
}

// UnmarshalFrom is very similar to Unmarshal, but also returns the count of data it read.
func (obj *Transaction) UnmarshalFrom(data []byte) (int, error) {
	var n int

	if len(data[n:]) < 1 {
		return 0, errors.Wrap(ranger.ErrShortRead, "Transaction.Version")
	}
	obj.Version = data[n]
	n += 1

	if len(data[n:]) < 1 {
		return 0, errors.Wrap(ranger.ErrShortRead, "Transaction.Body")
	}
	{
		var intf uint8
		intf = data[n]
		n += 1

		var v TransactionBody

		switch intf {
		case TxTypeTransfer:
			v = &TransferTransaction{}
		case TxTypeGenesis:
			v = &GenesisTransaction{}
		case TxTypeGlobalConfig:
			v = &GlobalConfigTransaction{}
		case TxTypeEscrowOpen:
			v = &EscrowOpenTransaction{}
		default:
			return 0, errors.Wrap(ranger.ErrBadInterface, "Transaction.Body")
		}

		obj.Body = v
	}
	{
		ni, err := obj.Body.UnmarshalFrom(data[n:])
		if err != nil {
			return 0, errors.Wrap(err, "Transaction.Body")
		}
		n += ni
	}

	if len(data[n:]) < 2 {
		return 0, errors.Wrap(ranger.ErrShortRead, "Transaction.Flags")
	}
	obj.Flags = binary.LittleEndian.Uint16(data[n:])
	n += 2

	return n, nil
}

type TransactionInput struct {
	Outpoint
	ScriptSig  []byte
	SequenceNo uint32
}

// Marshal returns a byte array containing the marshaled representation of TransactionInput, or nil and an error.
func (obj *TransactionInput) Marshal() ([]byte, error) {
	data := make([]byte, obj.Size())
	n, err := obj.MarshalTo(data)
	if err != nil {
		return nil, errors.Wrap(err, "TransactionInput")
	}

	if n != len(data) {
		return nil, errors.Wrapf(ranger.ErrMarshalLength, "%s %d %d", "TransactionInput", n, len(data))
	}

	return data, nil
}

// MarshalTo accepts a byte array with pre-allocated space (see Size()) for TransactionInput.
// It returns how many bytes it wrote to the array, or 0 and an error.
func (obj *TransactionInput) MarshalTo(data []byte) (int, error) {

	var n int

	if len(data[n:]) < obj.Outpoint.Size() {
		return 0, errors.Wrap(ranger.ErrShortWrite, "TransactionInput.Outpoint")
	}

	{
		ni, err := obj.Outpoint.MarshalTo(data[n : n+obj.Outpoint.Size()])
		if err != nil {
			return 0, errors.Wrap(err, "TransactionInput.Outpoint")
		}
		n += ni
	}

	if len(obj.ScriptSig) > 520 {
		return 0, errors.Wrap(ranger.ErrTooMany, "TransactionInput.ScriptSig")
	}

	if len(data[n:]) < ranger.UvarintSize(uint64(len(obj.ScriptSig)))+len(obj.ScriptSig) {
		return 0, errors.Wrap(ranger.ErrShortWrite, "TransactionInput.ScriptSig")
	}

	n += binary.PutUvarint(data[n:], uint64(len(obj.ScriptSig)))
	n += copy(data[n:n+len(obj.ScriptSig)], obj.ScriptSig)

	if len(data[n:]) < ranger.UvarintSize(uint64(obj.SequenceNo)) {
		return 0, errors.Wrap(ranger.ErrShortWrite, "TransactionInput.SequenceNo")
	}

	n += binary.PutUvarint(data[n:], uint64(obj.SequenceNo))

	return n, nil
}

// Size returns the computed size of TransactionInput as would-be marshaled
// without actually marshaling it.
func (obj *TransactionInput) Size() int {
	var n int
	n += obj.Outpoint.Size()
	n += ranger.UvarintSize(uint64(len(obj.ScriptSig))) + len(obj.ScriptSig)
	n += ranger.UvarintSize(uint64(obj.SequenceNo))
	return n
}

// Unmarshal accepts TransactionInput's binary representation and transforms the
// TransactionInput used as the object. It returns any error.
func (obj *TransactionInput) Unmarshal(data []byte) error {
	_, err := obj.UnmarshalFrom(data)
	return err
}

// UnmarshalFrom is very similar to Unmarshal, but also returns the count of data it read.
func (obj *TransactionInput) UnmarshalFrom(data []byte) (int, error) {
	var n int

	{
		ni, err := obj.Outpoint.UnmarshalFrom(data[n:])
		if err != nil {
			return 0, errors.Wrap(err, "TransactionInput.Outpoint")
		}
		n += ni
	}

	if len(obj.ScriptSig) > 520 {
		return 0, errors.Wrap(ranger.ErrTooMany, "TransactionInput.ScriptSig")
	}

	if len(data[n:]) < 1 {
		return 0, errors.Wrap(ranger.ErrShortRead, "TransactionInput.ScriptSig")
	}
	{
		iL, ni := binary.Uvarint(data[n:])
		if ni <= 0 {
			return 0, errors.Wrap(ranger.ErrShortRead, "Obtaining length of TransactionInput.ScriptSig")
		}
		n += ni
		if iL > 520 {
			return 0, errors.Wrap(ranger.ErrTooMany, "TransactionInput.ScriptSig")
		}

		if iL > uint64(len(data[n:])) {
			return 0, errors.Wrap(ranger.ErrShortRead, "TransactionInput.ScriptSig")
		}
		byt := make([]byte, iL)
		n += copy(byt, data[n:uint64(n)+iL])
		obj.ScriptSig = (byt)
	}

	if len(data[n:]) < 1 {
		return 0, errors.Wrap(ranger.ErrShortRead, "TransactionInput.SequenceNo")
	}
	{
		iL, ni := binary.Uvarint(data[n:])
		if ni <= 0 {
			return 0, errors.Wrap(ranger.ErrShortRead, "TransactionInput.SequenceNo")
		}
		if iL&math.MaxUint32 != iL {
			return 0, errors.Wrap(ranger.ErrTooLarge, "TransactionInput.SequenceNo")
		}
		obj.SequenceNo = uint32(iL)
		n += ni
	}

	return n, nil
}

type TransactionOutput struct {
	Value        uint32
	ScriptPubKey []byte
}

// Marshal returns a byte array containing the marshaled representation of TransactionOutput, or nil and an error.
func (obj *TransactionOutput) Marshal() ([]byte, error) {
	data := make([]byte, obj.Size())
	n, err := obj.MarshalTo(data)
	if err != nil {
		return nil, errors.Wrap(err, "TransactionOutput")
	}

	if n != len(data) {
		return nil, errors.Wrapf(ranger.ErrMarshalLength, "%s %d %d", "TransactionOutput", n, len(data))
	}

	return data, nil
}

// MarshalTo accepts a byte array with pre-allocated space (see Size()) for TransactionOutput.
// It returns how many bytes it wrote to the array, or 0 and an error.
func (obj *TransactionOutput) MarshalTo(data []byte) (int, error) {

	var n int

	if len(data[n:]) < ranger.UvarintSize(uint64(obj.Value)) {
		return 0, errors.Wrap(ranger.ErrShortWrite, "TransactionOutput.Value")
	}

	n += binary.PutUvarint(data[n:], uint64(obj.Value))

	if len(obj.ScriptPubKey) > 20 {
		return 0, errors.Wrap(ranger.ErrTooMany, "TransactionOutput.ScriptPubKey")
	}

	if len(data[n:]) < ranger.UvarintSize(uint64(len(obj.ScriptPubKey)))+len(obj.ScriptPubKey) {
		return 0, errors.Wrap(ranger.ErrShortWrite, "TransactionOutput.ScriptPubKey")
	}

	n += binary.PutUvarint(data[n:], uint64(len(obj.ScriptPubKey)))
	n += copy(data[n:n+len(obj.ScriptPubKey)], obj.ScriptPubKey)

	return n, nil
}

// Size returns the computed size of TransactionOutput as would-be marshaled
// without actually marshaling it.
func (obj *TransactionOutput) Size() int {
	var n int
	n += ranger.UvarintSize(uint64(obj.Value))
	n += ranger.UvarintSize(uint64(len(obj.ScriptPubKey))) + len(obj.ScriptPubKey)
	return n
}

// Unmarshal accepts TransactionOutput's binary representation and transforms the
// TransactionOutput used as the object. It returns any error.
func (obj *TransactionOutput) Unmarshal(data []byte) error {
	_, err := obj.UnmarshalFrom(data)
	return err
}

// UnmarshalFrom is very similar to Unmarshal, but also returns the count of data it read.
func (obj *TransactionOutput) UnmarshalFrom(data []byte) (int, error) {
	var n int

	if len(data[n:]) < 1 {
		return 0, errors.Wrap(ranger.ErrShortRead, "TransactionOutput.Value")
	}
	{
		iL, ni := binary.Uvarint(data[n:])
		if ni <= 0 {
			return 0, errors.Wrap(ranger.ErrShortRead, "TransactionOutput.Value")
		}
		if iL&math.MaxUint32 != iL {
			return 0, errors.Wrap(ranger.ErrTooLarge, "TransactionOutput.Value")
		}
		obj.Value = uint32(iL)
		n += ni
	}

	if len(obj.ScriptPubKey) > 20 {
		return 0, errors.Wrap(ranger.ErrTooMany, "TransactionOutput.ScriptPubKey")
	}

	if len(data[n:]) < 1 {
		return 0, errors.Wrap(ranger.ErrShortRead, "TransactionOutput.ScriptPubKey")
	}
	{
		iL, ni := binary.Uvarint(data[n:])
		if ni <= 0 {
			return 0, errors.Wrap(ranger.ErrShortRead, "Obtaining length of TransactionOutput.ScriptPubKey")
		}
		n += ni
		if iL > 20 {
			return 0, errors.Wrap(ranger.ErrTooMany, "TransactionOutput.ScriptPubKey")
		}

		if iL > uint64(len(data[n:])) {
			return 0, errors.Wrap(ranger.ErrShortRead, "TransactionOutput.ScriptPubKey")
		}
		byt := make([]byte, iL)
		n += copy(byt, data[n:uint64(n)+iL])
		obj.ScriptPubKey = (byt)
	}

	return n, nil
}

type TransactionWitness struct {
	Data [][32]byte
}

// Marshal returns a byte array containing the marshaled representation of TransactionWitness, or nil and an error.
func (obj *TransactionWitness) Marshal() ([]byte, error) {
	data := make([]byte, obj.Size())
	n, err := obj.MarshalTo(data)
	if err != nil {
		return nil, errors.Wrap(err, "TransactionWitness")
	}

	if n != len(data) {
		return nil, errors.Wrapf(ranger.ErrMarshalLength, "%s %d %d", "TransactionWitness", n, len(data))
	}

	return data, nil
}

// MarshalTo accepts a byte array with pre-allocated space (see Size()) for TransactionWitness.
// It returns how many bytes it wrote to the array, or 0 and an error.
func (obj *TransactionWitness) MarshalTo(data []byte) (int, error) {

	var n int

	if len(obj.Data) > 20 {
		return 0, errors.Wrap(ranger.ErrTooMany, "TransactionWitness.Data")
	}

	n += binary.PutUvarint(data[n:], uint64(len(obj.Data)))
	for _, item := range obj.Data {

		if len(data[n:]) < int(32) {
			return 0, errors.Wrap(ranger.ErrShortWrite, "TransactionWitness.Data")
		}

		n += copy(data[n:], item[:])
	}
	return n, nil
}

// Size returns the computed size of TransactionWitness as would-be marshaled
// without actually marshaling it.
func (obj *TransactionWitness) Size() int {
	var n int
	if obj.Data == nil {
		// Cannot calculate the value for missing fields
		return 0
	} else {
		n += ranger.UvarintSize(uint64(len(obj.Data)))

		n += len(obj.Data) * int(32)
	}
	return n
}

// Unmarshal accepts TransactionWitness's binary representation and transforms the
// TransactionWitness used as the object. It returns any error.
func (obj *TransactionWitness) Unmarshal(data []byte) error {
	_, err := obj.UnmarshalFrom(data)
	return err
}

// UnmarshalFrom is very similar to Unmarshal, but also returns the count of data it read.
func (obj *TransactionWitness) UnmarshalFrom(data []byte) (int, error) {
	var n int

	{
		iLen, ni := binary.Uvarint(data[n:])
		if ni <= 0 {
			return 0, errors.Wrap(ranger.ErrShortRead, "Obtaining length of TransactionWitness.Data")
		}
		n += ni
		if iLen > 20 {
			return 0, errors.Wrap(ranger.ErrTooMany, "TransactionWitness.Data")
		}
		obj.Data = make([][32]byte, iLen)
		for i := uint64(0); i < iLen; i++ {

			if len(data[n:]) < 1 {
				return 0, errors.Wrap(ranger.ErrShortRead, "TransactionWitness.Data")
			}
			{
				iL := uint64(32)

				if iL > uint64(len(data[n:])) {
					return 0, errors.Wrap(ranger.ErrShortRead, "TransactionWitness.Data")
				}
				n += copy(obj.Data[i][:], data[n:uint64(n)+iL])
			}
		}
	}

	return n, nil
}

// Transactions is for holding a list of transaction structs.

type Transactions struct {
	// Transactions is the list of transactions
	Transactions []*Transaction
}

// Marshal returns a byte array containing the marshaled representation of Transactions, or nil and an error.
func (obj *Transactions) Marshal() ([]byte, error) {
	data := make([]byte, obj.Size())
	n, err := obj.MarshalTo(data)
	if err != nil {
		return nil, errors.Wrap(err, "Transactions")
	}

	if n != len(data) {
		return nil, errors.Wrapf(ranger.ErrMarshalLength, "%s %d %d", "Transactions", n, len(data))
	}

	return data, nil
}

// MarshalTo accepts a byte array with pre-allocated space (see Size()) for Transactions.
// It returns how many bytes it wrote to the array, or 0 and an error.
func (obj *Transactions) MarshalTo(data []byte) (int, error) {

	var n int

	if len(obj.Transactions) > 20 {
		return 0, errors.Wrap(ranger.ErrTooMany, "Transactions.Transactions")
	}

	n += binary.PutUvarint(data[n:], uint64(len(obj.Transactions)))
	for _, item := range obj.Transactions {

		if item == nil {
			return 0, errors.Wrap(ranger.ErrShortWrite, "Transactions.Transactions is nil, cannot serialise")
		}

		if len(data[n:]) < item.Size() {
			return 0, errors.Wrap(ranger.ErrShortWrite, "Transactions.Transactions")
		}

		{
			ni, err := item.MarshalTo(data[n : n+item.Size()])
			if err != nil {
				return 0, errors.Wrap(err, "Transactions.Transactions")
			}
			n += ni
		}

	}
	return n, nil
}

// Size returns the computed size of Transactions as would-be marshaled
// without actually marshaling it.
func (obj *Transactions) Size() int {
	var n int
	if obj.Transactions == nil {
		// Cannot calculate the value for missing fields
		return 0
	} else {
		n += ranger.UvarintSize(uint64(len(obj.Transactions)))
		for _, item := range obj.Transactions {
			n += item.Size()
		}
	}
	return n
}

// Unmarshal accepts Transactions's binary representation and transforms the
// Transactions used as the object. It returns any error.
func (obj *Transactions) Unmarshal(data []byte) error {
	_, err := obj.UnmarshalFrom(data)
	return err
}

// UnmarshalFrom is very similar to Unmarshal, but also returns the count of data it read.
func (obj *Transactions) UnmarshalFrom(data []byte) (int, error) {
	var n int

	{
		iLen, ni := binary.Uvarint(data[n:])
		if ni <= 0 {
			return 0, errors.Wrap(ranger.ErrShortRead, "Obtaining length of Transactions.Transactions")
		}
		n += ni
		if iLen > 20 {
			return 0, errors.Wrap(ranger.ErrTooMany, "Transactions.Transactions")
		}
		obj.Transactions = make([]*Transaction, iLen)
		for i := uint64(0); i < iLen; i++ {
			obj.Transactions[i] = &Transaction{}

			if len(data[n:]) < 4 {
				return 0, errors.Wrap(ranger.ErrShortRead, "Transactions.Transactions")
			}
			{
				ni, err := obj.Transactions[i].UnmarshalFrom(data[n:])
				if err != nil {
					return 0, errors.Wrap(err, "Transactions.Transactions")
				}
				n += ni
			}
		}
	}

	return n, nil
}

type TransferTransaction struct {
	Inputs    []*TransactionInput
	Outputs   []*TransactionOutput
	Witnesses []*TransactionWitness
	LockTime  uint32
}

// Marshal returns a byte array containing the marshaled representation of TransferTransaction, or nil and an error.
func (obj *TransferTransaction) Marshal() ([]byte, error) {
	data := make([]byte, obj.Size())
	n, err := obj.MarshalTo(data)
	if err != nil {
		return nil, errors.Wrap(err, "TransferTransaction")
	}

	if n != len(data) {
		return nil, errors.Wrapf(ranger.ErrMarshalLength, "%s %d %d", "TransferTransaction", n, len(data))
	}

	return data, nil
}

// MarshalTo accepts a byte array with pre-allocated space (see Size()) for TransferTransaction.
// It returns how many bytes it wrote to the array, or 0 and an error.
func (obj *TransferTransaction) MarshalTo(data []byte) (int, error) {

	var n int

	if len(obj.Inputs) > 20 {
		return 0, errors.Wrap(ranger.ErrTooMany, "TransferTransaction.Inputs")
	}

	if len(obj.Inputs) != len(obj.Witnesses) {
		return 0, errors.Wrap(ranger.ErrLengthMismatch, "TransferTransaction: Inputs and Witnesses")
	}
	n += binary.PutUvarint(data[n:], uint64(len(obj.Inputs)))
	for _, item := range obj.Inputs {

		if item == nil {
			return 0, errors.Wrap(ranger.ErrShortWrite, "TransferTransaction.Inputs is nil, cannot serialise")
		}

		if len(data[n:]) < item.Size() {
			return 0, errors.Wrap(ranger.ErrShortWrite, "TransferTransaction.Inputs")
		}

		{
			ni, err := item.MarshalTo(data[n : n+item.Size()])
			if err != nil {
				return 0, errors.Wrap(err, "TransferTransaction.Inputs")
			}
			n += ni
		}

	}
	if len(obj.Outputs) > 20 {
		return 0, errors.Wrap(ranger.ErrTooMany, "TransferTransaction.Outputs")
	}

	n += binary.PutUvarint(data[n:], uint64(len(obj.Outputs)))
	for _, item := range obj.Outputs {

		if item == nil {
			return 0, errors.Wrap(ranger.ErrShortWrite, "TransferTransaction.Outputs is nil, cannot serialise")
		}

		if len(data[n:]) < item.Size() {
			return 0, errors.Wrap(ranger.ErrShortWrite, "TransferTransaction.Outputs")
		}

		{
			ni, err := item.MarshalTo(data[n : n+item.Size()])
			if err != nil {
				return 0, errors.Wrap(err, "TransferTransaction.Outputs")
			}
			n += ni
		}

	}
	if len(obj.Witnesses) > 20 {
		return 0, errors.Wrap(ranger.ErrTooMany, "TransferTransaction.Witnesses")
	}

	if len(obj.Witnesses) != len(obj.Inputs) {
		return 0, errors.Wrap(ranger.ErrLengthMismatch, "TransferTransaction: Witnesses and Inputs")
	}
	n += binary.PutUvarint(data[n:], uint64(len(obj.Witnesses)))
	for _, item := range obj.Witnesses {

		if item == nil {
			return 0, errors.Wrap(ranger.ErrShortWrite, "TransferTransaction.Witnesses is nil, cannot serialise")
		}

		if len(data[n:]) < item.Size() {
			return 0, errors.Wrap(ranger.ErrShortWrite, "TransferTransaction.Witnesses")
		}

		{
			ni, err := item.MarshalTo(data[n : n+item.Size()])
			if err != nil {
				return 0, errors.Wrap(err, "TransferTransaction.Witnesses")
			}
			n += ni
		}

	}
	return n, nil
}

// Size returns the computed size of TransferTransaction as would-be marshaled
// without actually marshaling it.
func (obj *TransferTransaction) Size() int {
	var n int
	if obj.Inputs == nil {
		// Cannot calculate the value for missing fields
		return 0
	} else {
		n += ranger.UvarintSize(uint64(len(obj.Inputs)))
		for _, item := range obj.Inputs {
			n += item.Size()
		}
	}
	if obj.Outputs == nil {
		// Cannot calculate the value for missing fields
		return 0
	} else {
		n += ranger.UvarintSize(uint64(len(obj.Outputs)))
		for _, item := range obj.Outputs {
			n += item.Size()
		}
	}
	if obj.Witnesses == nil {
		// Cannot calculate the value for missing fields
		return 0
	} else {
		n += ranger.UvarintSize(uint64(len(obj.Witnesses)))
		for _, item := range obj.Witnesses {
			n += item.Size()
		}
	}
	return n
}

// Unmarshal accepts TransferTransaction's binary representation and transforms the
// TransferTransaction used as the object. It returns any error.
func (obj *TransferTransaction) Unmarshal(data []byte) error {
	_, err := obj.UnmarshalFrom(data)
	return err
}

// UnmarshalFrom is very similar to Unmarshal, but also returns the count of data it read.
func (obj *TransferTransaction) UnmarshalFrom(data []byte) (int, error) {
	var n int

	{
		iLen, ni := binary.Uvarint(data[n:])
		if ni <= 0 {
			return 0, errors.Wrap(ranger.ErrShortRead, "Obtaining length of TransferTransaction.Inputs")
		}
		n += ni
		if iLen > 20 {
			return 0, errors.Wrap(ranger.ErrTooMany, "TransferTransaction.Inputs")
		}
		obj.Inputs = make([]*TransactionInput, iLen)
		for i := uint64(0); i < iLen; i++ {
			obj.Inputs[i] = &TransactionInput{}

			if len(data[n:]) < 4 {
				return 0, errors.Wrap(ranger.ErrShortRead, "TransferTransaction.Inputs")
			}
			{
				ni, err := obj.Inputs[i].UnmarshalFrom(data[n:])
				if err != nil {
					return 0, errors.Wrap(err, "TransferTransaction.Inputs")
				}
				n += ni
			}
		}
	}

	{
		iLen, ni := binary.Uvarint(data[n:])
		if ni <= 0 {
			return 0, errors.Wrap(ranger.ErrShortRead, "Obtaining length of TransferTransaction.Outputs")
		}
		n += ni
		if iLen > 20 {
			return 0, errors.Wrap(ranger.ErrTooMany, "TransferTransaction.Outputs")
		}
		obj.Outputs = make([]*TransactionOutput, iLen)
		for i := uint64(0); i < iLen; i++ {
			obj.Outputs[i] = &TransactionOutput{}

			if len(data[n:]) < 2 {
				return 0, errors.Wrap(ranger.ErrShortRead, "TransferTransaction.Outputs")
			}
			{
				ni, err := obj.Outputs[i].UnmarshalFrom(data[n:])
				if err != nil {
					return 0, errors.Wrap(err, "TransferTransaction.Outputs")
				}
				n += ni
			}
		}
	}

	{
		iLen, ni := binary.Uvarint(data[n:])
		if ni <= 0 {
			return 0, errors.Wrap(ranger.ErrShortRead, "Obtaining length of TransferTransaction.Witnesses")
		}
		n += ni
		if iLen > 20 {
			return 0, errors.Wrap(ranger.ErrTooMany, "TransferTransaction.Witnesses")
		}
		obj.Witnesses = make([]*TransactionWitness, iLen)
		for i := uint64(0); i < iLen; i++ {
			obj.Witnesses[i] = &TransactionWitness{}

			if len(data[n:]) < 1 {
				return 0, errors.Wrap(ranger.ErrShortRead, "TransferTransaction.Witnesses")
			}
			{
				ni, err := obj.Witnesses[i].UnmarshalFrom(data[n:])
				if err != nil {
					return 0, errors.Wrap(err, "TransferTransaction.Witnesses")
				}
				n += ni
			}
		}
	}

	if len(obj.Inputs) != len(obj.Witnesses) {
		return 0, errors.Wrap(ranger.ErrLengthMismatch, "TransferTransaction: Inputs and Witnesses")
	}

	if len(obj.Witnesses) != len(obj.Inputs) {
		return 0, errors.Wrap(ranger.ErrLengthMismatch, "TransferTransaction: Witnesses and Inputs")
	}
	return n, nil
}
