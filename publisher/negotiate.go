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
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"golang.org/x/crypto/ed25519"
)

type agreement struct {
	cache *ParticipatingCache
	err   error
}

func addrFromCacheDescription(cache *ccmsg.CacheDescription) string {
	return fmt.Sprintf("%s:%d", cache.ExternalIp, cache.Port)
}

func InitEscrows(ctx context.Context, s *publisherServer, caches []*ccmsg.CacheDescription) error {
	if len(s.publisher.escrows) > 0 {
		// escrow already exists
		return nil
	}

	if len(caches) < 4 {
		return errors.New("not enough caches available")
	}

	s.l.Info("no existing escrow, creating one")
	escrow, err := CreateEscrow(ctx, s.publisher, caches)
	if err != nil {
		return errors.Wrap(err, "failed to create escrow")
	}

	s.l.Info("successfully created an escrow")
	err = s.publisher.AddEscrowToDatabase(ctx, escrow)
	if err != nil {
		return errors.Wrap(err, "failed to add escrow to database")
	}
	s.l.Info("escrow fully setup")

	return nil
}

func CreateEscrow(ctx context.Context, publisher *ContentPublisher, cacheDescriptions []*ccmsg.CacheDescription) (*Escrow, error) {
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
		EscrowId:       escrowID[:],
		InnerMasterKey: innerMasterKey,
		OuterMasterKey: outerMasterKey,
		Slots:          2500,
		PublisherAddr:  publisher.PublisherAddr,
	}

	num := len(cacheDescriptions)
	ch := make(chan agreement, num)

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
		Publisher: publisher,
		Inner: models.Escrow{
			Txid:       escrowID,
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
	conn, err := common.GRPCDial(addr)
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

func UpdateKnownCaches(ctx context.Context, s *publisherServer, caches []*ccmsg.CacheDescription) error {
	for _, cache := range caches {
		model, err := models.Caches(qm.Where("public_key = ?", cache.PublicKey)).One(ctx, s.publisher.db)
		if err != nil {
			continue
		}

		inetAddr := net.ParseIP(cache.ExternalIp)
		port := cache.Port

		if !model.Inetaddr.Equal(inetAddr) || model.Port != port {
			s.l.Infof("Updating address of cache that is part of an escrow from %s to %s", model.Inetaddr, inetAddr)

			model.Inetaddr = inetAddr
			model.Port = port

			// update escrows in memory too
			c, ok := s.publisher.caches[string(cache.PublicKey)]
			if ok {
				c.participation.Cache.Inetaddr = inetAddr
				c.participation.Cache.Port = port
			}

			_, err = model.Update(ctx, s.publisher.db, boil.Infer())
			if err != nil {
				return errors.Wrap(err, "failed to update known cache")
			}
		}
	}

	return nil
}
