package kv

import (
	"bytes"
	"crypto/rand"
	"sync"
)

type casvalue struct {
	value []byte
	nonce []byte
}

// MemoryTable is the underlying data structure used to manage memory in the
// MemoryDriver.
type MemoryTable map[string]map[string]casvalue

// MemoryDriver implements a very basic key/value store over a map.
type MemoryDriver struct {
	table MemoryTable
	mutex sync.Mutex
}

// NewMemoryDriver returns a Driver that is basically a locked map. You can
// provide a series of initial values or just pass nil to get an empty set.
func NewMemoryDriver(preinit MemoryTable) Driver {
	if preinit == nil {
		preinit = MemoryTable{}
	}

	return &MemoryDriver{table: preinit}
}

// lock free get
func (md *MemoryDriver) get(member, key string) ([]byte, []byte, error) {
	memberTable, ok := md.table[member]
	if ok {
		value, ok := memberTable[key]
		if ok {
			return value.value, value.nonce, nil
		}
	}

	return nil, nil, ErrUnsetValue
}

// lock-free set
func (md *MemoryDriver) set(member, key string, onlyNotExists bool, value []byte) ([]byte, error) {
	memberTable, ok := md.table[member]
	if !ok {
		memberTable = map[string]casvalue{}
		md.table[member] = memberTable
	}
	buf := make([]byte, 12)

	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}

	if _, ok := memberTable[key]; ok && onlyNotExists {
		return nil, ErrAlreadySet
	}

	memberTable[key] = casvalue{value: value, nonce: buf}
	return buf, nil
}

// Create creates a key
func (md *MemoryDriver) Create(member, key string, value []byte) ([]byte, error) {
	return md.set(member, key, true, value)
}

// Delete deletes a key with optional CAS nonce
func (md *MemoryDriver) Delete(member, key string, nonce []byte) error {
	md.mutex.Lock()
	defer md.mutex.Unlock()

	memberTable, ok := md.table[member]
	if !ok {
		if nonce != nil {
			return ErrNotEqual
		}
		return ErrUnsetValue
	}

	val, ok := memberTable[key]
	if !ok {
		if nonce != nil {
			return ErrNotEqual
		}
		return ErrUnsetValue
	}

	if nonce != nil {
		if !bytes.Equal(val.nonce, nonce) {
			return ErrNotEqual
		}
	}

	delete(memberTable, key)
	return nil
}

// Get retrieves a value from the k/v store.
func (md *MemoryDriver) Get(member, key string) ([]byte, error) {
	md.mutex.Lock()
	defer md.mutex.Unlock()

	value, _, err := md.get(member, key)
	return value, err
}

// Set sets the value in the k/v store.
func (md *MemoryDriver) Set(member, key string, value []byte) ([]byte, error) {
	md.mutex.Lock()
	defer md.mutex.Unlock()

	return md.set(member, key, false, value)
}

// CAS implements compare-and-swap for the k/v store.
func (md *MemoryDriver) CAS(member, key string, origSentinel, origValue, value []byte) ([]byte, error) {
	md.mutex.Lock()
	defer md.mutex.Unlock()

	out, nonce, err := md.get(member, key)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(out, origValue) || !bytes.Equal(nonce, origSentinel) {
		return nil, ErrNotEqual
	}

	return md.set(member, key, false, value)
}
