package cache

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/kelleyk/go-cachecash/batchsignature"
	"github.com/kelleyk/go-cachecash/ccmsg"
	"github.com/kelleyk/go-cachecash/common"
	"github.com/kelleyk/go-cachecash/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Escrow struct {
	InnerMasterKey []byte // XXX: Shared with provider?
	OuterMasterKey []byte

	ProviderCacheServiceAddr string
}

func (e *Escrow) Active() bool {
	return true
}

type Cache struct {
	l *logrus.Logger

	Escrows map[common.EscrowID]*Escrow

	Storage *CacheStorage
}

func NewCache(l *logrus.Logger) (*Cache, error) {
	s, err := NewCacheStorage(l)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cache storage")
	}

	return &Cache{
		l:       l,
		Escrows: make(map[common.EscrowID]*Escrow),
		Storage: s,
	}, nil
}

func (c *Cache) getDataBlock(ctx context.Context, escrowID *ccmsg.EscrowID, objectID uint64, blockIdx, blockID uint64) ([]byte, error) {
	// XXX: Need to decide on a type and replace this with a real escrow ID.
	var intEscrowID uint64

	// Can we satisfy the request out of cache?

	data, err := c.Storage.GetData(intEscrowID, blockID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get data block")
	}

	if data == nil {
		escrow, err := c.getEscrow(escrowID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get escrow")
		}

		// XXX: No transport security!
		// XXX: Should not create a new connection for each attempt.
		c.l.Info("dialing provider's cache-facing service")
		conn, err := grpc.Dial(escrow.ProviderCacheServiceAddr, grpc.WithInsecure())
		if err != nil {
			return nil, errors.Wrap(err, "failed to dial")
		}
		grpcClient := ccmsg.NewCacheProviderClient(conn)

		// First, contact the provider's cache-facing service and ask where to fetch this block from.
		c.l.Info("asking provider for cache-miss info")
		resp, err := grpcClient.CacheMiss(ctx, &ccmsg.CacheMissRequest{
			ObjectId:   objectID,
			RangeBegin: blockIdx,
			RangeEnd:   blockIdx + 1,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch upstream info from provider")
		}

		// Fetch it.
		c.l.Info("fetching data from HTTP upstream")
		source := resp.Source.(*ccmsg.CacheMissResponse_Http).Http
		req, err := http.NewRequest("GET", source.Url, nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed to build HTTP request")
		}
		req.Header.Set("Range", fmt.Sprintf("%v-%v", source.RangeBegin, source.RangeEnd))
		httpClient := &http.Client{}
		httpResp, err := httpClient.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch object from HTTP upstream")
		}
		defer httpResp.Body.Close()
		data, err = ioutil.ReadAll(httpResp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read body of object from HTTP upstream")
		}

		// Insert it into the cache.
		c.l.Info("inserting data into cache")
		if err := c.Storage.PutMetadata(intEscrowID, blockID, resp.Metadata); err != nil {
			return nil, errors.Wrap(err, "failed to store metadata in cache")
		}
		if err := c.Storage.PutData(intEscrowID, blockID, data); err != nil {
			return nil, errors.Wrap(err, "failed to store data in cache")
		}
	}

	c.l.Info("cache returns data")
	return data, nil
}

/*
func (c *Cache) getDataBlock(ctx context.Context, escrowID *ccmsg.EscrowID, objectID uint64, blockIdx uint64) ([]byte, error) {
	escrow, err := c.getEscrow(escrowID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get escrow")
	}

	obj, ok := escrow.Objects[objectID]
	if !ok {
		return nil, errors.New("no such object")
	}

	return obj.GetBlock(uint32(blockIdx)) // XXX: Fix typing.
}
*/

func (c *Cache) getEscrow(escrowID *ccmsg.EscrowID) (*Escrow, error) {
	// XXX: Don't just ignore the escrow ID!
	var escrowIDbuf common.EscrowID

	escrow, ok := c.Escrows[escrowIDbuf]
	if !ok {
		return nil, errors.New("no such escrow")
	}
	return escrow, nil
}

func (c *Cache) storeTicketL1(req *ccmsg.ClientCacheRequest) error {
	return nil
}

func (c *Cache) storeTicketL2(req *ccmsg.ClientCacheRequest) error {
	return nil
}

func (c *Cache) HandleRequest(ctx context.Context, req *ccmsg.ClientCacheRequest) (*ccmsg.ClientCacheResponse, error) {
	// Make sure that we're participating in this escrow.
	escrow, err := c.getEscrow(req.BundleRemainder.EscrowId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch escrow information")
	}
	if !escrow.Active() {
		return nil, errors.New("not actively participating in given escrow")
	}

	// Verify that the request that the client is presenting is covered by a valid signature.
	if err := VerifyRequest(req); err != nil {
		return nil, errors.Wrap(err, "failed to verify batch signature")
	}

	// Verify that the signing key is authorized to sign tickets for this escrow.
	// TODO:
	// - Verify that the batch-signer is the subject of the certificate.
	// - Verify that the certificate applies to this escrow.
	// - Verify that the certificate is signed by the escrow private key.

	switch req.Ticket.(type) {
	case *ccmsg.ClientCacheRequest_TicketRequest:
		return c.handleDataRequest(ctx, escrow, req)
	case *ccmsg.ClientCacheRequest_TicketL1:
		return c.handleTicketL1Request(ctx, escrow, req)
	case *ccmsg.ClientCacheRequest_TicketL2:
		return c.handleTicketL2Request(ctx, escrow, req)
	default:
		return nil, errors.New("unexpected ticket type in client request")
	}
}

