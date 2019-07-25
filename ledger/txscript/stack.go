package txscript

import (
	"errors"
	"fmt"
)

// ScriptStack represents a Bitcoin script stack.  This type offers utility functions for interpreting values that will
// be pushed or popped as various types.
type ScriptStack struct {
	data [][]byte
}

func (st *ScriptStack) Size() int {
	return len(st.data)
}

func (st *ScriptStack) PushBytes(v []byte) {
	st.data = append(st.data, v)
}

// TODO: Tests should cover the case where `idx<0`.
func (st *ScriptStack) PeekBytes(idx int) ([]byte, error) {
	if idx < 0 || idx >= len(st.data) {
		return nil, fmt.Errorf("stack offset %v out of range", idx)
	}
	return st.data[idx], nil
}

func (st *ScriptStack) PopBytes() ([]byte, error) {
	if len(st.data) == 0 {
		return nil, errors.New("stack empty")
	}
	v := st.data[len(st.data)-1]
	st.data = st.data[:len(st.data)-1]
	return v, nil
}

func (st *ScriptStack) PushInt(v scriptNum) {
	st.PushBytes(v.Bytes())
}

func (st *ScriptStack) PeekInt(idx int) (scriptNum, error) {
	v, err := st.PeekBytes(idx)
	if err != nil {
		return 0, err
	}
	return makeScriptNum(v, true, 4) // XXX: Hardwiring these things
}

func (st *ScriptStack) PopInt() (scriptNum, error) {
	v, err := st.PopBytes()
	if err != nil {
		return 0, err
	}
	return makeScriptNum(v, true, 4) // XXX: Hardwiring these things
}

func (st *ScriptStack) PushBool(v bool) {
	st.PushBytes(fromBool(v))
}

func (st *ScriptStack) PeekBool(idx int) (bool, error) {
	v, err := st.PeekBytes(idx)
	if err != nil {
		return false, err
	}
	return asBool(v), nil
}

func (st *ScriptStack) PopBool() (bool, error) {
	v, err := st.PopBytes()
	if err != nil {
		return false, err
	}
	return asBool(v), nil
}

// From `btcsuite/btcd`; available under the ISC license.
func fromBool(v bool) []byte {
	var vb []byte
	if v {
		vb = []byte{1}
	}
	return vb
}

// asBool gets the boolean value of the byte array.
//
// From `btcsuite/btcd`; available under the ISC license.
func asBool(t []byte) bool {
	for i := range t {
		if t[i] != 0 {
			// Negative 0 is also considered false.
			if i == len(t)-1 && t[i] == 0x80 {
				return false
			}
			return true
		}
	}
	return false
}
