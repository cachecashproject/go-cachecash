package bootstrap

import (
	"context"
	"time"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ed25519"
	"google.golang.org/grpc"
)

type Client struct {
	l          *logrus.Logger
	grpcClient ccmsg.NodeBootstrapdClient
}

func NewClient(l *logrus.Logger, addr string) (*Client, error) {
	// XXX: No transport security!
	// XXX: Should not create a new connection for each attempt.
	l.Info("dialing bootstrap service: ", addr)
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial bootstrap service")
	}

	grpcClient := ccmsg.NewNodeBootstrapdClient(conn)

	return &Client{
		l:          l,
		grpcClient: grpcClient,
	}, nil
}

func (c *Client) AnnounceCache(ctx context.Context, publicKey ed25519.PublicKey, port uint32, startupTime time.Time, info *CacheInfo) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	c.l.Info("announcing our cache to the bootstrap service")
	_, err := c.grpcClient.AnnounceCache(ctx, &ccmsg.CacheAnnounceRequest{
		PublicKey:   publicKey,
		Version:     "todo", // TODO
		FreeMemory:  info.FreeMemory,
		TotalMemory: info.TotalMemory,
		FreeDisk:    info.FreeDisk,
		TotalDisk:   info.TotalDisk,
		StartupTime: startupTime.Unix(),
		ContactUrl:  "", // TODO
		Port:        port,
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
