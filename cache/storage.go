package cache

import (
	"os"

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
	err := os.MkdirAll(badgerDirectory, 0700)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create badger directory")
	}

	opts := badger.DefaultOptions
	opts.Dir = badgerDirectory
	opts.ValueDir = badgerDirectory
	opts.Logger = l.WithFields(logrus.Fields{
		"badger": badgerDirectory,
	})
	kv, err := badger.Open(opts)

	if err != nil {
		return nil, errors.Wrap(err, "failed to open badger database")
	}

	return &CacheStorage{
		l:  l,
		kv: kv,
	}, nil
}

func makeDataKey(escrowID common.EscrowID, chunkID common.ChunkID) []byte {
	return []byte("data-" + string(escrowID[:]) + "-" + string(chunkID[:]))
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

func (s *CacheStorage) PutRawBytes(key []byte, bytes []byte) error {
	err := s.kv.Update(func(txn *badger.Txn) error {
		return txn.Set(key, bytes)
	})

	if err != nil {
		return errors.Wrap(err, "failed to write to badger db")
	}

	return nil
}

func (s *CacheStorage) DeleteRawBytes(key []byte) error {
	err := s.kv.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
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

func (s *CacheStorage) GetData(escrowID common.EscrowID, chunkID common.ChunkID) ([]byte, error) {
	s.l.WithFields(logrus.Fields{
		"escrowID": escrowID,
		"chunkID":  chunkID,
	}).Debug("(*CacheStorage).GetData")

	key := makeDataKey(escrowID, chunkID)

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

	err = s.PutRawBytes(key, bytes)
	if err != nil {
		return errors.Wrap(err, "failed to put bytes into kv")
	}

	return nil
}

func (s *CacheStorage) PutData(escrowID common.EscrowID, chunkID common.ChunkID, data []byte) error {
	s.l.WithFields(logrus.Fields{
		"escrowID": escrowID,
		"chunkID":  chunkID,
	}).Debug("(*CacheStorage).PutData")

	key := makeDataKey(escrowID, chunkID)

	return s.PutRawBytes(key, data)
}

func (s *CacheStorage) DeleteData(escrowID common.EscrowID, chunkID common.ChunkID) error {
	s.l.WithFields(logrus.Fields{
		"escrowID": escrowID,
		"chunkID":  chunkID,
	}).Debug("(*CacheStorage).DeleteData")

	key := makeDataKey(escrowID, chunkID)

	return s.DeleteRawBytes(key)
}
