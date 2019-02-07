package main

import (
	"context"
	"fmt"

	cachecash "github.com/cachecashproject/go-cachecash"
	"github.com/cachecashproject/go-cachecash/cache"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/publisher"
	"github.com/cachecashproject/go-cachecash/testdatagen"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// testserverd runs a content publisher and a set of caches.  Running them in a single process is more convenient for
// development purposes, but most importantly lets us proceed before cache/publisher interactions are fleshed out.

type TestServer struct {
	l *logrus.Logger

	publisher *publisher.ContentPublisher
	escrow   *publisher.Escrow
	obj      cachecash.ContentObject
	caches   []*cache.Cache

	publisherApp publisher.Application
	cacheApps   []cache.Application
}

// XXX: Cribbed from `integration_test.go`.
func (ts *TestServer) setup() error {
	ts.l = logrus.New()
	ts.l.SetLevel(logrus.DebugLevel)

	scen, err := testdatagen.GenerateTestScenario(ts.l, &testdatagen.TestScenarioParams{
		BlockSize:  128 * 1024,
		ObjectSize: 128 * 1024 * 16,
	})
	if err != nil {
		return err
	}

	ts.publisher = scen.Publisher
	ts.escrow = scen.Escrow
	ts.obj = scen.Obj
	ts.caches = scen.Caches

	return nil
}

func (ts *TestServer) Start() error {
	ps, err := publisher.NewApplication(ts.l, ts.publisher, &publisher.Config{})
	if err != nil {
		return errors.Wrap(err, "failed to create publisher application")
	}
	if err := ps.Start(); err != nil {
		return errors.Wrap(err, "failed to start publisher application")
	}
	ts.publisherApp = ps

	for i, c := range ts.caches {
		cs, err := cache.NewApplication(ts.l, c, &cache.Config{
			// XXX: This must match what is set up in the Escrow struct on the publisher side so that the publisher sends
			// clients to the right place.
			ClientProtocolAddr: fmt.Sprintf(":%v", 9000+i),
			StatusAddr:         fmt.Sprintf(":%v", 9100+i),
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
	if err := ts.publisherApp.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to shut down publisher application")
	}
	for _, a := range ts.cacheApps {
		if err := a.Shutdown(ctx); err != nil {
			return errors.Wrap(err, "failed to shut down cache application")
		}
	}
	return nil
}

func main() {
	ts := &TestServer{}
	if err := ts.setup(); err != nil {
		panic(err)
	}
	if err := common.RunStarterShutdowner(ts); err != nil {
		panic(err)
	}
}
