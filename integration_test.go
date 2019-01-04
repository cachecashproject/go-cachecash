package cachecash_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"testing"

	cachecash "github.com/kelleyk/go-cachecash"
	"github.com/kelleyk/go-cachecash/cache"
	"github.com/kelleyk/go-cachecash/catalog"
	"github.com/kelleyk/go-cachecash/ccmsg"
	"github.com/kelleyk/go-cachecash/colocationpuzzle"
	"github.com/kelleyk/go-cachecash/common"
	"github.com/kelleyk/go-cachecash/provider"
	"github.com/kelleyk/go-cachecash/testutil"
	"github.com/kelleyk/go-cachecash/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/ed25519"
)

type IntegrationTestSuite struct {
	suite.Suite
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// func (suite *IntegrationTestSuite) SetupTest() {
// 	t := suite.T()
// }

func (suite *IntegrationTestSuite) TestTransfer() {
	t := suite.T()

	err := suite.testTransferC()
	assert.Nil(t, err)
}

func (suite *IntegrationTestSuite) testTransferC() error {
	t := suite.T()

	l := logrus.New()

	ctx := context.Background()

	// Create upstream with random data.
	upstream, err := catalog.NewMockUpstream(l)
	if err != nil {
		return errors.Wrap(err, "failed to create mock upstream")
	}
	upstream.AddRandomObject("/foo/bar", 16*1024*1024)

	// Create content catalog.
	cat, err := catalog.NewCatalog(l, upstream)
	if err != nil {
		return errors.Wrap(err, "failed to create catalog")
	}

	// Create a provider.
	_, providerPrivateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return errors.Wrap(err, "failed to generate provider keypair")
	}
	prov, err := provider.NewContentProvider(l, cat, providerPrivateKey)
	if err != nil {
		return err
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
		return err
	}
	if err := prov.AddEscrow(escrow); err != nil {
		return err
	}

	/*
		// Create a content object.
		obj, err := cachecash.RandomContentBuffer(16, 128*1024) // 16 blocks of 128 KiB
		if err != nil {
			return err
		}
		escrow.Objects["/foo/bar"] = provider.EscrowObjectInfo{
			Object: obj,
			ID:     999,
		}
	*/

	// Create caches that are participating in this escrow.
	var caches []*cache.Cache
	for i := 0; i < 4; i++ {
		public, private, err := ed25519.GenerateKey(nil)
		if err != nil {
			return errors.Wrap(err, "failed to generate cache keypair")
		}

		// XXX: generate master-key
		innerMasterKey := testutil.RandBytes(16)

		escrow.Caches = append(escrow.Caches, &provider.ParticipatingCache{
			InnerMasterKey: innerMasterKey,
			PublicKey:      public,
			Inetaddr:       net.ParseIP("127.0.0.1"),
			Port:           uint32(9000 + i),
		})

		c := &cache.Cache{
			Escrows: make(map[common.EscrowID]*cache.Escrow),
		}
		ce := &cache.Escrow{
			InnerMasterKey: innerMasterKey,
			OuterMasterKey: testutil.RandBytes(16),
			// Objects:        make(map[uint64]cachecash.ContentObject),
		}
		// ce.Objects[999] = obj
		c.Escrows[escrow.ID()] = ce
		caches = append(caches, c)
		_ = private
	}

	// Create a client keypair.
	clientPublicKey, clientPrivateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return errors.Wrap(err, "failed to generate client keypair")
	}

	// Create a client request.
	bundleReq := &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(clientPublicKey),
		Path:            "/foo/bar",
		RangeBegin:      0,
		RangeEnd:        4,
		SequenceNo:      1,
	}

	// Get a ticket bundle from the provider.
	bundle, err := prov.HandleContentRequest(context.Background(), bundleReq)
	if err != nil {
		return err
	}

	bundleJSON, err := json.Marshal(bundle)
	if err != nil {
		return err
	}
	fmt.Printf("ticket bundle:\n%s\n", bundleJSON)

	// Exchange request tickets with each cache.
	var doubleEncryptedBlocks [][]byte
	for i, cache := range caches {
		msg, err := bundle.BuildClientCacheRequest(bundle.TicketRequest[i])
		if err != nil {
			return errors.Wrap(err, "client failed to build request for cache")
		}
		resp, err := cache.HandleRequest(ctx, msg)
		if err != nil {
			return errors.Wrap(err, "cache failed to handle ticket request")
		}

		subMsg, ok := resp.Msg.(*ccmsg.ClientCacheResponse_DataResponse)
		if !ok {
			return errors.Wrap(err, "unexpected response type from request message")
		}
		doubleEncryptedBlocks = append(doubleEncryptedBlocks, subMsg.DataResponse.Data)
	}

	// Give each cache its L1 ticket; receive the outer session key for that cache's block in exchange.
	var outerSessionKeys []*ccmsg.BlockKey
	for i, cache := range caches {
		msg, err := bundle.BuildClientCacheRequest(bundle.TicketL1[i])
		if err != nil {
			return errors.Wrap(err, "client failed to build request for cache")
		}
		resp, err := cache.HandleRequest(ctx, msg)
		if err != nil {
			return errors.Wrap(err, "cache failed to handle ticket L1")
		}

		subMsg, ok := resp.Msg.(*ccmsg.ClientCacheResponse_L1Response)
		if !ok {
			return errors.Wrap(err, "unexpected response type from L1 message")
		}
		outerSessionKeys = append(outerSessionKeys, subMsg.L1Response.OuterKey)
	}

	// Decrypt once to reveal singly-encrypted blocks.
	var singleEncryptedBlocks [][]byte
	for i, ciphertext := range doubleEncryptedBlocks {
		plaintext, err := util.EncryptDataBlock(
			bundle.TicketRequest[i].BlockIdx,
			bundle.Remainder.RequestSequenceNo,
			outerSessionKeys[i].Key,
			ciphertext)
		if err != nil {
			return errors.Wrap(err, "failed to decrypt doubly-encrypted block")
		}

		singleEncryptedBlocks = append(singleEncryptedBlocks, plaintext)
	}

	// Solve colocation puzzle.
	pi := bundle.Remainder.PuzzleInfo
	secret, _, err := colocationpuzzle.Solve(colocationpuzzle.Parameters{
		Rounds:      pi.Rounds,
		StartOffset: uint32(pi.StartOffset),
		StartRange:  uint32(pi.StartRange),
	}, singleEncryptedBlocks, pi.Goal)
	if err != nil {
		return err
	}

	// Decrypt L2 ticket.
	ticketL2, err := common.DecryptTicketL2(secret, bundle.EncryptedTicketL2)
	if err != nil {
		return err
	}

	// Give L2 tickets to caches.
	for _, cache := range caches {
		msg, err := bundle.BuildClientCacheRequest(&ccmsg.TicketL2Info{
			EncryptedTicketL2: bundle.EncryptedTicketL2,
			PuzzleSecret:      secret,
		})
		if err != nil {
			return errors.Wrap(err, "client failed to build request for cache")
		}
		resp, err := cache.HandleRequest(ctx, msg)
		if err != nil {
			return errors.Wrap(err, "cache failed to handle ticket L2")
		}

		subMsg, ok := resp.Msg.(*ccmsg.ClientCacheResponse_L2Response)
		if !ok {
			return errors.Wrap(err, "unexpected response type from L2 message")

		}
		// TODO: Check that response is successful, at least?
		_ = subMsg
	}

	// Decrypt singly-encrypted data blocks to reveal plaintext data.
	var plaintextBlocks [][]byte
	for i, ciphertext := range doubleEncryptedBlocks {
		plaintext, err := util.EncryptDataBlock(
			bundle.TicketRequest[i].BlockIdx,
			bundle.Remainder.RequestSequenceNo,
			ticketL2.InnerSessionKey[i].Key,
			ciphertext)
		if err != nil {
			return errors.Wrap(err, "failed to decrypt singly-encrypted block")
		}
		plaintextBlocks = append(plaintextBlocks, plaintext)
	}

	// TODO: XXX: Re-enable this verification step.
	/*
		// Verify that the plaintext data the client has received matches what the provider and caches have.
		for i, b := range plaintextBlocks {
			expected, err := upstream.GetBlock("/foo/bar", uint(bundle.TicketRequest[i].BlockIdx))
			if err != nil {
				return errors.Wrap(err, "failed to get expected block contents")
			}
			if !bytes.Equal(expected, b) {
				return errors.New("plaintext data received by client does not match expected value")
			}
		}
	*/

	_ = t
	_ = clientPrivateKey
	return nil
}
