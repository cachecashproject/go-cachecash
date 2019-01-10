package main

import (
	"net"

	cachecash "github.com/kelleyk/go-cachecash"
	"github.com/kelleyk/go-cachecash/cache"
	"github.com/kelleyk/go-cachecash/ccmsg"
	"github.com/kelleyk/go-cachecash/provider"
	"github.com/kelleyk/go-cachecash/testutil"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ed25519"
)

// TEMP: Cribbed from `integration_test.go`.
func makeProvider() (*provider.ContentProvider, error) {
	l := logrus.New()

	// Create a provider.
	_, providerPrivateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate provider keypair")
	}
	// XXX: The addition of the content catalog has broken this program.  Once things stabilize, we should rebuild this
	// application based on the relevant portions of `testserverd`.
	prov, err := provider.NewContentProvider(l, nil, providerPrivateKey)
	if err != nil {
		return nil, err
	}

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

	// Create a content object: 16 blocks of 128 KiB.
	blockSize := uint32(128 * 1024)
	blockCount := uint32(16)
	objectSize := blockSize * blockCount
	obj, err := cachecash.RandomContentBuffer(blockCount, blockSize)
	if err != nil {
		return nil, err
	}
	escrow.Objects["/foo/bar"] = provider.EscrowObjectInfo{
		Object: obj,
		ID:     999,
	}

	// Create caches that are participating in this escrow.
	var caches []*cache.Cache
	for i := 0; i < 4; i++ {
		public, private, err := ed25519.GenerateKey(nil)
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

		c, err := cache.NewCache(l)
		if err != nil {
			return nil, err
		}
		ce := &cache.Escrow{
			InnerMasterKey: innerMasterKey,
			OuterMasterKey: testutil.RandBytes(16),
		}
		if err := c.Storage.PutMetadata(42, 999, &ccmsg.ObjectMetadata{
			BlockSize:  uint64(blockSize),
			ObjectSize: uint64(objectSize),
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
		caches = append(caches, c)
		_ = private
	}

	return prov, nil
}
