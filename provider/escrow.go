package provider

import (
	"context"
	"crypto"
	"net"

	cachecash "github.com/kelleyk/go-cachecash"
	"github.com/kelleyk/go-cachecash/batchsignature"
	"github.com/kelleyk/go-cachecash/ccmsg"
	"github.com/kelleyk/go-cachecash/colocationpuzzle"
	"github.com/kelleyk/go-cachecash/common"
	"github.com/kelleyk/go-cachecash/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ed25519"
)

type ParticipatingCache struct {
	InnerMasterKey []byte // ...?

	PublicKey crypto.PublicKey

	Inetaddr net.IP
	Port     uint32
}

type Escrow struct {
	Provider *ContentProvider

	PublicKey  crypto.PublicKey
	PrivateKey crypto.PrivateKey

	Info *ccmsg.EscrowInfo

	Objects map[string]EscrowObjectInfo

	Caches []*ParticipatingCache
}

type EscrowObjectInfo struct {
	Object cachecash.ContentObject // XXX: This has to be removed in favor of the content catalog.
	ID     uint64
}

// The info object does not need to have its keys populated.
func (p *ContentProvider) NewEscrow(info *ccmsg.EscrowInfo) (*Escrow, error) {
	if info.DrawDelay == 0 {
		return nil, errors.New("draw delay must be at least one block")
	}
	if info.ExpirationDelay == 0 {
		return nil, errors.New("expiration delay must be at least one block")
	}
	if info.StartBlock == 0 {
		return nil, errors.New("start block number must be set")
	}
	// TODO: Perform additional validation on TicketsPerBlock.
	if len(info.TicketsPerBlock) == 0 {
		return nil, errors.New("tickets-per-block may not be empty")
	}
	// XXX: Validate info.EscrowID

	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate keypair")
	}

	// TODO: Should we set info.PublicKey and info.ProviderPublicKey?
	return &Escrow{
		Provider:   p,
		PublicKey:  pub,
		PrivateKey: priv,
		Info:       info,

		Objects: make(map[string]EscrowObjectInfo),
	}, nil
}

func (e *Escrow) reserveTicketNumbers(qty int) ([]uint64, error) {
	nos := make([]uint64, qty)
	for i := uint64(0); i < uint64(qty); i++ {
		nos = append(nos, i)
	}
	return nos, nil
}

// BundleParams is everything necessary to generate a complete TicketBundle message.
type BundleParams struct {
	Escrow            *Escrow // XXX: Do we need this?
	ObjectID          uint64  // This is a per-escrow value.
	Entries           []BundleEntryParams
	RequestSequenceNo uint64
	ClientPublicKey   crypto.PublicKey
	Object            cachecash.ContentObject
}
type BundleEntryParams struct {
	TicketNo uint64
	BlockIdx uint32
	Cache    *ParticipatingCache
}

func (e *Escrow) ID() common.EscrowID {
	// XXX: Temporary; replace me!
	var id common.EscrowID
	return id
}

func (e *Escrow) GetObjectByPath(ctx context.Context, path string) (cachecash.ContentObject, uint64, error) {
	info, ok := e.Objects[path]
	if !ok {
		return nil, 0, errors.New("no such object")
	}
	return info.Object, info.ID, nil
}

func NewBundleGenerator(l *logrus.Logger, signer batchsignature.BatchSigner) *BundleGenerator {
	return &BundleGenerator{
		l:      l,
		Signer: signer,
		PuzzleParams: &colocationpuzzle.Parameters{
			Rounds:      2,
			StartOffset: 0, // TODO: Not respected yet.
			StartRange:  0,
		},
	}
}

type BundleGenerator struct {
	l            *logrus.Logger
	PuzzleParams *colocationpuzzle.Parameters
	Signer       batchsignature.BatchSigner
}

