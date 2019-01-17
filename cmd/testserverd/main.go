package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	cachecash "github.com/kelleyk/go-cachecash"
	"github.com/kelleyk/go-cachecash/cache"
	"github.com/kelleyk/go-cachecash/common"
	"github.com/kelleyk/go-cachecash/provider"
	"github.com/kelleyk/go-cachecash/testdatagen"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

/*
testserverd runs a content provider and a set of caches.  Running them in a single process is more convenient for
development purposes, but most importantly lets us proceed before cache/provider interactions are fleshed out.

The intent is that
  $ ./bin/cachecash-curl cachecash://localhost:8080/file0.bin
should return the same output as
  $ curl http://localhost:8081/file0.bin
(The CacheCash provider runs on port 8080; the HTTP upstream that it pulls content from runs on port 8081.)

TODO:
- Using the `cachecash-curl` binary to fetch an object that doesn't exist should return a 404 error.
- Eventually, this and the `testserverd_randomdata` binary should share most of their code.
*/

var (
	logLevelStr = flag.String("logLevel", "info", "Verbosity of log output")
)

type TestServer struct {
	l *logrus.Logger

	conf *Config

	provider *provider.ContentProvider
	escrow   *provider.Escrow
	obj      cachecash.ContentObject
	caches   []*cache.Cache

	providerApp provider.Application
	cacheApps   []cache.Application

	originServer *http.Server
}

type Config struct {
	DataPath string
}

// XXX: Cribbed from `integration_test.go`.
func (ts *TestServer) setup() error {
	ts.l = logrus.New()

	logLevel, err := logrus.ParseLevel(*logLevelStr)
	if err != nil {
		return errors.Wrap(err, "failed to parse log level")
	}
	ts.l.SetLevel(logLevel)

	scen, err := testdatagen.GenerateTestScenario(ts.l, &testdatagen.TestScenarioParams{
		L:          ts.l,
		BlockSize:  128 * 1024,
		ObjectSize: 128 * 1024 * 16,
	})
	if err != nil {
		return err
	}

	ts.provider = scen.Provider
	ts.escrow = scen.Escrow
	ts.obj = scen.Obj
	ts.caches = scen.Caches

	return nil
}

func (ts *TestServer) Start() error {
	// Start origin HTTP server.
	ts.originServer = &http.Server{
		Addr:    ":8081",
		Handler: http.FileServer(http.Dir(ts.conf.DataPath)),
	}
	go func() {
		// XXX: This will probably need to be improved to allow for graceful shutdown, and to allow the program to abort
		// if unable to listen.
		if err := ts.originServer.ListenAndServe(); err != nil {
			ts.l.WithError(err).Warn("originServer.ListenAndServe() returned error")
		}
	}()

	// Start provider.
	ps, err := provider.NewApplication(ts.l, ts.provider, &provider.Config{})
	if err != nil {
		return errors.Wrap(err, "failed to create provider application")
	}
	if err := ps.Start(); err != nil {
		return errors.Wrap(err, "failed to start provider application")
	}
	ts.providerApp = ps

	// Start caches.
	for i, c := range ts.caches {
		cs, err := cache.NewApplication(ts.l, c, &cache.Config{
			// XXX: This must match what is set up in the Escrow struct on the provider side so that the provider sends
			// clients to the right place.
			ClientProtocolAddr: fmt.Sprintf(":%v", 9000+i),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create cache application")
		}
		if err := cs.Start(); err != nil {
			return errors.Wrap(err, "failed to start cache application")
		}
		ts.cacheApps = append(ts.cacheApps, cs)
	}

	return nil
}

func (ts *TestServer) Shutdown(ctx context.Context) error {
	// TODO: @kelleyk: Should do these simultaneously, not sequentially; I have code in my personal library that does this
	if err := ts.providerApp.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to shut down provider application")
	}
	for _, a := range ts.cacheApps {
		if err := a.Shutdown(ctx); err != nil {
			return errors.Wrap(err, "failed to shut down cache application")
		}
	}
	return nil
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	ts := &TestServer{
		conf: &Config{
			DataPath: "./testdata/content",
		},
	}
	if err := ts.setup(); err != nil {
		panic(err)
	}
	if err := common.RunStarterShutdowner(ts); err != nil {
		panic(err)
	}
}
