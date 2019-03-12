package cache

import (
	"time"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/dgraph-io/badger"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type CacheStorage struct {
	l  *logrus.Logger
	kv *badger.DB
}

func NewCacheStorage(l *logrus.Logger, badgerDirectory string) (*CacheStorage, error) {
	opts := badger.DefaultOptions
	opts.Dir = badgerDirectory
	opts.ValueDir = badgerDirectory
	kv, err := badger.Open(opts)

	if err != nil {
		return nil, errors.Wrap(err, "failed to open badger database")
	}

	return &CacheStorage{
		l:  l,
		kv: kv,
	}, nil
}

func makeDataKey(escrowID common.EscrowID, blockID common.BlockID) []byte {
	return []byte("data-" + string(escrowID[:]) + "-" + string(blockID[:]))
}

func makeMetaKey(escrowID common.EscrowID, objectID common.ObjectID) []byte {
	return []byte("meta-" + string(escrowID[:]) + "-" + string(objectID[:]))
}

func (s *CacheStorage) Close() error {
	return s.kv.Close()
}

func (s *CacheStorage) GetRawBytes(key []byte) ([]byte, error) {
	var value []byte

	err := s.kv.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			// if the key doesn't exist, return nil
			return nil
		}

		value, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to select from badger db")
	}

	return value, nil
}

func (s *CacheStorage) PutRawBytes(key []byte, bytes []byte, ttl *time.Duration) error {
	err := s.kv.Update(func(txn *badger.Txn) error {
		if ttl == nil {
			return txn.Set(key, bytes)
		} else {
			return txn.SetWithTTL(key, bytes, *ttl)
		}
	})

	if err != nil {
		return errors.Wrap(err, "failed to write to badger db")
	}

	return nil
}

// Returns (nil, nil) if the object does not exist; the error part of the rval is reserved for e.g. storage engine
// errors.
func (s *CacheStorage) GetMetadata(escrowID common.EscrowID, objectID common.ObjectID) (*ccmsg.ObjectMetadata, error) {
	s.l.WithFields(logrus.Fields{
		"escrowID": escrowID,
		"objectID": objectID,
	}).Debug("(*CacheStorage).GetMetadata")

	key := makeMetaKey(escrowID, objectID)

	bytes, err := s.GetRawBytes(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get bytes from kv")
	}

	// key not found
	if bytes == nil {
		return nil, nil
	}

	meta := &ccmsg.ObjectMetadata{}
	err = proto.Unmarshal(bytes, meta)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal metadata")
	}

	return meta, nil
}

func (s *CacheStorage) GetData(escrowID common.EscrowID, blockID common.BlockID) ([]byte, error) {
	s.l.WithFields(logrus.Fields{
		"escrowID": escrowID,
		"blockID":  blockID,
	}).Debug("(*CacheStorage).GetData")

	key := makeDataKey(escrowID, blockID)

	return s.GetRawBytes(key)
}

func (s *CacheStorage) PutMetadata(escrowID common.EscrowID, objectID common.ObjectID, m *ccmsg.ObjectMetadata) error {
	s.l.WithFields(logrus.Fields{
		"escrowID": escrowID,
		"objectID": objectID,
	}).Debug("(*CacheStorage).PutMetadata")

	key := makeMetaKey(escrowID, objectID)

	bytes, err := proto.Marshal(m)
	if err != nil {
		return errors.Wrap(err, "failed to marshal metadata")
	}

	err = s.PutRawBytes(key, bytes, nil)
	if err != nil {
		return errors.Wrap(err, "failed to put bytes into kv")
	}

	return nil
}

func (s *CacheStorage) PutData(escrowID common.EscrowID, blockID common.BlockID, data []byte) error {
	s.l.WithFields(logrus.Fields{
		"escrowID": escrowID,
		"blockID":  blockID,
	}).Debug("(*CacheStorage).PutData")

	key := makeDataKey(escrowID, blockID)

	return s.PutRawBytes(key, data, nil)
}
