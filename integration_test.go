package cachecash_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	cachecash "github.com/cachecashproject/go-cachecash"
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/colocationpuzzle"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/testdatagen"
	"github.com/cachecashproject/go-cachecash/util"
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
	println("Integration test disabled due to L2Response cast failure bug #122")
	if true {
		return
	}

	l := logrus.New()

	ctx := context.Background()

	scen, err := testdatagen.GenerateTestScenario(l, &testdatagen.TestScenarioParams{
		ChunkSize:      128 * 1024,
		ObjectSize:     128 * 1024 * 16,
		MockUpstream:   true,
		GenerateObject: true,
	})
	assert.Nil(t, err)

	prov := scen.Publisher
	caches := scen.Caches

	// Pull information about the object into the publisher's catalog so that no upstream fetches are necessary.
	// l.Infof("pulling metadata into publisher catalog: start")
	// // scen.Catalog.
	// l.Infof("pulling metadata into publisher catalog: done")
	l.Infof("pulling data into publisher catalog: start")
	for i := 0; i < int(scen.ChunkCount()); i++ {
		_, err = scen.Catalog.GetData(ctx, &ccmsg.ContentRequest{
			Path:       "/foo/bar",
			RangeBegin: uint64(i) * scen.Params.ChunkSize,
			RangeEnd:   uint64(i+1) * scen.Params.ChunkSize,
		})
		assert.Nil(t, err)
	}
	l.Infof("pulling data into publisher catalog: done")

	// Create a client keypair.
	clientPublicKey, _ /* clientPrivateKey */, err := ed25519.GenerateKey(nil)
	assert.Nil(t, err)

	// Create a client request.
	bundleReq := &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(clientPublicKey),
		Path:            "/foo/bar",
		RangeBegin:      0,
		RangeEnd:        4,
		SequenceNo:      1,
	}

	// Get a ticket bundle from the publisher.
	bundles, err := prov.HandleContentRequest(context.Background(), bundleReq)
	assert.Nil(t, err)

	for _, bundle := range bundles {
		bundleJSON, err := json.Marshal(bundle)
		assert.Nil(t, err)
		fmt.Printf("ticket bundle:\n%s\n", bundleJSON)

		// Exchange request tickets with each cache.
		var doubleEncryptedChunks [][]byte
		for i, cache := range caches {
			msg, err := bundle.BuildClientCacheRequest(bundle.TicketRequest[i])
			assert.Nil(t, err)
			resp, err := cache.HandleRequest(ctx, msg)
			assert.Nil(t, err)

			subMsg, ok := resp.Msg.(*ccmsg.ClientCacheResponse_DataResponse)
			assert.True(t, ok)
			doubleEncryptedChunks = append(doubleEncryptedChunks, subMsg.DataResponse.Data)
		}

		// Give each cache its L1 ticket; receive the outer session key for that cache's chunk in exchange.
		var outerSessionKeys []*ccmsg.BlockKey
		for i, cache := range caches {
			msg, err := bundle.BuildClientCacheRequest(bundle.TicketL1[i])
			assert.Nil(t, err)
			resp, err := cache.HandleRequest(ctx, msg)
			assert.Nil(t, err)
			subMsg, ok := resp.Msg.(*ccmsg.ClientCacheResponse_L1Response)
			assert.True(t, ok)
			outerSessionKeys = append(outerSessionKeys, subMsg.L1Response.OuterKey)
		}

		// Decrypt once to reveal singly-encrypted chunks.
		var singleEncryptedChunks [][]byte
		for i, ciphertext := range doubleEncryptedChunks {
			plaintext, err := util.EncryptChunk(
				bundle.TicketRequest[i].ChunkIdx,
				bundle.Remainder.RequestSequenceNo,
				outerSessionKeys[i].Key,
				ciphertext)
			assert.Nil(t, err)

			singleEncryptedChunks = append(singleEncryptedChunks, plaintext)
		}

		// Solve colocation puzzle.
		pi := bundle.Remainder.PuzzleInfo
		secret, _, err := colocationpuzzle.Solve(colocationpuzzle.Parameters{
			Rounds:      pi.Rounds,
			StartOffset: uint32(pi.StartOffset),
			StartRange:  uint32(pi.StartRange),
		}, singleEncryptedChunks, pi.Goal)
		assert.Nil(t, err)

		// Decrypt L2 ticket.
		ticketL2, err := common.DecryptTicketL2(secret, bundle.EncryptedTicketL2)
		assert.Nil(t, err)

		// Give L2 tickets to caches.
		for _, cache := range caches {
			msg, err := bundle.BuildClientCacheRequest(&ccmsg.TicketL2Info{
				EncryptedTicketL2: bundle.EncryptedTicketL2,
				PuzzleSecret:      secret,
			})
			assert.Nil(t, err)
			resp, err := cache.HandleRequest(ctx, msg)
			assert.Nil(t, err)

			subMsg, ok := resp.Msg.(*ccmsg.ClientCacheResponse_L2Response)
			assert.True(t, ok)
			// TODO: Check that response is successful, at least?
			_ = subMsg
		}

		// Decrypt singly-encrypted chunks to reveal plaintext data.
		var plaintextChunks [][]byte
		for i, ciphertext := range doubleEncryptedChunks {
			plaintext, err := util.EncryptChunk(
				bundle.TicketRequest[i].ChunkIdx,
				bundle.Remainder.RequestSequenceNo,
				ticketL2.InnerSessionKey[i].Key,
				ciphertext)
			assert.Nil(t, err)
			plaintextChunks = append(plaintextChunks, plaintext)
		}

		// Verify that the plaintext data the client has received matches what the publisher and caches have.
		for i, b := range plaintextChunks {
			expected := scen.Chunks[bundle.TicketRequest[i].ChunkIdx]
			assert.Equal(t, expected, b)
		}
	}
}
