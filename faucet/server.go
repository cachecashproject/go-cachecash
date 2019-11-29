package faucet

import (
	"context"
	"database/sql"
	"net"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Application interface {
	common.StarterShutdowner
}

type ConfigFile struct {
	FaucetAddr string
	LedgerAddr string

	Database string `json:"database"`
	Insecure bool   `json:"insecure"`
}

type application struct {
	l *logrus.Logger

	faucetServer *faucetServer
	// statusServer         *statusServer
	// TODO: ...
}

var _ Application = (*application)(nil)

// XXX: Should this take p as an argument, or be responsible for setting it up?
func NewApplication(l *logrus.Logger, p *Faucet, db *sql.DB, conf *ConfigFile) (Application, error) {
	faucetServer, err := newFaucetServer(l, p, db, conf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client protocol server")
	}

	return &application{
		l:            l,
		faucetServer: faucetServer,
	}, nil
}

func (a *application) Start() error {
	if err := a.faucetServer.Start(); err != nil {
		return errors.Wrap(err, "failed to start client protocol server")
	}
	return nil
}

func (a *application) Shutdown(ctx context.Context) error {
	if err := a.faucetServer.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to shut down client protocol server")
	}
	return nil
}

type faucetServer struct {
	l          *logrus.Logger
	conf       *ConfigFile
	faucet     *Faucet
	grpcServer *grpc.Server
}

var _ common.StarterShutdowner = (*faucetServer)(nil)

func newFaucetServer(l *logrus.Logger, f *Faucet, db *sql.DB, conf *ConfigFile) (*faucetServer, error) {
	grpcServer := common.NewDBGRPCServer(db)
	ccmsg.RegisterFaucetServer(grpcServer, &grpcFaucetServer{faucet: f})

	return &faucetServer{
		l:          l,
		conf:       conf,
		faucet:     f,
		grpcServer: grpcServer,
	}, nil
}

func (s *faucetServer) Start() error {
	s.l.Info("faucetServer - Start - enter")

	lis, err := net.Listen("tcp", s.conf.FaucetAddr)
	if err != nil {
		return errors.Wrap(err, "failed to bind listener")
	}

	go func() {
		// This will block until we call `Stop`.
		if err := s.grpcServer.Serve(lis); err != nil {
			s.l.WithError(err).Error("failed to serve faucetServer(grpc)")
		}
	}()

	s.l.Info("faucetServer - Start - exit")
	return nil
}

func (s *faucetServer) Shutdown(ctx context.Context) error {
	// TODO: Should use `GracefulStop` until context expires, and then fall back on `Stop`.
	s.grpcServer.Stop()

	return nil
}
