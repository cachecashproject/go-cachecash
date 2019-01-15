package common

import (
	"context"
	"encoding/hex"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
)

const (
	BlockIDSize = 16
)

type BlockID [BlockIDSize]byte

func (id BlockID) String() string {
	return hex.EncodeToString(id[:])
}

type StarterShutdowner interface {
	// Start kicks off any internal goroutines that are necessary and returns synchronously.  It may temporarily block
	// until whatever needs to be started has been started.
	Start() error
	// Shutdown blocks until a graceful shutdown takes place.  If the provided context expires, the shutdown is forced.
	Shutdown(context.Context) error
}

func RunStarterShutdowner(o StarterShutdowner) error {
	// When we receive SIGINT or SIGTERM, exit.
	sigCh := make(chan os.Signal)
	stopCtx, stop := context.WithCancel(context.Background())
	go func() {
		for range sigCh {
			stop()
		}
	}()
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	log.Printf("starting...\n")
	if err := o.Start(); err != nil {
		return errors.Wrap(err, "failed to start")
	}

	// Block until a signal causes us to cancel stopCtx.
	log.Printf("ready...\n")
	<-stopCtx.Done()

	// Wait until outstanding requests finish or our timeout expires.
	log.Printf("graceful shutdown...\n")
	gracefulCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := o.Shutdown(gracefulCtx); err != nil {
		return errors.Wrap(err, "shutdown error")
	}

	log.Printf("shutdown complete")
	return nil
}
