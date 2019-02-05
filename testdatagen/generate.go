package testdatagen

import (
	"crypto/sha512"
	"math"
	"net"

	cachecash "github.com/cachecashproject/go-cachecash"
	"github.com/cachecashproject/go-cachecash/cache"
	"github.com/cachecashproject/go-cachecash/catalog"
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/provider"
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
}

type TestScenario struct {
	L *logrus.Logger

	Upstream catalog.Upstream

	Provider *provider.ContentProvider
	Catalog  catalog.ContentCatalog
	Escrow   *provider.Escrow
	EscrowID common.EscrowID

	Params   *TestScenarioParams
	Obj      cachecash.ContentObject
	ObjectID common.ObjectID
	Caches   []*cache.Cache
}

func (ts *TestScenario) BlockCount() uint64 {
	return uint64(math.Ceil(float64(ts.Params.ObjectSize) / float64(ts.Params.BlockSize)))
}

// XXX: This shouldn't be necessary; we should reduce/eliminate the use of ContentObject.
func contentObjectToBytes(obj cachecash.ContentObject) ([]byte, error) {
	var data []byte
	for i := 0; i < obj.BlockCount(); i++ {
		block, err := obj.GetBlock(uint32(i))
		if err != nil {
			return nil, errors.New("failed to get data block")
		}
		data = append(data, block...)
	}
	return data, nil
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
		obj, err := cachecash.RandomContentBuffer(uint32(ts.BlockCount()), uint32(ts.Params.BlockSize))
		if err != nil {
			return nil, err
		}
		ts.Obj = obj
	}

	// Create upstream.
	if ts.Upstream == nil {
		if params.MockUpstream {
			mockUpstream, err := catalog.NewMockUpstream(ts.L)
			if err != nil {
				return nil, errors.Wrap(err, "failed to create mock upstream")
			}
			if params.GenerateObject {
				objData, err := contentObjectToBytes(ts.Obj)
				if err != nil {
					return nil, errors.Wrap(err, "failed to convert ContentObject to bytes")
				}
				mockUpstream.Objects["/foo/bar"] = objData
			}
			ts.Upstream = mockUpstream
		} else {
			ts.Upstream, err = catalog.NewHTTPUpstream(ts.L, "http://localhost:8081")
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

	// Create a provider.
	_, providerPrivateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate provider keypair")
	}
	prov, err := provider.NewContentProvider(ts.L, cat, providerPrivateKey)
	if err != nil {
		return nil, err
	}
	ts.Provider = prov

	// Create escrow and add it to the provider.
	escrow, err := prov.NewEscrow(&ccmsg.EscrowInfo{
		Id:              ts.EscrowID[:],
		DrawDelay:       5,
		ExpirationDelay: 5,
		StartBlock:      42,
		TicketsPerBlock: []*ccmsg.Segment{
			&ccmsg.Segment{Length: 10, Value: 100},
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

		escrow.Caches = append(escrow.Caches, &provider.ParticipatingCache{
			InnerMasterKey: innerMasterKey,
			PublicKey:      public,
			Inetaddr:       net.ParseIP("127.0.0.1"),
			Port:           uint32(9000 + i),
		})

		c, err := cache.NewCache(ts.L)
		if err != nil {
			return nil, err
		}
		ce := &cache.Escrow{
			InnerMasterKey:           innerMasterKey,
			OuterMasterKey:           testutil.RandBytes(16),
			ProviderCacheServiceAddr: "localhost:8082",
		}
		if params.GenerateObject {
			if err := c.Storage.PutMetadata(ts.EscrowID, ts.ObjectID, &ccmsg.ObjectMetadata{
				BlockSize:  ts.Params.BlockSize,
				ObjectSize: ts.Params.ObjectSize,
			}); err != nil {
				return nil, err
			}
			for j := 0; j < ts.Obj.BlockCount(); j++ {
				data, err := ts.Obj.GetBlock(uint32(j))
				if err != nil {
					return nil, err
				}
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
