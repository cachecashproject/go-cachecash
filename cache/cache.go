package cache

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cachecashproject/go-cachecash/batchsignature"
	"github.com/cachecashproject/go-cachecash/cache/models"
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/keypair"
	"github.com/cachecashproject/go-cachecash/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"golang.org/x/crypto/ed25519"
	"google.golang.org/grpc/peer"
)

/*
Notes & TODOs:
- As-is, the cache does not actually use the metadata that it stores.  Why did we want the metadata to be provided?
- No support yet for multiple-chunk fetches.  (When implementing this, need to make sure that we're using the chunk
  ID(s) returned by the publisher.
*/

type Escrow struct {
	Inner models.Escrow
}

func (e *Escrow) Active() bool {
	return true
}

func (e *Escrow) InnerMasterKey() []byte {
	return e.Inner.InnerMasterKey
}

func (e *Escrow) OuterMasterKey() []byte {
	return e.Inner.OuterMasterKey
}

func (e *Escrow) Slots() uint64 {
	return e.Inner.Slots
}

func (e *Escrow) PublisherCacheAddr() string {
	return e.Inner.PublisherCacheAddr
}

type Cache struct {
	l  *logrus.Logger
	db *sql.DB

	PublicKey   ed25519.PublicKey
	Escrows     map[common.EscrowID]*Escrow
	Storage     *CacheStorage
	StoragePath string
	StartupTime time.Time
}

func NewCache(l *logrus.Logger, db *sql.DB, cf *ConfigFile, kp *keypair.KeyPair) (*Cache, error) {
	l.WithFields(logrus.Fields{
		"badger_storage": cf.BadgerDirectory,
	}).Info("setting up storage")
	s, err := NewCacheStorage(l, cf.BadgerDirectory)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cache storage")
	}

	return &Cache{
		l:           l,
		db:          db,
		PublicKey:   kp.PublicKey,
		Escrows:     make(map[common.EscrowID]*Escrow),
		Storage:     s,
		StoragePath: cf.BadgerDirectory,
		StartupTime: time.Now(),
	}, nil
}

func (c *Cache) Close() error {
	return c.Storage.Close()
}

func (c *Cache) LoadFromDatabase(ctx context.Context) (int, error) {
	escrows, err := models.Escrows().All(ctx, c.db)

	if err != nil {
		return 0, errors.Wrap(err, "failed to query Escrows")
	}

	for _, e := range escrows {
		c.Escrows[e.Txid] = &Escrow{
			Inner: *e,
		}
	}

	return len(escrows), nil
}

func (c *Cache) AddEscrowToDatabase(ctx context.Context, escrow *Escrow) error {
	return escrow.Inner.Insert(ctx, c.db, boil.Infer())
}