// XXX: Attach this function to a struct containing configuration data (like e.g. puzzle parameters), or add those
// things as arguments.
func (gen *BundleGenerator) GenerateTicketBundle(bp *BundleParams) (*ccmsg.TicketBundle, error) {
	resp := &ccmsg.TicketBundle{
		// ProviderPublicKey: cachecash.PublicKeyMessage(e.Provider.PublicKey),
		// EscrowPublicKey:   cachecash.PublicKeyMessage(e.PublicKey),
		Remainder: &ccmsg.TicketBundleRemainder{
			RequestSequenceNo: bp.RequestSequenceNo,
			EscrowId:          nil, // XXX: Should be `bp.Escrow.ID()`
			ObjectId:          bp.ObjectID,
			// PuzzleInfo is filled in later
			ClientPublicKey: cachecash.PublicKeyMessage(bp.ClientPublicKey),
		},
	}

	if len(bp.Entries) == 0 {
		return nil, errors.New("must serve client at least one block")
	}

	// Generate inner keys (one per cache) using our keyed PRF.
	var innerKeys, innerIVs [][]byte
	for _, bep := range bp.Entries {
		prfInput := []byte(bp.ClientPublicKey.(ed25519.PublicKey)) // XXX:
		k, err := util.KeyedPRF(prfInput, uint32(bp.RequestSequenceNo), bep.Cache.InnerMasterKey)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate inner key")
		}
		innerKeys = append(innerKeys, k)

		iv, err := util.KeyedPRF(util.Uint64ToLE(uint64(bep.BlockIdx)), uint32(bp.RequestSequenceNo), k)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate inner IV")
		}
		innerIVs = append(innerIVs, iv)
	}

	blockIndices := make([]uint32, len(bp.Entries))
	for i, bep := range bp.Entries {
		blockIndices[i] = bep.BlockIdx

		// Generate a ticket-request for each cache.
		resp.TicketRequest = append(resp.TicketRequest, &ccmsg.TicketRequest{
			BlockIdx:       uint64(bep.BlockIdx),
			BlockId:        uint64(10000 + bep.BlockIdx), // XXX: This should be a hash, not just a reused block index.
			CachePublicKey: cachecash.PublicKeyMessage(bep.Cache.PublicKey),

			// XXX: Why is 'inner_key' in this message?  Regardless, we need the submessage not to be nil, or we'll get
			// a nil deref when computing the digest.
			InnerKey: &ccmsg.BlockKey{Key: nil},
		})

		resp.CacheInfo = append(resp.CacheInfo, &ccmsg.CacheInfo{
			Addr: &ccmsg.NetworkAddress{
				Inetaddr: bep.Cache.Inetaddr,
				Port:     bep.Cache.Port,
			},
		})

		// Generate a lottery ticket 1 for each cache.
		resp.TicketL1 = append(resp.TicketL1, &ccmsg.TicketL1{
			TicketNo:       bep.TicketNo,
			CachePublicKey: cachecash.PublicKeyMessage(bep.Cache.PublicKey), // XXX: Does this need to be repeated here?
		})
	}

	// Generate a colocation puzzle for the client to solve.
	gen.l.WithFields(logrus.Fields{
		"blockIdx": blockIndices,
	}).Info("generating puzzle")
	puzzle, err := colocationpuzzle.Generate(*gen.PuzzleParams, bp.Object, blockIndices, innerKeys, innerIVs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate colocation puzzle")
	}
	resp.Remainder.PuzzleInfo = &ccmsg.ColocationPuzzleInfo{
		Goal:        puzzle.Goal,
		Rounds:      gen.PuzzleParams.Rounds,
		StartOffset: uint64(gen.PuzzleParams.StartOffset), // XXX: Make typing consistent!
		StartRange:  uint64(gen.PuzzleParams.StartRange),
	}
	gen.l.WithFields(logrus.Fields{
		"initialOffset": puzzle.Offset,
		// "goal":          hex.EncodeToString(puzzle.Goal),
		// "secret":        hex.EncodeToString(puzzle.Secret),
	}).Info("generated colocation puzzle")

	// Generate a lottery ticket 2 and then marshal and encrypt it using a key and IV taken from the colocation puzzle's secret.
	ticketL2 := &ccmsg.TicketL2{}
	for _, k := range innerKeys {
		ticketL2.InnerSessionKey = append(ticketL2.InnerSessionKey, &ccmsg.BlockKey{Key: k})
	}
	resp.EncryptedTicketL2, err = common.EncryptTicketL2(puzzle, ticketL2)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal and encrypt ticket L2")
	}

	// Generate our batch signature (BHT).
	cd := resp.CanonicalDigest()
	sig, err := gen.Signer.BatchSign(cd)
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign ticket bundle")
	}
	resp.BatchSig = sig

	// Done!
	return resp, nil
}
