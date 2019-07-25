package ledgerservice

import (
	"context"
	"net"
	"net/http"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ed25519"
	"google.golang.org/grpc"
)

type Application interface {
	common.StarterShutdowner
}

type ConfigFile struct {
	LedgerProtocolAddr string
	StatusAddr         string

	PrivateKey ed25519.PrivateKey `json:"privateKey"`
	Database   string             `json:"database"`
}

func (c *ConfigFile) FillDefaults() {
	if c.LedgerProtocolAddr == "" {
		c.LedgerProtocolAddr = ":9090"
	}
	if c.StatusAddr == "" {
		c.StatusAddr = ":8100"
	}
}

type application struct {
	l *logrus.Logger

	ledgerProtocolServer *ledgerProtocolServer
	statusServer         *statusServer
	// TODO: ...
}

var _ Application = (*application)(nil)

// XXX: Should this take p as an argument, or be responsible for setting it up?
func NewApplication(l *logrus.Logger, p *LedgerService, conf *ConfigFile) (Application, error) {
	conf.FillDefaults()

	ledgerProtocolServer, err := newLedgerProtocolServer(l, p, conf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client protocol server")
	}

	statusServer, err := newStatusServer(l, p, conf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create status server")
	}

	return &application{
		l:                    l,
		ledgerProtocolServer: ledgerProtocolServer,
		statusServer:         statusServer,
	}, nil
}

func (a *application) Start() error {
	if err := a.ledgerProtocolServer.Start(); err != nil {
		return errors.Wrap(err, "failed to start client protocol server")
	}
	if err := a.statusServer.Start(); err != nil {
		return errors.Wrap(err, "failed to start status server")
	}
	return nil
}

func (a *application) Shutdown(ctx context.Context) error {
	if err := a.ledgerProtocolServer.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to shut down client protocol server")
	}
	if err := a.statusServer.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to shut down status server")
	}
	return nil
}

type ledgerProtocolServer struct {
	l             *logrus.Logger
	conf          *ConfigFile
	ledgerService *LedgerService
	grpcServer    *grpc.Server
	httpServer    *http.Server
}

var _ common.StarterShutdowner = (*ledgerProtocolServer)(nil)

func newLedgerProtocolServer(l *logrus.Logger, s *LedgerService, conf *ConfigFile) (*ledgerProtocolServer, error) {
	grpcServer := common.NewGRPCServer()
	ccmsg.RegisterLedgerServer(grpcServer, &grpcLedgerServer{ledgerService: s})

	httpServer := wrapGrpc(grpcServer)

	return &ledgerProtocolServer{
		l:             l,
		conf:          conf,
		ledgerService: s,
		grpcServer:    grpcServer,
		httpServer:    httpServer,
	}, nil
}

func wrapGrpc(grpcServer *grpc.Server) *http.Server {
	wrappedServer := grpcweb.WrapServer(grpcServer)

	handler := func(resp http.ResponseWriter, req *http.Request) {
		wrappedServer.ServeHTTP(resp, req)
	}

	return &http.Server{
		// Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(handler),
	}
}

func (s *ledgerProtocolServer) Start() error {
	s.l.Info("ledgerProtocolServer - Start - enter")

	lis, err := net.Listen("tcp", s.conf.LedgerProtocolAddr)
	if err != nil {
		return errors.Wrap(err, "failed to bind listener")
	}

	// httpLis, err := net.Listen("tcp", s.conf.LedgerProtocolHttpAddr)
	httpLis, err := net.Listen("tcp", ":8043")
	if err != nil {
		return errors.Wrap(err, "failed to bind listener")
	}

	go func() {
		// This will block until we call `Stop`.
		if err := s.grpcServer.Serve(lis); err != nil {
			s.l.WithError(err).Error("failed to serve ledgerProtocolServer(grpc)")
		}
	}()

	go func() {
		// This will block until we call `Stop`.
		if err := s.httpServer.Serve(httpLis); err != nil {
			s.l.WithError(err).Error("failed to serve ledgerProtocolServer(http)")
		}
	}()

	s.l.Info("ledgerProtocolServer - Start - exit")
	return nil
}

func (s *ledgerProtocolServer) Shutdown(ctx context.Context) error {
	// TODO: Should use `GracefulStop` until context expires, and then fall back on `Stop`.
	s.grpcServer.Stop()

	return nil
}