func (c *Cache) getChunk(ctx context.Context, escrowID common.EscrowID, objectID common.ObjectID, chunkIdx uint64,
	chunkID common.ChunkID) ([]byte, error) {

	// Can we satisfy the request out of cache?
	data, err := c.Storage.GetData(escrowID, chunkID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get chunk from storage")
	}

	if data == nil {
		escrow, err := c.getEscrow(escrowID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get escrow")
		}

		// XXX: Should not create a new connection for each attempt.
		c.l.Info("dialing publisher's cache-facing service: ", escrow.PublisherCacheAddr())
		conn, err := common.GRPCDial(escrow.PublisherCacheAddr())
		if err != nil {
			return nil, errors.Wrap(err, "failed to dial")
		}
		grpcClient := ccmsg.NewCachePublisherClient(conn)

		// First, contact the publisher's cache-facing service and ask where to fetch this chunk from.
		c.l.Info("asking publisher for cache-miss info")
		resp, err := grpcClient.CacheMiss(ctx, &ccmsg.CacheMissRequest{
			ObjectId:   objectID[:],
			RangeBegin: chunkIdx,
			RangeEnd:   chunkIdx + 1,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch upstream info from publisher")
		}

		for _, chunk := range resp.Chunks {
			c.l.WithFields(logrus.Fields{
				"slot_idx":    chunk.SlotIdx,
				"chunk_id":    chunkID,
				"object_size": resp.Metadata.ObjectSize,
				"chunk_size":  resp.Metadata.ChunkSize,
			}).Debug("cache-miss response")

			if chunk.SlotIdx >= escrow.Slots() {
				return nil, errors.New("slot number out of range")
			}

			// Setup data retrieval
			switch source := chunk.Source.(type) {
			case *ccmsg.Chunk_Http:
				data, err = c.getChunkHTTP(source)
				if err != nil {
					return nil, err
				}
			case *ccmsg.Chunk_Inline:
				data = bytes.Join(source.Inline.Chunk, nil)
			default:
				return nil, fmt.Errorf("unexpected chunk source type: %T", chunk.Source)
			}

			// update LogicalCacheMapping
			if err = c.updateLogicalCacheMapping(ctx, chunk, escrowID, chunkID); err != nil {
				return nil, errors.Wrap(err, "failed to update lcm")
			}

			// Insert it into the cache.
			c.l.Info("inserting data into cache")
			if err := c.Storage.PutMetadata(escrowID, objectID, resp.Metadata); err != nil {
				return nil, errors.Wrap(err, "failed to store metadata in cache")
			}
			if err := c.Storage.PutData(escrowID, chunkID, data); err != nil {
				return nil, errors.Wrap(err, "failed to store data in cache")
			}
		}
	} else {
		c.l.Debug("using data from cache")
	}

	c.l.Info("cache returns data")
	return data, nil
}

func (c *Cache) getChunkHTTP(source *ccmsg.Chunk_Http) ([]byte, error) {
	c.l.Info("sending request")
	req, err := http.NewRequest("GET", source.Http.Url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build HTTP request")
	}
	// N.B.: HTTP ranges are inclusive; our ranges are [inclusive, exclusive).
	req.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", source.Http.RangeBegin, source.Http.RangeEnd-1))

	// Make request to upstream.
	c.l.Infof("fetching data from HTTP upstream; req=%v", req)
	httpClient := &http.Client{}
	httpResp, err := httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch object from HTTP upstream")
	}
	defer func() {
		_ = httpResp.Body.Close()
	}()

	// Interpret response.
	switch {
	case httpResp.StatusCode == http.StatusOK:
	case httpResp.StatusCode == http.StatusPartialContent:
	default:
		return nil, fmt.Errorf("unexpected status from HTTP upstream: %v", httpResp.Status)
	}
	data, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read body of object from HTTP upstream")
	}

	c.l.WithFields(logrus.Fields{
		"len": len(data),
	}).Info("got response from HTTP upstream")

	if httpResp.StatusCode == http.StatusOK {
		c.l.Warn("server doesn't support range requests, slicing range from full response")

		// XXX: make sure RangeEnd doesn't go beyond the file length.
		// This should not happen with a correctly calculated RangeEnd
		rangeEnd := source.Http.RangeEnd
		fileLen := uint64(len(data))
		if rangeEnd > fileLen {
			rangeEnd = fileLen
		}

		data = data[source.Http.RangeBegin:rangeEnd]
		c.l.WithFields(logrus.Fields{
			"len": len(data),
		}).Info("sliced to correct range")
	}

	return data, nil
}

func (c *Cache) updateLogicalCacheMapping(ctx context.Context, chunk *ccmsg.Chunk, txid common.EscrowID, chunkID common.ChunkID) error {
	// test if slot is getting re-assigned
	slot, err := models.LogicalCacheMappings(qm.Where("txid=? and slot_idx=?", txid, chunk.SlotIdx)).One(ctx, c.db)
	if err != nil {
		// missing row is fine, fall through in that case
		if err != sql.ErrNoRows {
			return errors.Wrap(err, "failed to select logical cache mapping from database")
		}
	} else {
		// slot is already in use, removing old data
		if err = c.Storage.DeleteData(slot.Txid, slot.ChunkID); err != nil {
			return errors.Wrap(err, "failed to remove old key from badger")
		}

		if _, err = slot.Delete(ctx, c.db); err != nil {
			return errors.Wrap(err, "failed to remove old logical cache mapping from database")
		}
	}

	// add the slot to the database
	lcm := &models.LogicalCacheMapping{
		Txid:          txid,
		SlotIdx:       chunk.SlotIdx,
		BlockEscrowID: "TODO",
		ChunkID:       chunkID,
	}
	err = lcm.Insert(ctx, c.db, boil.Infer())
	if err != nil {
		return errors.Wrap(err, "failed to add cache to database")
	}

	return nil
}

func (c *Cache) getEscrow(txid common.EscrowID) (*Escrow, error) {
	c.l.Debug("getting escrow reference ", txid)

	// try to pick an escrow from memory and fall back to the database
	e, ok := c.Escrows[txid]
	if ok {
		return e, nil
	}

	c.l.Debug("not found in memory, selecting from database")
	escrow, err := models.Escrows(qm.Where("txid=?", txid)).One(context.TODO(), c.db)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query Escrow")
	}

	return &Escrow{
		Inner: *escrow,
	}, nil
}

func (c *Cache) storeTicketL1(req *ccmsg.ClientCacheRequest) error {
	return nil
}

func (c *Cache) storeTicketL2(req *ccmsg.ClientCacheRequest) error {
	return nil
}

