package common

import (
	"encoding/hex"
	"errors"
)

const (
	EscrowIDSize = 16
	ObjectIDSize = 16
	BlockIDSize  = 16
)

type EscrowID [EscrowIDSize]byte

func BytesToEscrowID(x []byte) (EscrowID, error) {
	if len(x) != EscrowIDSize {
		return EscrowID{}, errors.New("bad size for escrow ID")
	}

	var id EscrowID
	copy(id[:], x[0:EscrowIDSize])
	return id, nil
}

func (id EscrowID) String() string {
	return hex.EncodeToString(id[:])
}

type ObjectID [ObjectIDSize]byte

func BytesToObjectID(x []byte) (ObjectID, error) {
	if len(x) != ObjectIDSize {
		return ObjectID{}, errors.New("bad size for object ID")
	}

	var id ObjectID
	copy(id[:], x[0:ObjectIDSize])
	return id, nil
}

func (id ObjectID) String() string {
	return hex.EncodeToString(id[:])
}

type BlockID [BlockIDSize]byte

func BytesToBlockID(x []byte) (BlockID, error) {
	if len(x) != BlockIDSize {
		return BlockID{}, errors.New("bad size for block ID")
	}

	var id BlockID
	copy(id[:], x[0:BlockIDSize])
	return id, nil
}

func (id BlockID) String() string {
	return hex.EncodeToString(id[:])
}
