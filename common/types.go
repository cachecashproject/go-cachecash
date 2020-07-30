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
	ChunkIDSize  = 16
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

type ChunkID [ChunkIDSize]byte

func BytesToChunkID(x []byte) (ChunkID, error) {
	if len(x) != ChunkIDSize {
		return ChunkID{}, errors.New("bad size for chunk ID")
	}

	var id ChunkID
	copy(id[:], x[0:ChunkIDSize])
	return id, nil
}

func (id ChunkID) String() string {
	return hex.EncodeToString(id[:])
}

func (id ChunkID) Value() (driver.Value, error) {
	return id[:], nil
}

func (id *ChunkID) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		val, err := BytesToChunkID(src)
		if err != nil {
			return err
		}
		*id = val
		return nil
	default:
		return errors.New("incompatible type for ChunkID")
	}
}

// TODO: I don't think that we need to support `fieldType` here, but do we need to support `shouldBeNull`?  Their
// purpose is explained in the Randomizer docs.
func (id *ChunkID) Randomize(nextInt func() int64, fieldType string, shouldBeNull bool) {
	// TODO: The point of nextInt(), which returns sequential "random" integers, is to ensure that generated values are
	// unique in situations that require it.  This construction may not actually do that.
	var val ChunkID
	for i := 0; i < len(val); i++ {
		val[i] = byte(nextInt() % 256)
	}
	*id = val
}
