package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/cachecashproject/go-cachecash/client"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/log"
	"github.com/pkg/errors"

	"go.opencensus.io/trace"
)

// `cachecash-curl` is a simple command-line utility that retrieves a file being served via CacheCash.

var (
	outputPath = flag.String("o", "", "Path where retrieved file will be written")
	traceAPI   = flag.String("trace", "", "Jaeger API for tracing")
)

func main() {
	common.Main(mainC)
}

func mainC() error {
	l := log.NewCLILogger("cachecash-curl", log.CLIOpt{})
	flag.Parse()

	startTime := time.Now()

	p, err := common.NewConfigParser(&l.Logger, "cachecash-curl")
	if err != nil {
		return err
	}
	insecure := p.GetInsecure()

	if err := l.ConfigureLogger(); err != nil {
		return errors.Wrap(err, "failed to configure logger")
	}

	defer common.SetupTracing(*traceAPI, "cachecash-curl", &l.Logger).Flush()
	// As a rarely used CLI tool, trace always.
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	// e.g. "cachecash://localhost:7070/foo/bar"
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

	interruptCh := make(chan os.Signal, 1) //TODO: OUTPUT TO FILE
	signal.Notify(interruptCh, os.Interrupt)
	go func() {
		select {
		case <-ctx.Done():
			return
		case <-interruptCh:
			cancel()
		}
	}()

	cl, err := client.New(&l.Logger, publisherAddr, insecure) // e.g. "localhost:7070"
	if err != nil {
		return errors.Wrap(err, "failed to create client")
	}
	l.Info("created client")

	// Allow disk IO to backlog a little, but not too much
	o := make(chan *client.OutputChunk, 50)
	go cl.GetObject(ctx, objPath, o) // e.g. "/foo/bar"
	var outFile *os.File
	if *outputPath != "" {
		l.Info("writing data to file: ", outputPath)
		made, err := os.OpenFile(*outputPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return errors.Wrap(err, "failed to open file")
		}
		outFile = made
		defer outFile.Close()
	}
	for chunk := range o {
		if chunk.Err != nil {
			return errors.Wrap(chunk.Err, "failed to fetch object")
		}
		if *outputPath != "" {
			l.Infof("writing %d bytes to file: %s", len(chunk.Data), *outputPath)
			written, err := outFile.Write(chunk.Data)
			if err != nil {
				return errors.Wrap(err, "failed to write data to file")
			}
			if written != len(chunk.Data) {
				return fmt.Errorf("short write %d of %d bytes", written, len(chunk.Data))
			}
		}
	}
	if *outputPath != "" {
		// Close even though a defer close is pending so we can return any OS
		// level errors at file close time (e.g. NFS only reports some errors at
		// close())
		if err := outFile.Close(); err != nil {
			return errors.Wrap(err, "failed to write data to file")
		}
	}
	endTime := time.Now()
	elapsed := endTime.Sub(startTime)
	l.Info("fetch complete; shutting down client")
	l.Info("elapsedTime: ", elapsed)
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 999*time.Second)
	defer shutdownCancel()
	if err := cl.Close(shutdownCtx); err != nil {
		return errors.Wrap(err, "failed to shut down client")
	}

	l.Info("completed without error")
	return nil
}
