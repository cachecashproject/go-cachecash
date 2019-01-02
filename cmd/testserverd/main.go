package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	cachecash "github.com/kelleyk/go-cachecash"
	"github.com/kelleyk/go-cachecash/cache"
	"github.com/kelleyk/go-cachecash/ccmsg"
	"github.com/kelleyk/go-cachecash/common"
	"github.com/kelleyk/go-cachecash/provider"
	"github.com/kelleyk/go-cachecash/testutil"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ed25519"
)

/*
testserverd runs a content provider and a set of caches.  Running them in a single process is more convenient for
development purposes, but most importantly lets us proceed before cache/provider interactions are fleshed out.

TODO:
- We need a way to signal how large the object is, so that the client knows how many requests to make.
- Using the `cachecash-curl` binary to fetch an object that doesn't exist should return a 404 error.
- Should actually serve the requested object instead of random data.
- Eventually, this and the `testserverd_randomdata` binary should share most of their code.
*/

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

	// Create a provider.
	_, providerPrivateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return errors.Wrap(err, "failed to generate provider keypair")
	}
	prov, err := provider.NewContentProvider(ts.l, providerPrivateKey)
	if err != nil {
		return err
	}
	ts.provider = prov

	// Create escrow and add it to the provider.
	escrow, err := prov.NewEscrow(&ccmsg.EscrowInfo{
		DrawDelay:       5,
		ExpirationDelay: 5,
		StartBlock:      42,
		TicketsPerBlock: []*ccmsg.Segment{
			&ccmsg.Segment{Length: 10, Value: 100},
		},
	})
	if err != nil {
		return err
	}
	if err := prov.AddEscrow(escrow); err != nil {
		return err
	}
	ts.escrow = escrow

	// Create a content object.
	obj, err := cachecash.RandomContentBuffer(16, 128*1024) // 16 blocks of 128 KiB
	if err != nil {
		return err
	}
	escrow.Objects["/foo/bar"] = provider.EscrowObjectInfo{
		Object: obj,
		// N.B.: This becomes BundleParams.ObjectID
		// XXX: What's that used for?  Are it and BlockIdx used to uniquely identify blocks on the cache?  I think we're
		// using content addressing, aren't we?
		ID: 999,
	}
	ts.obj = obj

	// Create caches that are participating in this escrow.
	var caches []*cache.Cache
	for i := 0; i < 4; i++ {
		public, private, err := ed25519.GenerateKey(nil)
		if err != nil {
			return errors.Wrap(err, "failed to generate cache keypair")
		}

		// XXX: generate master-key
		innerMasterKey := testutil.RandBytes(16)

		escrow.Caches = append(escrow.Caches, &provider.ParticipatingCache{
			InnerMasterKey: innerMasterKey,
			PublicKey:      public,
			Inetaddr:       net.ParseIP("127.0.0.1"),
			Port:           uint32(9000 + i),
		})

		c := &cache.Cache{
			Escrows: make(map[ccmsg.EscrowID]*cache.Escrow),
		}
		ce := &cache.Escrow{
			InnerMasterKey: innerMasterKey,
			OuterMasterKey: testutil.RandBytes(16),
			Objects:        make(map[uint64]cachecash.ContentObject),
		}
		ce.Objects[999] = obj
		c.Escrows[*(escrow.ID())] = ce
		caches = append(caches, c)
		_ = private
	}
	ts.caches = caches

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
	ps, err := provider.NewApplication(ts.l, ts.provider, &provider.Config{
		ClientProtocolAddr: ":8080",
	})
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
