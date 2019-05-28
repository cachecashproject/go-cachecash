package common

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type StarterShutdowner interface {
	// Start kicks off any internal goroutines that are necessary and returns synchronously.  It may temporarily block
	// until whatever needs to be started has been started.
	Start() error
	// Shutdown blocks until a graceful shutdown takes place.  If the provided context expires, the shutdown is forced.
	Shutdown(context.Context) error
}

func RunStarterShutdowner(l *logrus.Logger, o StarterShutdowner) error {
	// When we receive SIGINT or SIGTERM, exit.
	sigCh := make(chan os.Signal)
	stopCtx, stop := context.WithCancel(context.Background())
	go func() {
		for range sigCh {
			stop()
		}
	}()
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	l.Info("starting...")
	if err := o.Start(); err != nil {
		return errors.Wrap(err, "failed to start")
	}

	// Block until a signal causes us to cancel stopCtx.
	l.Info("ready...")
	<-stopCtx.Done()

	// Wait until outstanding requests finish or our timeout expires.
	l.Info("graceful shutdown...")
	gracefulCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := o.Shutdown(gracefulCtx); err != nil {
		return errors.Wrap(err, "shutdown error")
	}

	l.Info("shutdown complete")
	return nil
}
