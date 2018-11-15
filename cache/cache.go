package cache

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"

	cachecash "github.com/kelleyk/go-cachecash"
	"github.com/kelleyk/go-cachecash/batchsignature"
	"github.com/kelleyk/go-cachecash/ccmsg"
	"github.com/kelleyk/go-cachecash/util"
	"github.com/pkg/errors"
)

type Escrow struct {
	InnerMasterKey []byte // XXX: Shared with provider?
	OuterMasterKey []byte

	// XXX: Temporary:
	Objects map[uint64]cachecash.ContentObject
}

func (e *Escrow) Active() bool {
	return true
}

type Cache struct {
	Escrows map[ccmsg.EscrowID]*Escrow
}

func (c *Cache) getDataBlock(escrowID *ccmsg.EscrowID, objectID uint64, blockIdx uint64) ([]byte, error) {
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

func (c *Cache) getEscrow(escrowID *ccmsg.EscrowID) (*Escrow, error) {
	escrow, ok := c.Escrows[*escrowID]
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

func (c *Cache) HandleRequest(req *ccmsg.ClientCacheRequest) (*ccmsg.ClientCacheResponse, error) {
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
		return c.handleDataRequest(escrow, req)
	case *ccmsg.ClientCacheRequest_TicketL1:
		return c.handleTicketL1Request(escrow, req)
	case *ccmsg.ClientCacheRequest_TicketL2:
		return c.handleTicketL2Request(escrow, req)
	default:
		return nil, errors.New("unexpected ticket type in client request")
	}
}

func (c *Cache) handleDataRequest(escrow *Escrow, req *ccmsg.ClientCacheRequest) (*ccmsg.ClientCacheResponse, error) {
	// XXX: Refactoring dust!
	ticketRequest := req.Ticket.(*ccmsg.ClientCacheRequest_TicketRequest).TicketRequest

	// If we don't have the block, ask the CP how to get it.
	// XXX: This will need more arguments.
	dataBlock, err := c.getDataBlock(req.BundleRemainder.EscrowId, req.BundleRemainder.ObjectId, ticketRequest.BlockIdx)
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

func (c *Cache) handleTicketL1Request(escrow *Escrow, req *ccmsg.ClientCacheRequest) (*ccmsg.ClientCacheResponse, error) {
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

func (c *Cache) handleTicketL2Request(escrow *Escrow, req *ccmsg.ClientCacheRequest) (*ccmsg.ClientCacheResponse, error) {
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
