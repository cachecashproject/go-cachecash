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

// XXX: Replace this with something that uses `testdatagen`.
func (suite *IntegrationTestSuite) testTransferC() error {
	t := suite.T()

	l := logrus.New()

	ctx := context.Background()

	scen, err := testdatagen.GenerateTestScenario(l, &testdatagen.TestScenarioParams{
		BlockSize:      128 * 1024,
		ObjectSize:     128 * 1024 * 16,
		MockUpstream:   true,
		GenerateObject: true,
	})
	if err != nil {
		return err
	}

	prov := scen.Provider
	caches := scen.Caches

	// Pull information about the object into the provider's catalog so that no upstream fetches are necessary.
	// l.Infof("pulling metadata into provider catalog: start")
	// // scen.Catalog.
	// l.Infof("pulling metadata into provider catalog: done")
	l.Infof("pulling data into provider catalog: start")
	for i := 0; i < int(scen.BlockCount()); i++ {
		_, err = scen.Catalog.GetData(ctx, &ccmsg.ContentRequest{
			Path:       "/foo/bar",
			RangeBegin: uint64(i) * scen.Params.BlockSize,
			RangeEnd:   uint64(i+1) * scen.Params.BlockSize,
		})
		if err != nil {
			return err
		}
	}
	l.Infof("pulling data into provider catalog: done")

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
