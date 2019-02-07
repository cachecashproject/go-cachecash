package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/cachecashproject/go-cachecash/client"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// `cachecash-curl` is a simple command-line utility that retrieves a file being served via CacheCash.

var (
	outputPath  = flag.String("o", "", "Path where retrieved file will be written")
	logLevelStr = flag.String("logLevel", "info", "Verbosity of log output")
)

func main() {
	if err := mainC(); err != nil {
		if _, err := os.Stderr.WriteString(err.Error() + "\n"); err != nil {
			panic(err)
		}
		os.Exit(1)
	}
}

func mainC() error {
	flag.Parse()
	log.SetFlags(0)

	l := logrus.New()
	logLevel, err := logrus.ParseLevel(*logLevelStr)
	if err != nil {
		return errors.Wrap(err, "failed to parse log level")
	}
	l.SetLevel(logLevel)

	// e.g. "cachecash://localhost:8080/foo/bar"
	rawURI := flag.Arg(0)

	// TODO: This URI parsing should probably be moved into a library function somewhere.  The reason that it's being
	// done here is that `client.Client` only supports a single publisher right now, so the `GetObject` function can't
	// take whole URIs directly.
	// XXX: We silently ignore other parts of the URI (user, password, query string).  That's probably not a good idea;
	// we should either support them or return an error if they are present.
	u, err := url.Parse(rawURI)
	if err != nil {
		return errors.Wrap(err, "failed to parse URI")
	}
	if u.Scheme != "cachecash" {
		return errors.New("unexpected scheme in URI")
	}
	publisherAddr := u.Hostname()
	if u.Port() != "" {
		publisherAddr += ":" + u.Port()
	}
	objPath := u.Path

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interruptCh := make(chan os.Signal, 1)
	signal.Notify(interruptCh, os.Interrupt)
	go func() {
		select {
		case <-ctx.Done():
			return
		case <-interruptCh:
			cancel()
		}
	}()

	cl, err := client.New(l, publisherAddr) // e.g. "localhost:8080"
	if err != nil {
		return errors.Wrap(err, "failed to create client")
	}
	log.Printf("created client\n")

	o, err := cl.GetObject(ctx, objPath) // e.g. "/foo/bar"
	if err != nil {
		return errors.Wrap(err, "failed to fetch object")
	}

	log.Printf("fetch complete; shutting down client\n")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 999*time.Second)
	defer shutdownCancel()
	if err := cl.Close(shutdownCtx); err != nil {
		return errors.Wrap(err, "failed to shut down client")
	}

	if *outputPath != "" {
		log.Printf("writing data to file")
		if err := ioutil.WriteFile(*outputPath, o.Data(), 0644); err != nil {
			return errors.Wrap(err, "failed to write data to file")
		}
	}

	log.Printf("completed without error\n")
	return nil
}
