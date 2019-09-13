package blockexplorer

import (
	"context"
	"net/http"

	"github.com/cachecashproject/go-cachecash/common"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Daemon defines the interface the CLI uses to start and stop the daemon.
type Daemon interface {
	common.StarterShutdowner
}

// daemon ties together the ledger client, the gin API server and the
// prometheus status server
type daemon struct {
	l            *logrus.Logger
	statusServer *statusServer
	ledgerClient *LedgerClient
	conf         *ConfigFile
	httpServer   *http.Server
}

var _ Daemon = (*daemon)(nil)

// NewDaemon creates a block explorer daemon
func NewDaemon(l *logrus.Logger, conf *ConfigFile) (Daemon, error) {
	client, err := NewLedgerClient(l, conf.LedgerAddr, conf.Insecure)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create ledger client")
	}

	httpServer, err := newBlockExplorerServer(l, client, conf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create explorer server")
	}

	statusServer := newStatusServer(l, conf)

	return &daemon{
		conf:         conf,
		l:            l,
		ledgerClient: client,
		statusServer: statusServer,
		httpServer:   httpServer,
	}, nil
}

func (a *daemon) Start() error {
	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil {
			a.l.WithError(err).Warn("httpServer.ListenAndServe() returned error")
		}
	}()
	if err := a.statusServer.Start(); err != nil {
		return errors.Wrap(err, "failed to start status server")
	}

	return nil
}

func (a *daemon) Shutdown(ctx context.Context) error {
	if err := a.httpServer.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to shut down httpd server")
	}
	if err := a.statusServer.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to shut down status server")
	}
	return nil
}
