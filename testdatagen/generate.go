package testdatagen

import (
	"math"
	"net"

	cachecash "github.com/kelleyk/go-cachecash"
	"github.com/kelleyk/go-cachecash/cache"
	"github.com/kelleyk/go-cachecash/catalog"
	"github.com/kelleyk/go-cachecash/ccmsg"
	"github.com/kelleyk/go-cachecash/provider"
	"github.com/kelleyk/go-cachecash/testutil"
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
	Upstream catalog.Upstream
}

type TestScenario struct {
	L *logrus.Logger

	Upstream catalog.Upstream

	Provider *provider.ContentProvider
	Catalog  catalog.ContentCatalog
	Escrow   *provider.Escrow

	Params *TestScenarioParams
	Obj    cachecash.ContentObject
	Caches []*cache.Cache
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

func GenerateTestScenario(l *logrus.Logger, params *TestScenarioParams) (*TestScenario, error) {
	var err error

	ts := &TestScenario{
		L:        l,
		Params:   params,
		Upstream: params.Upstream,
	}

	ts.L = logrus.New()
	ts.L.SetLevel(logrus.DebugLevel)

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

	// Add object to escrow.  XXX: Is this still necessary?
	if params.GenerateObject {
		escrow.Objects["/foo/bar"] = provider.EscrowObjectInfo{
			Object: ts.Obj,
			// N.B.: This becomes BundleParams.ObjectID
			// XXX: What's that used for?  Are it and BlockIdx used to uniquely identify blocks on the cache?  I think we're
			// using content addressing, aren't we?
			ID: 999,
		}
	}

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
			InnerMasterKey: innerMasterKey,
			OuterMasterKey: testutil.RandBytes(16),
		}
		if params.GenerateObject {
			if err := c.Storage.PutMetadata(42, 999, &ccmsg.ObjectMetadata{
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
				blockID := 10000 + uint64(j) // XXX: This is hardwired here and in the provider; need to replace with e.g. something content-based.
				intEscrowID := uint64(42)    // XXX: This is hardwired here and in the provider.
				if err := c.Storage.PutData(intEscrowID, blockID, data); err != nil {
					return nil, err
				}
			}
		}
		c.Escrows[escrow.ID()] = ce
		ts.Caches = append(ts.Caches, c)
	}

	return ts, nil
}
