package metricsproxy

import (
	"context"
	"net"

	"github.com/cachecashproject/go-cachecash/metrics"

	"github.com/cachecashproject/go-cachecash/common"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// An Application is the top-level content publisher.  It takes a configuration struct.  Its children are the several
// protocol servers (that deal with clients, caches, and so forth).
type Application interface {
	common.StarterShutdowner
}

// ConfigFile defines the configuration available for the metrics proxy
type ConfigFile struct {
	MetricsGRPCAddr string
	StatusAddr      string
}

type application struct {
	l *logrus.Logger

	clientProtocolServer *clientProtocolServer
	statusServer         *statusServer
}

var _ Application = (*application)(nil)

// NewApplication makes a new Application
func NewApplication(l *logrus.Logger, conf *ConfigFile) (Application, error) {
	clientProtocolServer, metricsServer, err := newClientProtocolServer(l, conf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client protocol server")
	}

	statusServer, err := newStatusServer(l, conf, metricsServer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create status server")
	}

	return &application{
		l:                    l,
		clientProtocolServer: clientProtocolServer,
		statusServer:         statusServer,
	}, nil
}

func (a *application) Start() error {
	if err := a.clientProtocolServer.Start(); err != nil {
		return errors.Wrap(err, "failed to start client protocol server")
	}
	if err := a.statusServer.Start(); err != nil {
		return errors.Wrap(err, "failed to start status server")
	}
	return nil
}

func (a *application) Shutdown(ctx context.Context) error {
	if err := a.clientProtocolServer.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to shut down client protocol server")
	}
	if err := a.statusServer.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to shut down status server")
	}
	return nil
}

type clientProtocolServer struct {
	l          *logrus.Logger
	conf       *ConfigFile
	grpcServer *grpc.Server
	quitCh     chan bool
}

var _ common.StarterShutdowner = (*clientProtocolServer)(nil)

func newClientProtocolServer(l *logrus.Logger, conf *ConfigFile) (*clientProtocolServer, *grpcMetricsProxyServer, error) {
	grpcServer := common.NewGRPCServer()
	metricsServer := newGRPCMetricsProxyServer(l)
	metrics.RegisterMetricsServer(grpcServer, metricsServer)
	grpc_prometheus.Register(grpcServer)

	return &clientProtocolServer{
		l:          l,
		conf:       conf,
		grpcServer: grpcServer,
	}, metricsServer, nil
}

func (s *clientProtocolServer) Start() error {
	s.l.Info("clientProtocolServer - Start - enter")

	grpcLis, err := net.Listen("tcp", s.conf.MetricsGRPCAddr)
	if err != nil {
		return errors.Wrap(err, "failed to bind listener")
	}

	go func() {
		// This will block until we call `Stop`.
		if err := s.grpcServer.Serve(grpcLis); err != nil {
			s.l.WithError(err).Error("failed to serve clientProtocolServer(grpc)")
		}
	}()

	quit := make(chan bool, 1)
	s.quitCh = quit

	s.l.Info("clientProtocolServer - Start - exit")
	return nil
}

func (s *clientProtocolServer) Shutdown(ctx context.Context) error {
	// stop announcing our cache
	s.quitCh <- true

	// TODO: Should use `GracefulStop` until context expires, and then fall back on `Stop`.
	s.grpcServer.Stop()

	return nil
}
