package bootstrap

import (
	"context"
	"time"

	cachecash "github.com/cachecashproject/go-cachecash"
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ed25519"
)

type Client struct {
	l          *logrus.Logger
	grpcClient ccmsg.NodeBootstrapdClient
}

func NewClient(l *logrus.Logger, addr string, insecure bool) (*Client, error) {
	// XXX: Should not create a new connection for each attempt.
	l.Info("dialing bootstrap service: ", addr)
	conn, err := common.GRPCDial(addr, insecure)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial bootstrap service")
	}

	grpcClient := ccmsg.NewNodeBootstrapdClient(conn)

	return &Client{
		l:          l,
		grpcClient: grpcClient,
	}, nil
}

type BootstrapInfo struct {
	PublicKey   ed25519.PublicKey
	Stats       *CacheStats
	StartupTime time.Time
	Port        uint32
	ContactUrl  string
}

func (c *Client) AnnounceCache(ctx context.Context, info BootstrapInfo) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	c.l.Info("announcing our cache to the bootstrap service")
	_, err := c.grpcClient.AnnounceCache(ctx, &ccmsg.CacheAnnounceRequest{
		PublicKey:   info.PublicKey,
		Version:     cachecash.CurrentVersion,
		FreeMemory:  info.Stats.FreeMemory,
		TotalMemory: info.Stats.TotalMemory,
		FreeDisk:    info.Stats.FreeDisk,
		TotalDisk:   info.Stats.TotalDisk,
		StartupTime: info.StartupTime.Unix(),
		ContactUrl:  info.ContactUrl,
		Port:        info.Port,
	})
	if err != nil {
		return errors.Wrap(err, "failed to announce our cache to the bootstrap service")
	}
	return nil
}

func (c *Client) FetchCaches(ctx context.Context) ([]*ccmsg.CacheDescription, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	c.l.Info("fetching caches from bootstrap service")
	resp, err := c.grpcClient.FetchCaches(ctx, &ccmsg.CacheFetchRequest{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch caches from the bootstrap service")
	}
	return resp.Caches, nil
}
