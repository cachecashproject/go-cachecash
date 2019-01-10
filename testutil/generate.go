package testutil

import (
	"math"
	"net"

	cachecash "github.com/kelleyk/go-cachecash"
	"github.com/kelleyk/go-cachecash/cache"
	"github.com/kelleyk/go-cachecash/catalog"
	"github.com/kelleyk/go-cachecash/ccmsg"
	"github.com/kelleyk/go-cachecash/provider"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ed25519"
)

type TestScenarioParams struct {
	BlockSize  uint64
	ObjectSize uint64
}

type TestScenario struct {
	L *logrus.Logger

	Provider *provider.ContentProvider
	Escrow   *provider.Escrow

	Params *TestScenarioParams
	Obj    cachecash.ContentObject
	Caches []*cache.Cache
}

func (ts *TestScenario) BlockCount() uint64 {
	return uint64(math.Ceil(float64(ts.Params.ObjectSize) / float64(ts.Params.BlockSize)))
}

func GenerateTestScenario(l *logrus.Logger, params *TestScenarioParams) (*TestScenario, error) {
	ts := &TestScenario{
		L:      l,
		Params: params,
	}

	ts.L = logrus.New()
	ts.L.SetLevel(logrus.DebugLevel)

	// Create content catalog.
	upstream, err := catalog.NewHTTPUpstream(ts.L, "http://localhost:8081")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create HTTP upstream")
	}
	cat, err := catalog.NewCatalog(ts.L, upstream)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create catalog")
	}

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

	// Create a content object.
	obj, err := cachecash.RandomContentBuffer(uint32(ts.BlockCount()), uint32(ts.Params.BlockSize))
	if err != nil {
		return nil, err
	}
	escrow.Objects["/foo/bar"] = provider.EscrowObjectInfo{
		Object: obj,
		// N.B.: This becomes BundleParams.ObjectID
		// XXX: What's that used for?  Are it and BlockIdx used to uniquely identify blocks on the cache?  I think we're
		// using content addressing, aren't we?
		ID: 999,
	}
	ts.Obj = obj

	// Create caches that are participating in this escrow.
	for i := 0; i < 4; i++ {
		public, _, err := ed25519.GenerateKey(nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate cache keypair")
		}

		// XXX: generate master-key
		innerMasterKey := RandBytes(16)

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
			OuterMasterKey: RandBytes(16),
		}
		if err := c.Storage.PutMetadata(42, 999, &ccmsg.ObjectMetadata{
			BlockSize:  ts.Params.BlockSize,
			ObjectSize: ts.Params.ObjectSize,
		}); err != nil {
			return nil, err
		}
		for j := 0; j < obj.BlockCount(); j++ {
			data, err := obj.GetBlock(uint32(j))
			if err != nil {
				return nil, err
			}
			if err := c.Storage.PutData(42, uint64(j), data); err != nil {
				return nil, err
			}
		}
		c.Escrows[escrow.ID()] = ce
		ts.Caches = append(ts.Caches, c)
	}

	return ts, nil
}
