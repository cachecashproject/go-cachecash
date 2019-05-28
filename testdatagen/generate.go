package testdatagen

import (
	"crypto/sha512"
	"fmt"
	"math"
	"net"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cachecashproject/go-cachecash/cache"
	cacheModels "github.com/cachecashproject/go-cachecash/cache/models"
	"github.com/cachecashproject/go-cachecash/catalog"
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/keypair"
	"github.com/cachecashproject/go-cachecash/publisher"
	publisherModels "github.com/cachecashproject/go-cachecash/publisher/models"
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

	PublisherCacheAddr string
}

type TestScenario struct {
	L *logrus.Logger

	Upstream catalog.Upstream

	Publisher           *publisher.ContentPublisher
	PublisherPrivateKey ed25519.PrivateKey
	Catalog             catalog.ContentCatalog
	Escrow              *publisher.Escrow
	EscrowID            common.EscrowID

	Params       *TestScenarioParams
	DataBlocks   [][]byte
	ObjectID     common.ObjectID
	Caches       []*cache.Cache
	CacheConfigs []*cache.ConfigFile
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
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create stub database connection")
	}

	innerMasterKeys := [][]byte{}
	cachePublicKeys := []ed25519.PublicKey{}

	for i := 0; i < 4; i++ {
		public, _, err := ed25519.GenerateKey(nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate cache keypair")
		}
		cachePublicKeys = append(cachePublicKeys, public)
		innerMasterKeys = append(innerMasterKeys, testutil.RandBytes(16))
	}

	rows := sqlmock.NewRows([]string{"id", "escrow_id", "cache_id", "inner_master_key"}).
		AddRow(1, 0, 123, innerMasterKeys[0]).
		AddRow(2, 0, 124, innerMasterKeys[1]).
		AddRow(3, 0, 125, innerMasterKeys[2]).
		AddRow(4, 0, 126, innerMasterKeys[3])
	mock.ExpectQuery("^SELECT \\* FROM \"escrow_caches\" WHERE \\(escrow_id = \\$1\\);").
		WithArgs(0).
		WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "public_key", "inetaddr", "port"}).
		AddRow(123, cachePublicKeys[0], net.ParseIP("127.0.0.1"), 9000).
		AddRow(124, cachePublicKeys[1], net.ParseIP("127.0.0.1"), 9001).
		AddRow(125, cachePublicKeys[2], net.ParseIP("127.0.0.1"), 9002).
		AddRow(126, cachePublicKeys[3], net.ParseIP("127.0.0.1"), 9003)
	mock.ExpectQuery("^SELECT \\* FROM \"cache\" WHERE \\(\"id\" IN \\(\\$1,\\$2,\\$3,\\$4\\)\\);").
		WithArgs(123, 124, 125, 126).
		WillReturnRows(rows)

	prov, err := publisher.NewContentPublisher(ts.L, db, "", cat, publisherPrivateKey)
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
		ts.Escrow.Caches = append(ts.Escrow.Caches, &publisher.ParticipatingCache{
			InnerMasterKey: innerMasterKeys[i],
			Cache: publisherModels.Cache{
				PublicKey: cachePublicKeys[i],
				Inetaddr:  net.ParseIP("127.0.0.1"),
				Port:      uint32(9000 + i),
			},
		})

		cacheConfig := &cache.ConfigFile{
			BadgerDirectory: fmt.Sprintf("./unittestdata/cache-%d/", i),
		}
		keypair := &keypair.KeyPair{
			PublicKey: cachePublicKeys[i],
		}
		ts.CacheConfigs = append(ts.CacheConfigs, cacheConfig)

		c, err := cache.NewCache(ts.L, nil, cacheConfig, keypair)
		if err != nil {
			return nil, err
		}

		if params.PublisherCacheAddr == "" {
			params.PublisherCacheAddr = "localhost:8082"
		}

		ce := &cache.Escrow{
			Inner: cacheModels.Escrow{
				InnerMasterKey:     innerMasterKeys[i],
				OuterMasterKey:     testutil.RandBytes(16),
				Slots:              uint64(2500 + i),
				PublisherCacheAddr: params.PublisherCacheAddr,
			},
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
		c.Escrows[escrow.Inner.Txid] = ce
		ts.Caches = append(ts.Caches, c)
	}

	return ts, nil
}
