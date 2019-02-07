package publisher

import (
	"context"
	"net"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ed25519"
	"google.golang.org/grpc"
)

// An Application is the top-level content publisher.  It takes a configuration struct.  Its children are the several
// protocol servers (that deal with clients, caches, and so forth).
type Application interface {
	common.StarterShutdowner
}

type ConfigFile struct {
	Config      *Config            `json:"config"`
	Escrows     []*Escrow          `json:"escrows"`
	UpstreamURL string             `json:"upstreamURL"`
	PrivateKey  ed25519.PrivateKey `json:"privateKey"`
}

// XXX: Right now, this is shared between the client- and cache-facing servers.
type Config struct {
	ClientProtocolAddr string
	CacheProtocolAddr  string
}

func (c *Config) FillDefaults() {
	if c.ClientProtocolAddr == "" {
		c.ClientProtocolAddr = ":8080"
	}
	if c.CacheProtocolAddr == "" {
		c.CacheProtocolAddr = ":8082"
	}
}

type application struct {
	l *logrus.Logger

	clientProtocolServer *clientProtocolServer
	cacheProtocolServer  *cacheProtocolServer
	// TODO: ...
}

var _ Application = (*application)(nil)

// XXX: Should this take p as an argument, or be responsible for setting it up?
func NewApplication(l *logrus.Logger, p *ContentPublisher, conf *Config) (Application, error) {
	conf.FillDefaults()

	clientProtocolServer, err := newClientProtocolServer(l, p, conf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client protocol server")
	}

	cacheProtocolServer, err := newCacheProtocolServer(l, p, conf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cache protocol server")
	}

	return &application{
		l:                    l,
		clientProtocolServer: clientProtocolServer,
		cacheProtocolServer:  cacheProtocolServer,
	}, nil
}

func (a *application) Start() error {
	if err := a.clientProtocolServer.Start(); err != nil {
		return errors.Wrap(err, "failed to start client protocol server")
	}
	if err := a.cacheProtocolServer.Start(); err != nil {
		return errors.Wrap(err, "failed to start cache protocol server")
	}
	return nil
}

func (a *application) Shutdown(ctx context.Context) error {
	return a.clientProtocolServer.Shutdown(ctx)
}

type clientProtocolServer struct {
	l          *logrus.Logger
	conf       *Config
	publisher  *ContentPublisher
	grpcServer *grpc.Server
}

var _ common.StarterShutdowner = (*clientProtocolServer)(nil)

func newClientProtocolServer(l *logrus.Logger, p *ContentPublisher, conf *Config) (*clientProtocolServer, error) {
	grpcServer := grpc.NewServer()
	ccmsg.RegisterClientPublisherServer(grpcServer, &grpcClientPublisherServer{publisher: p})

	return &clientProtocolServer{
		l:          l,
		conf:       conf,
		publisher:  p,
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

type cacheProtocolServer struct {
	l          *logrus.Logger
	conf       *Config
	publisher  *ContentPublisher
	grpcServer *grpc.Server
}

var _ common.StarterShutdowner = (*cacheProtocolServer)(nil)

func newCacheProtocolServer(l *logrus.Logger, p *ContentPublisher, conf *Config) (*cacheProtocolServer, error) {
	grpcServer := grpc.NewServer()
	ccmsg.RegisterCachePublisherServer(grpcServer, &grpcCachePublisherServer{publisher: p})

	return &cacheProtocolServer{
		l:          l,
		conf:       conf,
		publisher:  p,
		grpcServer: grpcServer,
	}, nil
}

func (s *cacheProtocolServer) Start() error {
	s.l.Info("cacheProtocolServer - Start - enter")

	lis, err := net.Listen("tcp", s.conf.CacheProtocolAddr)
	if err != nil {
		return errors.Wrap(err, "failed to bind listener")
	}

	go func() {
		// This will block until we call `Stop`.
		if err := s.grpcServer.Serve(lis); err != nil {
			s.l.WithError(err).Error("failed to serve cacheProtocolServer")
		}
	}()

	s.l.Info("cacheProtocolServer - Start - exit")
	return nil
}

func (s *cacheProtocolServer) Shutdown(ctx context.Context) error {
	// TODO: Should use `GracefulStop` until context expires, and then fall back on `Stop`.
	s.grpcServer.Stop()

	return nil
}
