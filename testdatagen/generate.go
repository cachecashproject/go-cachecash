package testdatagen

import (
	"crypto/sha512"
	"math"
	"net"
	"time"

	"github.com/cachecashproject/go-cachecash/cache"
	"github.com/cachecashproject/go-cachecash/catalog"
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/publisher"
	"github.com/cachecashproject/go-cachecash/testutil"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ed25519"
)

type TestScenarioParams struct {
	BlockSize  uint64
	ObjectSize uint64

	// MockUpstream indicates whether a mock upstream should be generated in place of the default HTTP upstream.  If
	// Upstream is non-nil, no upstream is generated, and this value has no effect.
	MockUpstream bool
	// GenerateObject indicates whether or not a random test object (with path "/foo/bar") should be generated.  If
	// MockUpstream is also true, the object is inserted into the mock upstream.
	GenerateObject bool

	// These are optional.  If provided, they override the default that would have been generated.
	L        *logrus.Logger
	Upstream catalog.Upstream

	PublisherCacheServiceAddr string
}

type TestScenario struct {
	L *logrus.Logger

	Upstream catalog.Upstream

	Publisher           *publisher.ContentPublisher
	PublisherPrivateKey ed25519.PrivateKey
	Catalog             catalog.ContentCatalog
	Escrow              *publisher.Escrow
	EscrowID            common.EscrowID

	Params     *TestScenarioParams
	DataBlocks [][]byte
	ObjectID   common.ObjectID
	Caches     []*cache.Cache
}

func (ts *TestScenario) BlockCount() uint64 {
	return uint64(math.Ceil(float64(ts.Params.ObjectSize) / float64(ts.Params.BlockSize)))
}

func (ts *TestScenario) ObjectData() []byte {
	var data []byte
	for _, b := range ts.DataBlocks {
		data = append(data, b...)
	}
	return data
}

func generateBlockID(data []byte) common.BlockID {
	var id common.BlockID
	digest := sha512.Sum384(data)
	copy(id[:], digest[0:common.BlockIDSize])
	return id
}

func GenerateTestScenario(l *logrus.Logger, params *TestScenarioParams) (*TestScenario, error) {
	var err error

	escrowID, err := common.BytesToEscrowID(testutil.RandBytes(common.EscrowIDSize))
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate escrow ID")
	}

	objectID, err := common.BytesToObjectID(testutil.RandBytes(common.ObjectIDSize))
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate object ID")
	}

	ts := &TestScenario{
		L:        params.L,
		Params:   params,
		Upstream: params.Upstream,

		EscrowID: escrowID,
		ObjectID: objectID,
	}

	if ts.L == nil {
		ts.L = logrus.New()
		ts.L.SetLevel(logrus.DebugLevel)
	}

	// Create a content object.
	if params.GenerateObject {
		ts.DataBlocks = make([][]byte, 0, ts.BlockCount())
		for i := 0; i < cap(ts.DataBlocks); i++ {
			ts.DataBlocks = append(ts.DataBlocks, testutil.RandBytes(int(ts.Params.BlockSize)))
		}
	}

	// Create upstream.
	if ts.Upstream == nil {
		if params.MockUpstream {
			mockUpstream, err := catalog.NewMockUpstream(ts.L)
			if err != nil {
				return nil, errors.Wrap(err, "failed to create mock upstream")
			}
			if params.GenerateObject {
				mockUpstream.Objects["/foo/bar"] = ts.ObjectData()
			}
			ts.Upstream = mockUpstream
		} else {
			ts.Upstream, err = catalog.NewHTTPUpstream(ts.L, "http://localhost:8081", 5*time.Minute)
			if err != nil {
				return nil, errors.Wrap(err, "failed to create HTTP upstream")
			}
		}
	}

	// Create content catalog.
	cat, err := catalog.NewCatalog(ts.L, ts.Upstream)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create catalog")
	}
	ts.Catalog = cat

	// Create a keypair for the publisher.
	_, publisherPrivateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate publisher keypair")
	}
	ts.PublisherPrivateKey = publisherPrivateKey

	// Create the publisher.
	prov, err := publisher.NewContentPublisher(ts.L, cat, publisherPrivateKey)
	if err != nil {
		return nil, err
	}
	ts.Publisher = prov

	// Create escrow and add it to the publisher.
	escrow, err := prov.NewEscrow(&ccmsg.EscrowInfo{
		Id:              ts.EscrowID[:],
		DrawDelay:       5,
		ExpirationDelay: 5,
		StartBlock:      42,
		TicketsPerBlock: []*ccmsg.Segment{
			{Length: 10, Value: 100},
		},
	})
	if err != nil {
		return nil, err
	}
	if err := prov.AddEscrow(escrow); err != nil {
		return nil, err
	}
	ts.Escrow = escrow

	// Create caches that are participating in this escrow.
	for i := 0; i < 4; i++ {
		public, _, err := ed25519.GenerateKey(nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate cache keypair")
		}

		// XXX: generate master-key
		innerMasterKey := testutil.RandBytes(16)

		escrow.Caches = append(escrow.Caches, &publisher.ParticipatingCache{
			InnerMasterKey: innerMasterKey,
			PublicKey:      public,
			Inetaddr:       net.ParseIP("127.0.0.1"),
			Port:           uint32(9000 + i),
		})

		c, err := cache.NewCache(ts.L)
		if err != nil {
			return nil, err
		}

		if params.PublisherCacheServiceAddr == "" {
			params.PublisherCacheServiceAddr = "localhost:8082"
		}

		ce := &cache.Escrow{
			InnerMasterKey:            innerMasterKey,
			OuterMasterKey:            testutil.RandBytes(16),
			PublisherCacheServiceAddr: params.PublisherCacheServiceAddr,
		}
		if params.GenerateObject {
			if err := c.Storage.PutMetadata(ts.EscrowID, ts.ObjectID, &ccmsg.ObjectMetadata{
				BlockSize:  ts.Params.BlockSize,
				ObjectSize: ts.Params.ObjectSize,
			}); err != nil {
				return nil, err
			}
			for j := 0; j < int(ts.BlockCount()); j++ {
				data := ts.DataBlocks[j]
				blockID := generateBlockID(data)
				if err := c.Storage.PutData(ts.EscrowID, blockID, data); err != nil {
					return nil, err
				}
			}
		}
		c.Escrows[escrow.ID] = ce
		ts.Caches = append(ts.Caches, c)
	}

	return ts, nil
}
