package common

//
// Types that will be stored in the database via `sqlboiler` must implement several interfaces:
//   - sqlboiler.randomize.Randomizer
//

import (
	"database/sql/driver"
	"encoding/base64"
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

func (id EscrowID) Value() (driver.Value, error) {
	return id[:], nil
}

func (id *EscrowID) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		val, err := BytesToEscrowID(src)
		if err != nil {
			return err
		}
		*id = val
		return nil
	default:
		return errors.New("incompatible type for EscrowID")
	}
}

// TODO: I don't think that we need to support `fieldType` here, but do we need to support `shouldBeNull`?  Their
// purpose is explained in the Randomizer docs.
func (id *EscrowID) Randomize(nextInt func() int64, fieldType string, shouldBeNull bool) {
	// TODO: The point of nextInt(), which returns sequential "random" integers, is to ensure that generated values are
	// unique in situations that require it.  This construction may not actually do that.
	var val EscrowID
	for i := 0; i < len(val); i++ {
		val[i] = byte(nextInt() % 256)
	}
	*id = val
}

// These are here to support marshaling as JSON map keys.
func (id EscrowID) MarshalText() ([]byte, error) {
	return []byte(base64.StdEncoding.EncodeToString(id[:])), nil
}

func (id EscrowID) UnmarshalText(x []byte) error {
	y, err := base64.StdEncoding.DecodeString(string(x))
	if err != nil {
		return err
	}
	val, err := BytesToEscrowID(y)
	if err != nil {
		return err
	}
	copy(id[:], val[:])
	return nil
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

func (id ObjectID) Value() (driver.Value, error) {
	return id[:], nil
}

func (id *ObjectID) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		val, err := BytesToObjectID(src)
		if err != nil {
			return err
		}
		*id = val
		return nil
	default:
		return errors.New("incompatible type for ObjectID")
	}
}

// TODO: I don't think that we need to support `fieldType` here, but do we need to support `shouldBeNull`?  Their
// purpose is explained in the Randomizer docs.
func (id *ObjectID) Randomize(nextInt func() int64, fieldType string, shouldBeNull bool) {
	// TODO: The point of nextInt(), which returns sequential "random" integers, is to ensure that generated values are
	// unique in situations that require it.  This construction may not actually do that.
	var val ObjectID
	for i := 0; i < len(val); i++ {
		val[i] = byte(nextInt() % 256)
	}
	*id = val
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

func (id BlockID) Value() (driver.Value, error) {
	return id[:], nil
}

func (id *BlockID) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		val, err := BytesToBlockID(src)
		if err != nil {
			return err
		}
		*id = val
		return nil
	default:
		return errors.New("incompatible type for BlockID")
	}
}

// TODO: I don't think that we need to support `fieldType` here, but do we need to support `shouldBeNull`?  Their
// purpose is explained in the Randomizer docs.
func (id *BlockID) Randomize(nextInt func() int64, fieldType string, shouldBeNull bool) {
	// TODO: The point of nextInt(), which returns sequential "random" integers, is to ensure that generated values are
	// unique in situations that require it.  This construction may not actually do that.
	var val BlockID
	for i := 0; i < len(val); i++ {
		val[i] = byte(nextInt() % 256)
	}
	*id = val
}
