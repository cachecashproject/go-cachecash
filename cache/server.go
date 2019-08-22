package cache

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cachecashproject/go-cachecash/bootstrap"
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/keypair"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// An Application is the top-level content publisher.  It takes a configuration struct.  Its children are the several
// protocol servers (that deal with clients, caches, and so forth).
type Application interface {
	common.StarterShutdowner
}

type ConfigFile struct {
	ClientProtocolGrpcAddr string
	ClientProtocolHttpAddr string
	StatusAddr             string
	BootstrapAddr          string

	BadgerDirectory string `json:"badger_directory"`
	Database        string `json:"database"`
	ContactUrl      string `json:"contact_url"`
	MetricsEndpoint string `json:"metrics_endpoint"`
	Insecure        bool   `json:"insecure"`
}

type application struct {
	l *logrus.Logger

	clientProtocolServer *clientProtocolServer
	statusServer         *statusServer
	metricsPush          *common.MetricsPusher
	// TODO: ...
}

var _ Application = (*application)(nil)

func NewApplication(l *logrus.Logger, c *Cache, conf *ConfigFile, kp *keypair.KeyPair) (Application, error) {
	clientProtocolServer, err := newClientProtocolServer(l, c, conf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client protocol server")
	}

	statusServer, err := newStatusServer(l, c, conf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create status server")
	}

	metricsPush := common.NewMetricsPusher(l, conf.MetricsEndpoint, conf.Insecure, kp)

	return &application{
		l:                    l,
		clientProtocolServer: clientProtocolServer,
		statusServer:         statusServer,
		metricsPush:          metricsPush,
	}, nil
}

func (a *application) Start() error {
	if err := a.clientProtocolServer.Start(); err != nil {
		return errors.Wrap(err, "failed to start client protocol server")
	}
	if err := a.statusServer.Start(); err != nil {
		return errors.Wrap(err, "failed to start status server")
	}
	if err := a.metricsPush.Start(); err != nil {
		return errors.Wrap(err, "failed to start metrics pusher")
	}
	return nil
}

func (a *application) Shutdown(ctx context.Context) error {
	if err := a.clientProtocolServer.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to shut down client protocol server")
	}
	if err := a.statusServer.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to shut down status server")
	}
	if err := a.metricsPush.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to shut down metrics push")
	}
	return nil
}

type clientProtocolServer struct {
	l              *logrus.Logger
	conf           *ConfigFile
	cache          *Cache
	grpcServer     *grpc.Server
	httpServer     *http.Server
	cancelFunction context.CancelFunc
}

var _ common.StarterShutdowner = (*clientProtocolServer)(nil)

func newClientProtocolServer(l *logrus.Logger, c *Cache, conf *ConfigFile) (*clientProtocolServer, error) {
	grpcServer := common.NewGRPCServer()
	ccmsg.RegisterClientCacheServer(grpcServer, &grpcClientCacheServer{cache: c})
	ccmsg.RegisterPublisherCacheServer(grpcServer, &grpcPublisherCacheServer{cache: c})
	grpc_prometheus.Register(grpcServer)

	httpServer := wrapGrpc(grpcServer)

	return &clientProtocolServer{
		l:          l,
		conf:       conf,
		cache:      c,
		grpcServer: grpcServer,
		httpServer: httpServer,
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

	/*
		if *enableTls {
			if err := httpServer.ListenAndServeTLS(*tlsCertFilePath, *tlsKeyFilePath); err != nil {
				grpclog.Fatalf("failed starting http2 server: %v", err)
			}
		} else {
			if err := httpServer.ListenAndServe(); err != nil {
				grpclog.Fatalf("failed starting http server: %v", err)
			}
		}
	*/
}

func (s *clientProtocolServer) Start() error {
	s.l.Info("clientProtocolServer - Start - enter")

	addrParts := strings.Split(s.conf.ClientProtocolGrpcAddr, ":")
	portStr := addrParts[len(addrParts)-1]
	port, err := strconv.ParseUint(portStr, 10, 32)
	if err != nil {
		return errors.Wrap(err, "failed to get port from client protocol grpc addr")
	}

	grpcLis, err := net.Listen("tcp", s.conf.ClientProtocolGrpcAddr)
	if err != nil {
		return errors.Wrap(err, "failed to bind listener")
	}

	httpLis, err := net.Listen("tcp", s.conf.ClientProtocolHttpAddr)
	if err != nil {
		return errors.Wrap(err, "failed to bind listener")
	}

	// TODO: BootstrapAddr should be optional
	bootstrapClient, err := bootstrap.NewClient(s.l, s.conf.BootstrapAddr, s.conf.Insecure)
	if err != nil {
		return errors.Wrap(err, "failed to create bootstrap client")
	}

	go func() {
		// This will block until we call `Stop`.
		if err := s.grpcServer.Serve(grpcLis); err != nil {
			s.l.WithError(err).Error("failed to serve clientProtocolServer(grpc)")
		}
	}()

	go func() {
		// This will block until we call `Stop`.
		if err := s.httpServer.Serve(httpLis); err != nil {
			s.l.WithError(err).Error("failed to serve clientProtocolServer(http)")
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			stats := bootstrap.NewCacheStats()
			stats.ReadMemoryStats()
			err = stats.ReadDiskStats(s.cache.StoragePath)
			if err != nil {
				s.l.Error("failed to read disk stats: ", err)
				continue
			}
			err = bootstrapClient.AnnounceCache(context.Background(), bootstrap.BootstrapInfo{
				PublicKey:   s.cache.PublicKey,
				Stats:       stats,
				StartupTime: s.cache.StartupTime,
				Port:        uint32(port),
				ContactUrl:  s.conf.ContactUrl,
			})
			if err != nil {
				s.l.Error("failed to announce cache: ", err)
			}

			select {
			// if a shutdown has been requested close the go channel
			case <-ctx.Done():
				return
			// after we waited for a shutdown request for x minutes, announce the cache again
			case <-time.After(1 * time.Minute):
				continue
			}
		}
	}()
	s.cancelFunction = cancel

	s.l.Info("clientProtocolServer - Start - exit")
	return nil
}

func (s *clientProtocolServer) Shutdown(ctx context.Context) error {
	// stop announcing our cache
	s.cancelFunction()

	// TODO: Should use `GracefulStop` until context expires, and then fall back on `Stop`.
	s.grpcServer.Stop()

	return s.httpServer.Shutdown(ctx)
}
