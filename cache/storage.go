package cache

import (
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/sirupsen/logrus"
)

type CacheStorage struct {
	l       *logrus.Logger
	escrows map[common.EscrowID]*escrowState
}

type escrowState struct {
	data     map[common.BlockID][]byte
	metadata map[common.ObjectID]*ccmsg.ObjectMetadata
}

func NewCacheStorage(l *logrus.Logger) (*CacheStorage, error) {
	return &CacheStorage{
		l:       l,
		escrows: make(map[common.EscrowID]*escrowState),
	}, nil
}

func newEscrowState() *escrowState {
	return &escrowState{
		data:     make(map[common.BlockID][]byte),
		metadata: make(map[common.ObjectID]*ccmsg.ObjectMetadata),
	}
}

func (s *CacheStorage) getEscrowState(escrowID common.EscrowID) *escrowState {
	es, ok := s.escrows[escrowID]
	if !ok {
		es = newEscrowState()
		s.escrows[escrowID] = es
	}
	return es
}

// Returns (nil, nil) if the object does not exist; the error part of the rval is reserved for e.g. storage engine
// errors.
func (s *CacheStorage) GetMetadata(escrowID common.EscrowID, objectID common.ObjectID) (*ccmsg.ObjectMetadata, error) {
	s.l.WithFields(logrus.Fields{
		"escrowID": escrowID,
		"objectID": objectID,
	}).Debug("(*CacheStorage).GetMetadata")

	es := s.getEscrowState(escrowID)
	m, _ := es.metadata[objectID]
	return m, nil
}

func (s *CacheStorage) GetData(escrowID common.EscrowID, blockID common.BlockID) ([]byte, error) {
	s.l.WithFields(logrus.Fields{
		"escrowID": escrowID,
		"blockID":  blockID,
	}).Debug("(*CacheStorage).GetData")

	es := s.getEscrowState(escrowID)
	b, _ := es.data[blockID]
	return b, nil
}

func (s *CacheStorage) PutMetadata(escrowID common.EscrowID, objectID common.ObjectID, m *ccmsg.ObjectMetadata) error {
	s.l.WithFields(logrus.Fields{
		"escrowID": escrowID,
		"objectID": objectID,
	}).Debug("(*CacheStorage).PutMetadata")

	es := s.getEscrowState(escrowID)
	es.metadata[objectID] = m
	return nil
}

func (s *CacheStorage) PutData(escrowID common.EscrowID, blockID common.BlockID, data []byte) error {
	s.l.WithFields(logrus.Fields{
		"escrowID": escrowID,
		"blockID":  blockID,
	}).Debug("(*CacheStorage).PutData")

	es := s.getEscrowState(escrowID)
	es.data[blockID] = data
	return nil
}
