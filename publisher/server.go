package publisher

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/cachecashproject/go-cachecash/bootstrap"
	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
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
	GrpcAddr             string
	StatusAddr           string
	BootstrapAddr        string
	DefaultCacheDuration time.Duration
	Insecure             bool

	UpstreamURL string `json:"upstreamURL"`
	Database    string `json:"database"`
}

type application struct {
	l *logrus.Logger

	publisherServer *publisherServer
	statusServer    *statusServer
	// TODO: ...
}

var _ Application = (*application)(nil)

// XXX: Should this take p as an argument, or be responsible for setting it up?
func NewApplication(l *logrus.Logger, p *ContentPublisher, conf *ConfigFile) (Application, error) {
	publisherServer, err := newPublisherServer(l, p, conf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create publisher server")
	}

	statusServer, err := newStatusServer(l, p, conf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create status server")
	}

	return &application{
		l:               l,
		publisherServer: publisherServer,
		statusServer:    statusServer,
	}, nil
}

func (a *application) Start() error {
	if err := a.publisherServer.Start(); err != nil {
		return errors.Wrap(err, "failed to start publisher server")
	}
	if err := a.statusServer.Start(); err != nil {
		return errors.Wrap(err, "failed to start status server")
	}
	return nil
}

func (a *application) Shutdown(ctx context.Context) error {
	if err := a.publisherServer.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to shut down publisher server")
	}
	if err := a.statusServer.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to shut down status server")
	}
	return nil
}

type publisherServer struct {
	l              *logrus.Logger
	conf           *ConfigFile
	publisher      *ContentPublisher
	grpcServer     *grpc.Server
	httpServer     *http.Server
	cancelFunction context.CancelFunc
}

var _ common.StarterShutdowner = (*publisherServer)(nil)

func newPublisherServer(l *logrus.Logger, p *ContentPublisher, conf *ConfigFile) (*publisherServer, error) {
	grpcServer := common.NewGRPCServer()
	ccmsg.RegisterCachePublisherServer(grpcServer, &grpcPublisherServer{publisher: p})
	ccmsg.RegisterClientPublisherServer(grpcServer, &grpcPublisherServer{publisher: p})
	grpc_prometheus.Register(grpcServer)

	httpServer := wrapGrpc(grpcServer)

	return &publisherServer{
		l:          l,
		conf:       conf,
		publisher:  p,
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
}

func (s *publisherServer) Start() error {
	s.l.Info("publisherServer - Start - enter")

	lis, err := net.Listen("tcp", s.conf.GrpcAddr)
	if err != nil {
		return errors.Wrap(err, "failed to bind listener")
	}

	// httpLis, err := net.Listen("tcp", s.conf.ClientProtocolHttpAddr)
	httpLis, err := net.Listen("tcp", ":8043")
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
		if err := s.grpcServer.Serve(lis); err != nil {
			s.l.WithError(err).Error("failed to serve publisherServer(grpc)")
		}
	}()

	go func() {
		// This will block until we call `Stop`.
		if err := s.httpServer.Serve(httpLis); err != nil {
			s.l.WithError(err).Error("failed to serve publisherServer(http)")
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			caches, err := bootstrapClient.FetchCaches(ctx)
			if err != nil {
				s.l.Error("Failed to fetch caches: ", err)
			} else {
				s.l.Info("Caches: ", caches)

				if err := UpdateKnownCaches(ctx, s, caches); err != nil {
					s.l.Error("failed to update known caches: ", err)
				}

				if err := InitEscrows(ctx, s, caches); err != nil {
					s.l.Error("failed to init escrows: ", err)
				}
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

	s.l.Info("publisherServer - Start - exit")
	return nil
}

func (s *publisherServer) Shutdown(ctx context.Context) error {
	// stop fetching caches
	s.cancelFunction()

	// TODO: Should use `GracefulStop` until context expires, and then fall back on `Stop`.
	s.grpcServer.Stop()

	return nil
}