func (c *Cache) handleDataRequest(ctx context.Context, escrow *Escrow, req *ccmsg.ClientCacheRequest) (*ccmsg.ClientCacheResponse, error) {
	// XXX: Refactoring dust!
	ticketRequest := req.Ticket.(*ccmsg.ClientCacheRequest_TicketRequest).TicketRequest

	// If we don't have the block, ask the CP how to get it.
	dataBlock, err := c.getDataBlock(ctx, req.BundleRemainder.EscrowId, req.BundleRemainder.ObjectId, ticketRequest.BlockIdx, ticketRequest.BlockId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get data block")
	}

	for _, masterKey := range [][]byte{escrow.InnerMasterKey, escrow.OuterMasterKey} {
		// XXX: Fix typing.
		seqNo := uint32(req.BundleRemainder.RequestSequenceNo)

		// Generate our session key frmo the master key.
		key, err := util.KeyedPRF(req.BundleRemainder.ClientPublicKey.PublicKey, seqNo, masterKey)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate key")
		}

		// Note that we use the key we've just generated (and not the master key) to generate the IV.
		iv, err := util.KeyedPRF(util.Uint64ToLE(ticketRequest.BlockIdx), seqNo, key)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate IV")
		}

		// Set up our cipher.
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, errors.Wrap(err, "failed to construct block cipher")
		}
		stream := cipher.NewCTR(block, iv)

		// Encrypt the data.
		ciphertext := make([]byte, len(dataBlock))
		stream.XORKeyStream(ciphertext, dataBlock)
		dataBlock = ciphertext
	}

	// Done!
	return &ccmsg.ClientCacheResponse{
		RequestSequenceNo: req.SequenceNo,
		Msg: &ccmsg.ClientCacheResponse_DataResponse{
			DataResponse: &ccmsg.ClientCacheResponseData{
				Data: dataBlock,
			},
		},
	}, nil
}

func (c *Cache) handleTicketL1Request(ctx context.Context, escrow *Escrow, req *ccmsg.ClientCacheRequest) (*ccmsg.ClientCacheResponse, error) {
	if err := c.storeTicketL1(req); err != nil {
		return nil, errors.Wrap(err, "failed to store ticket")
	}

	// XXX: Fix typing.
	seqNo := uint32(req.BundleRemainder.RequestSequenceNo)

	outerSessionKey, err := util.KeyedPRF(req.BundleRemainder.ClientPublicKey.PublicKey, seqNo, escrow.OuterMasterKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate key")
	}

	return &ccmsg.ClientCacheResponse{
		RequestSequenceNo: req.SequenceNo,
		Msg: &ccmsg.ClientCacheResponse_L1Response{
			L1Response: &ccmsg.ClientCacheResponseL1{
				OuterKey: &ccmsg.BlockKey{Key: outerSessionKey},
			},
		},
	}, nil
}

func (c *Cache) handleTicketL2Request(ctx context.Context, escrow *Escrow, req *ccmsg.ClientCacheRequest) (*ccmsg.ClientCacheResponse, error) {
	if err := c.storeTicketL2(req); err != nil {
		return nil, errors.Wrap(err, "failed to store ticket")
	}

	// XXX: Other than indicating success, we don't need to return anything here, do we?
	return &ccmsg.ClientCacheResponse{
		RequestSequenceNo: req.SequenceNo,
	}, nil
}

// TODO: This should go somewhere non-cache-specific; initially, it was in ccmsg, but that created an import cycle.
//
// N.B.: This does *not* consider anything about who has generated the batch signature; at present, it just checks that
// the data-carrying parts of the message are (indirectly) covered by that signature.
func VerifyRequest(m *ccmsg.ClientCacheRequest) error {
	switch subMsg := m.Ticket.(type) {
	case *ccmsg.ClientCacheRequest_TicketRequest:
		if !m.TicketBundleSubdigests.ContainsTicketRequestDigest(subMsg.TicketRequest.CanonicalDigest()) {
			return errors.New("ticket request digest not found")
		}
	case *ccmsg.ClientCacheRequest_TicketL1:
		if !m.TicketBundleSubdigests.ContainsTicketL1Digest(subMsg.TicketL1.CanonicalDigest()) {
			return errors.New("ticket L1 digest not found")
		}
	case *ccmsg.ClientCacheRequest_TicketL2:
		if !bytes.Equal(m.TicketBundleSubdigests.EncryptedTicketL2Digest, subMsg.TicketL2.EncryptedTicketL2Digest()) {
			return errors.New("encrypted ticket L2 digest mismatch")
		}
	default:
		return errors.New("unexpected ticket type in client request")
	}

	if !bytes.Equal(m.TicketBundleSubdigests.RemainderDigest, m.BundleRemainder.CanonicalDigest()) {
		return errors.New("ticket bundle remainder digest mismatch")
	}

	ok, err := batchsignature.Verify(m.TicketBundleSubdigests.CanonicalDigest(), m.BundleSig)
	if err != nil {
		return errors.Wrap(err, "failed to verify batch signature")
	}
	if !ok {
		return errors.New("batch signature invalid")
	}

	return nil
}
