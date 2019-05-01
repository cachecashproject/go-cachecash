package publisher

import (
	"context"
	"crypto/rand"
	"fmt"
	"net"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/publisher/models"
	"github.com/cachecashproject/go-cachecash/testutil"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ed25519"
	"google.golang.org/grpc"
)

type agreement struct {
	cache *ParticipatingCache
	err   error
}

func addrFromCacheDescription(cache *ccmsg.CacheDescription) string {
	return fmt.Sprintf("%s:%d", cache.ExternalIp, cache.Port)
}

func InitEscrows(s *cacheProtocolServer, caches []*ccmsg.CacheDescription) error {
	if len(s.publisher.escrows) > 0 {
		// escrow already exists
		return nil
	}

	if len(caches) < 4 {
		return errors.New("not enough caches available")
	}

	s.l.Info("no existing escrow, creating one")
	escrow, err := CreateEscrow(s.publisher, caches)
	if err != nil {
		return errors.Wrap(err, "failed to create escrow")
	}

	s.l.Info("successfully created an escrow")
	err = s.publisher.AddEscrowToDatabase(context.Background(), escrow)
	if err != nil {
		return errors.Wrap(err, "failed to add escrow to database")
	}
	s.l.Info("escrow fully setup")

	return nil
}

func CreateEscrow(publisher *ContentPublisher, cacheDescriptions []*ccmsg.CacheDescription) (*Escrow, error) {
	// TODO: remove testutil
	escrowID, err := common.BytesToEscrowID(testutil.RandBytes(common.EscrowIDSize))
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate escrow id")
	}

	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate publisher keypair")
	}

	innerMasterKey := make([]byte, 16)
	if _, err := rand.Read(innerMasterKey); err != nil {
		return nil, errors.Wrap(err, "failed to generate inner master key")
	}

	outerMasterKey := make([]byte, 16)
	if _, err := rand.Read(outerMasterKey); err != nil {
		return nil, errors.Wrap(err, "failed to generate inner master key")
	}

	offerRequest := &ccmsg.EscrowOfferRequest{
		EscrowId:                  escrowID[:],
		InnerMasterKey:            innerMasterKey,
		OuterMasterKey:            outerMasterKey,
		Slots:                     2500,
		PublisherCacheServiceAddr: publisher.PublisherCacheServiceAddr,
	}

	num := len(cacheDescriptions)
	ch := make(chan agreement, num)

	ctx := context.Background()

	for i, descr := range cacheDescriptions {
		go func(i int, descr *ccmsg.CacheDescription) {
			cache, err := OfferEscrow(ctx, publisher.l, offerRequest, descr)
			ch <- agreement{
				cache: cache,
				err:   err,
			}
		}(i, descr)
	}

	// collect results from goroutines
	caches := []*ParticipatingCache{}
	for i := 0; i < num; i++ {
		x := <-ch
		if x.err != nil {
			// TODO: don't give up on the whole escrow if one cache fails
			return nil, errors.Wrap(x.err, "failed to negotiate escrow")
		}
		caches = append(caches, x.cache)
	}
	publisher.l.Debug("Participating caches: ", caches)

	return &Escrow{
		ID:        escrowID,
		Publisher: publisher,
		Inner: models.Escrow{
			StartBlock: 0,
			EndBlock:   0,
			State:      "ok",
			PublicKey:  publicKey,
			PrivateKey: privateKey,
			Raw:        []byte{},
		},
		Caches: caches,
	}, nil
}

func OfferEscrow(ctx context.Context, l *logrus.Logger, offerRequest *ccmsg.EscrowOfferRequest, descr *ccmsg.CacheDescription) (*ParticipatingCache, error) {
	addr := addrFromCacheDescription(descr)
	l.Info("Offering escrow to ", addr)
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial bootstrap service")
	}

	client := ccmsg.NewPublisherCacheClient(conn)
	_, err = client.OfferEscrow(ctx, offerRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to offer escrow")
	}

	pubkey := ed25519.PublicKey(descr.PublicKey)
	inetaddr := net.ParseIP(descr.ExternalIp)
	if inetaddr == nil {
		return nil, errors.New("failed to parse ip")
	}
	port := descr.Port

	return &ParticipatingCache{
		Cache: models.Cache{
			PublicKey: pubkey,
			Inetaddr:  inetaddr,
			Port:      port,
		},
		InnerMasterKey: offerRequest.InnerMasterKey,
	}, nil
}
