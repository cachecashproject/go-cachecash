package metricsproxy

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type statusServer struct {
	l          *logrus.Logger
	conf       *ConfigFile
	httpServer *http.Server
}

func newStatusServer(l *logrus.Logger, conf *ConfigFile, metricsServer *grpcMetricsProxyServer) (*statusServer, error) {
	mux := http.NewServeMux()

	httpServer := &http.Server{
		Addr:    conf.StatusAddr,
		Handler: mux,

		// // XXX: These are arbitrary; taken from godoc.
		// ReadTimeout:    10 * time.Second,
		// WriteTimeout:   10 * time.Second,
		// MaxHeaderBytes: 1 << 20,
	}

	s := &statusServer{
		l:          l,
		conf:       conf,
		httpServer: httpServer,
	}

	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/metrics/proxied",
		promhttp.InstrumentMetricHandler(
			prometheus.DefaultRegisterer,
			promhttp.HandlerFor(metricsServer, promhttp.HandlerOpts{}),
		))

	return s, nil
}

func (s *statusServer) Start() error {
	go func() {
		// XXX: This will probably need to be improved to allow for graceful shutdown, and to allow the program to abort
		// if unable to listen.
		if err := s.httpServer.ListenAndServe(); err != nil {
			s.l.WithError(err).Warn("statusServer.ListenAndServe() returned error")
		}
	}()

	return nil
}

func (s *statusServer) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