func (c *Cache) HandleRequest(ctx context.Context, req *ccmsg.ClientCacheRequest) (*ccmsg.ClientCacheResponse, error) {
	// Make sure that we're participating in this escrow.
	escrowID, err := common.BytesToEscrowID(req.BundleRemainder.EscrowId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to interpret escrow ID")
	}

	escrow, err := c.getEscrow(escrowID)
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

	var chunkID common.ChunkID
	if len(ticketRequest.ChunkId) != common.ChunkIDSize {
		return nil, errors.New("unexpected size for chunk ID")
	}
	copy(chunkID[:], ticketRequest.ChunkId)

	// If we don't have the chunk, ask the CP how to get it.
	escrowID, err := common.BytesToEscrowID(req.BundleRemainder.EscrowId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to interpret escrow ID")
	}
	objectID, err := common.BytesToObjectID(req.BundleRemainder.ObjectId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to interpret object ID")
	}
	chunk, err := c.getChunk(ctx, escrowID, objectID, ticketRequest.ChunkIdx, chunkID)
	if err != nil {
		c.l.WithError(err).Error("failed to get chunk") // XXX: Should just be doing this at the top level so that we see all errors.
		return nil, errors.Wrap(err, "failed to get chunk")
	}

	for _, masterKey := range [][]byte{escrow.InnerMasterKey(), escrow.OuterMasterKey()} {
		// XXX: Fix typing.
		seqNo := uint32(req.BundleRemainder.RequestSequenceNo)

		// Generate our session key frmo the master key.
		key, err := util.KeyedPRF(req.BundleRemainder.ClientPublicKey.PublicKey, seqNo, masterKey)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate key")
		}

		// Note that we use the key we've just generated (and not the master key) to generate the IV.
		iv, err := util.KeyedPRF(util.Uint64ToLE(ticketRequest.ChunkIdx), seqNo, key)
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
		ciphertext := make([]byte, len(chunk))
		stream.XORKeyStream(ciphertext, chunk)
		chunk = ciphertext
	}

	// Done!
	return &ccmsg.ClientCacheResponse{
		RequestSequenceNo: req.SequenceNo,
		Msg: &ccmsg.ClientCacheResponse_DataResponse{
			DataResponse: &ccmsg.ClientCacheResponseData{
				Data: chunk,
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

	outerSessionKey, err := util.KeyedPRF(req.BundleRemainder.ClientPublicKey.PublicKey, seqNo, escrow.OuterMasterKey())
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

func getPublisherAddr(ctx context.Context, publisherCacheAddr string) (string, error) {
	// XXX. if an ip/hostname is set, try to use that. This could be an address in a private ip range though
	if !strings.HasPrefix(publisherCacheAddr, ":") {
		return publisherCacheAddr, nil
	}

	peer, ok := peer.FromContext(ctx)
	if !ok {
		return "", errors.New("failed to get grpc peer from ctx")
	}

	var srcIP net.IP
	switch addr := peer.Addr.(type) {
	case *net.UDPAddr:
		srcIP = addr.IP
	case *net.TCPAddr:
		srcIP = addr.IP
	}

	publisherAddr := strings.Split(publisherCacheAddr, ":")
	portStr := publisherAddr[len(publisherAddr)-1]
	port, err := strconv.ParseUint(portStr, 10, 32)
	if err != nil {
		return "", errors.Wrap(err, "invalid port")
	}

	return fmt.Sprintf("%s:%d", srcIP, port), nil
}

func (c *Cache) OfferEscrow(ctx context.Context, req *ccmsg.EscrowOfferRequest) (*ccmsg.EscrowOfferResponse, error) {
	// TODO: ensure we have enough resources

	txid, err := common.BytesToEscrowID(req.EscrowId)
	if err != nil {
		return nil, errors.Wrap(err, "invalid escrow id")
	}

	l := c.l.WithFields(logrus.Fields{
		"txid": txid,
	})
	l.Info("starting to create escrow...")

	publisherAddr, err := getPublisherAddr(ctx, req.PublisherCacheAddr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get PublisherCacheAddr")
	}
	l.Info("found publisher addr: ", publisherAddr)

	escrow := &Escrow{
		Inner: models.Escrow{
			Txid:               txid,
			InnerMasterKey:     req.InnerMasterKey,
			OuterMasterKey:     req.OuterMasterKey,
			Slots:              req.Slots,
			PublisherCacheAddr: publisherAddr,
		},
	}
	c.Escrows[txid] = escrow

	l.Info("adding escrow to database")
	if err = c.AddEscrowToDatabase(ctx, escrow); err != nil {
		l.Error("failed to add escrow to database: ", err)
		return nil, errors.Wrap(err, "failed to add escrow to database")
	}
	l.Info("escrow successfully created")

	return &ccmsg.EscrowOfferResponse{}, nil
}
