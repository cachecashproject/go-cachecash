package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"time"

	"github.com/cachecashproject/go-cachecash/bootstrap/models"
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"golang.org/x/crypto/ed25519"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type Bootstrapd struct {
	l  *logrus.Logger
	db *sql.DB
}

func NewBootstrapd(l *logrus.Logger, db *sql.DB) (*Bootstrapd, error) {
	return &Bootstrapd{
		l:  l,
		db: db,
	}, nil
}

func (b *Bootstrapd) verifyCacheIsReachable(ctx context.Context, srcIP net.IP, port uint32) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	addr := fmt.Sprintf("%s:%d", srcIP.String(), port)
	l := b.l.WithFields(logrus.Fields{
		"addr": addr,
	})
	l.Info("dialing cache back")
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		l.Error("failed to dial cache address")
		return errors.Wrap(err, "failed to dial cache address")
	}
	grpcClient := ccmsg.NewPublisherCacheClient(conn)
	_, err = grpcClient.PingCache(ctx, &ccmsg.PingCacheRequest{})
	if err != nil {
		l.Error("ping failed, cache seems defunct: ", err)
		return errors.Wrap(err, "ping failed")
	}
	l.Info("cache dailed successfully")
	return nil
}

func (b *Bootstrapd) HandleCacheAnnounceRequest(ctx context.Context, req *ccmsg.CacheAnnounceRequest) (*ccmsg.CacheAnnounceResponse, error) {
	startupTime := time.Unix(req.StartupTime, 0)

	peer, ok := peer.FromContext(ctx)
	if !ok {
		return nil, errors.New("failed to get grpc peer from ctx")
	}

	var srcIP net.IP
	switch addr := peer.Addr.(type) {
	case *net.UDPAddr:
		srcIP = addr.IP
	case *net.TCPAddr:
		srcIP = addr.IP
	}

	err := b.verifyCacheIsReachable(ctx, srcIP, req.Port)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect back to cache")
	}

	cache := models.Cache{
		PublicKey:   ed25519.PublicKey(req.PublicKey),
		Version:     req.Version,
		FreeMemory:  req.FreeMemory,
		TotalMemory: req.TotalMemory,
		FreeDisk:    req.FreeDisk,
		TotalDisk:   req.TotalDisk,
		StartupTime: startupTime,
		ExternalIP:  srcIP,
		Port:        req.Port,
		ContactURL:  req.ContactUrl,
		LastPing:    time.Now(),
	}

	// TODO: figure out how to do proper upserts
	/*
		err := cache.Upsert(ctx, b.db, true, []string{"public_key"}, boil.Infer(), boil.Infer())
		if err != nil {
			return nil, errors.Wrap(err, "failed to add cache to database")
		}
	*/

	// XXX: ignore duplicate key errors
	_ = cache.Insert(ctx, b.db, boil.Infer())

	// force an update in case the insert failed due to a conflict
	_, err = cache.Update(ctx, b.db, boil.Infer())
	if err != nil {
		return nil, err
	}

	err = b.reapStaleAnnouncements(ctx)
	if err != nil {
		return nil, err
	}

	return &ccmsg.CacheAnnounceResponse{}, nil
}

func (b *Bootstrapd) reapStaleAnnouncements(ctx context.Context) error {
	deadline := time.Now().Add(-5 * time.Minute)
	rows, err := models.Caches(qm.Where("last_ping<?", deadline)).DeleteAll(ctx, b.db)
	if err != nil {
		return err
	}
	b.l.Debugf("Removed %d stale caches from database", rows)
	return nil
}

func (b *Bootstrapd) HandleCacheFetchRequest(ctx context.Context, req *ccmsg.CacheFetchRequest) (*ccmsg.CacheFetchResponse, error) {
	err := b.reapStaleAnnouncements(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to reap stale announcement")
	}

	caches, err := models.Caches().All(ctx, b.db)
	if err != nil {
		return nil, err
	}

	resp := &ccmsg.CacheFetchResponse{}
	for _, c := range caches {
		resp.Caches = append(resp.Caches, &ccmsg.CacheDescription{
			PublicKey:   c.PublicKey,
			Version:     c.Version,
			FreeMemory:  c.FreeMemory,
			TotalMemory: c.TotalMemory,
			FreeDisk:    c.FreeDisk,
			TotalDisk:   c.TotalDisk,
			StartupTime: c.StartupTime.Unix(),
			ContactUrl:  c.ContactURL,
			ExternalIp:  c.ExternalIP.String(),
			Port:        c.Port,
		})
	}

	return resp, nil
}
