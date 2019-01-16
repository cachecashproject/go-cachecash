package cache

import (
	"context"
	"net"

	"github.com/kelleyk/go-cachecash/ccmsg"
	"github.com/kelleyk/go-cachecash/common"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// An Application is the top-level content provider.  It takes a configuration struct.  Its children are the several
// protocol servers (that deal with clients, caches, and so forth).
type Application interface {
	common.StarterShutdowner
}

type Config struct {
	ClientProtocolAddr string
}

func (c *Config) FillDefaults() {
	if c.ClientProtocolAddr == "" {
		c.ClientProtocolAddr = ":9000"
	}
}

type application struct {
	l *logrus.Logger

	clientProtocolServer *clientProtocolServer
	// TODO: ...
}

var _ Application = (*application)(nil)

// XXX: Should this take p as an argument, or be responsible for setting it up?
func NewApplication(l *logrus.Logger, c *Cache, conf *Config) (Application, error) {
	conf.FillDefaults()

	clientProtocolServer, err := newClientProtocolServer(l, c, conf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client protocol server")
	}

	return &application{
		l:                    l,
		clientProtocolServer: clientProtocolServer,
	}, nil
}

func (a *application) Start() error {
	if err := a.clientProtocolServer.Start(); err != nil {
		return errors.Wrap(err, "failed to start client protocol server")
	}
	return nil
}

func (a *application) Shutdown(ctx context.Context) error {
	return a.clientProtocolServer.Shutdown(ctx)
}

type clientProtocolServer struct {
	l          *logrus.Logger
	conf       *Config
	cache      *Cache
	grpcServer *grpc.Server
}

var _ common.StarterShutdowner = (*clientProtocolServer)(nil)

func newClientProtocolServer(l *logrus.Logger, c *Cache, conf *Config) (*clientProtocolServer, error) {
	grpcServer := grpc.NewServer()
	ccmsg.RegisterClientCacheServer(grpcServer, &grpcClientCacheServer{cache: c})

	return &clientProtocolServer{
		l:          l,
		conf:       conf,
		cache:      c,
		grpcServer: grpcServer,
	}, nil
}

func (s *clientProtocolServer) Start() error {
	s.l.Info("clientProtocolServer - Start - enter")

	lis, err := net.Listen("tcp", s.conf.ClientProtocolAddr)
	if err != nil {
		return errors.Wrap(err, "failed to bind listener")
	}

	go func() {
		// This will block until we call `Stop`.
		if err := s.grpcServer.Serve(lis); err != nil {
			s.l.WithError(err).Error("failed to serve clientProtocolServer")
		}
	}()

	s.l.Info("clientProtocolServer - Start - exit")
	return nil
}

func (s *clientProtocolServer) Shutdown(ctx context.Context) error {
	// TODO: Should use `GracefulStop` until context expires, and then fall back on `Stop`.
	s.grpcServer.Stop()

	return nil
}
