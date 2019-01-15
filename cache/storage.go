package cache

import (
	"github.com/kelleyk/go-cachecash/ccmsg"
	"github.com/kelleyk/go-cachecash/common"
	"github.com/sirupsen/logrus"
)

type CacheStorage struct {
	l       *logrus.Logger
	escrows map[uint64]*escrowState
}

type escrowState struct {
	data     map[common.BlockID][]byte
	metadata map[uint64]*ccmsg.ObjectMetadata
}

func NewCacheStorage(l *logrus.Logger) (*CacheStorage, error) {
	return &CacheStorage{
		l:       l,
		escrows: make(map[uint64]*escrowState),
	}, nil
}

func newEscrowState() *escrowState {
	return &escrowState{
		data:     make(map[common.BlockID][]byte),
		metadata: make(map[uint64]*ccmsg.ObjectMetadata),
	}
}

func (s *CacheStorage) getEscrowState(escrowID uint64) *escrowState {
	es, ok := s.escrows[escrowID]
	if !ok {
		es = newEscrowState()
		s.escrows[escrowID] = es
	}
	return es
}

// Returns (nil, nil) if the object does not exist; the error part of the rval is reserved for e.g. storage engine
// errors.
func (s *CacheStorage) GetMetadata(escrowID, objectID uint64) (*ccmsg.ObjectMetadata, error) {
	s.l.WithFields(logrus.Fields{
		"escrowID": escrowID,
		"objectID": objectID,
	}).Debug("(*CacheStorage).GetMetadata")

	es := s.getEscrowState(escrowID)
	m, _ := es.metadata[objectID]
	return m, nil
}

func (s *CacheStorage) GetData(escrowID uint64, blockID common.BlockID) ([]byte, error) {
	s.l.WithFields(logrus.Fields{
		"escrowID": escrowID,
		"blockID":  blockID,
	}).Debug("(*CacheStorage).GetData")

	es := s.getEscrowState(escrowID)
	b, _ := es.data[blockID]
	return b, nil
}

func (s *CacheStorage) PutMetadata(escrowID, objectID uint64, m *ccmsg.ObjectMetadata) error {
	s.l.WithFields(logrus.Fields{
		"escrowID": escrowID,
		"objectID": objectID,
	}).Debug("(*CacheStorage).PutMetadata")

	es := s.getEscrowState(escrowID)
	es.metadata[objectID] = m
	return nil
}

func (s *CacheStorage) PutData(escrowID uint64, blockID common.BlockID, data []byte) error {
	s.l.WithFields(logrus.Fields{
		"escrowID": escrowID,
		"blockID":  blockID,
	}).Debug("(*CacheStorage).PutData")

	es := s.getEscrowState(escrowID)
	es.data[blockID] = data
	return nil
}
