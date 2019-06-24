package bootstrap

import (
	"context"
	"net"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Application interface {
	common.StarterShutdowner
}

type ConfigFile struct {
	GrpcAddr   string `json:"grpc_addr"`
	Database   string `json:"database"`
	StatusAddr string
}

func (c *ConfigFile) FillDefaults() {
	if c.StatusAddr == "" {
		c.StatusAddr = ":8100"
	}
}

type application struct {
	l               *logrus.Logger
	bootstrapServer *bootstrapServer
	statusServer    *statusServer
}

var _ Application = (*application)(nil)

func NewApplication(l *logrus.Logger, b *Bootstrapd, conf *ConfigFile) (Application, error) {
	conf.FillDefaults()
	bootstrapServer, err := newBootstrapServer(l, b, conf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create bootstrap server")
	}

	statusServer, err := newStatusServer(l, conf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create status server")
	}

	return &application{
		l:               l,
		bootstrapServer: bootstrapServer,
		statusServer:    statusServer,
	}, nil
}

func (a *application) Start() error {
	if err := a.bootstrapServer.Start(); err != nil {
		return errors.Wrap(err, "failed to start bootstrap server")
	}
	if err := a.statusServer.Start(); err != nil {
		return errors.Wrap(err, "failed to start status server")
	}
	return nil
}

func (a *application) Shutdown(ctx context.Context) error {
	if err := a.bootstrapServer.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to shut down bootstrap server")
	}
	if err := a.statusServer.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to shut down status server")
	}
	return nil
}

type bootstrapServer struct {
	l          *logrus.Logger
	conf       *ConfigFile
	bootstrap  *Bootstrapd
	grpcServer *grpc.Server
}

var _ common.StarterShutdowner = (*bootstrapServer)(nil)

func newBootstrapServer(l *logrus.Logger, b *Bootstrapd, conf *ConfigFile) (*bootstrapServer, error) {
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)
	ccmsg.RegisterNodeBootstrapdServer(grpcServer, &grpcBootstrapServer{bootstrap: b})
	grpc_prometheus.Register(grpcServer)

	return &bootstrapServer{
		l:          l,
		conf:       conf,
		bootstrap:  b,
		grpcServer: grpcServer,
	}, nil
}

func (s *bootstrapServer) Start() error {
	s.l.Info("bootstrapServer - Start - enter")

	grpcLis, err := net.Listen("tcp", s.conf.GrpcAddr)
	if err != nil {
		return errors.Wrap(err, "failed to bind listener")
	}

	go func() {
		// This will block until we call `Stop`.
		if err := s.grpcServer.Serve(grpcLis); err != nil {
			s.l.WithError(err).Error("failed to serve bootstrapServer(grpc)")
		}
	}()

	s.l.Info("bootstrapServer - Start - exit")
	return nil
}

func (s *bootstrapServer) Shutdown(ctx context.Context) error {
	// TODO: Should use `GracefulStop` until context expires, and then fall back on `Stop`.
	s.grpcServer.Stop()
	return nil
}
