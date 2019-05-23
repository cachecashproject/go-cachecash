package cache

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type statusServer struct {
	l          *logrus.Logger
	conf       *ConfigFile
	cache      *Cache
	httpServer *http.Server
}

func newStatusServer(l *logrus.Logger, c *Cache, conf *ConfigFile) (*statusServer, error) {
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
		cache:      c,
		httpServer: httpServer,
	}

	mux.HandleFunc("/info", s.handleInfo)
	mux.Handle("/metrics", promhttp.Handler())

	return s, nil
}

type infoResponse struct {
	Escrows [][]byte
}

func (s *statusServer) handleInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	escrows := [][]byte{}
	for escrowID := range s.cache.Escrows {
		escrows = append(escrows, escrowID[:])
	}

	resp := &infoResponse{
		Escrows: escrows,
	}

	d, err := json.Marshal(resp)
	if err != nil {
		s.l.WithError(err).Error("failed to marshal response JSON")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(d); err != nil {
		s.l.WithError(err).Error("failed to write response")
	}
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
