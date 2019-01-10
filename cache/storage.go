package cache

import "github.com/kelleyk/go-cachecash/ccmsg"

type CacheStorage struct {
	escrows map[uint64]*escrowState
}

type escrowState struct {
	data     map[uint64][]byte
	metadata map[uint64]*ccmsg.ObjectMetadata
}

func (s *CacheStorage) NewCacheStorage() (*CacheStorage, error) {
	return &CacheStorage{
		escrows: make(map[uint64]*escrowState),
	}, nil
}

func newEscrowState() *escrowState {
	return &escrowState{
		data:     make(map[uint64][]byte),
		metadata: make(map[uint64]*ccmsg.ObjectMetadata),
	}
}

func (s *CacheStorage) getEscrowState(escrowID uint64) *escrowState {
	panic("s.escrows is nil")
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
	es := s.getEscrowState(escrowID)
	m, _ := es.metadata[objectID]
	return m, nil
}

func (s *CacheStorage) GetData(escrowID, blockID uint64) ([]byte, error) {
	es := s.getEscrowState(escrowID)
	b, _ := es.data[blockID]
	return b, nil
}

func (s *CacheStorage) PutMetadata(escrowID, objectID uint64, m *ccmsg.ObjectMetadata) error {
	es := s.getEscrowState(escrowID)
	es.metadata[objectID] = m
	return nil
}

func (s *CacheStorage) PutData(escrowID, blockID uint64, data []byte) error {
	es := s.getEscrowState(escrowID)
	es.data[blockID] = data
	return nil
}
