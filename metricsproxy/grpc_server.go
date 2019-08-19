package metricsproxy

import (
	"bytes"
	"encoding/hex"
	"errors"
	"io"
	"sync"
	"time"

	"golang.org/x/crypto/ed25519"

	"github.com/cachecashproject/go-cachecash/metrics"
	empty "github.com/golang/protobuf/ptypes/empty"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/sirupsen/logrus"
)

type grpcMetricsProxyServer struct {
	l *logrus.Logger
	// cache pubkey -> protobuf of metrics
	metrics map[string]*scrapeStatus
	lock    sync.Mutex
	// incremented when polled
	generation int64
}

type scrapeStatus struct {
	generation int64
	scrape     *metrics.Scrape
}

var _ metrics.MetricsServer = (*grpcMetricsProxyServer)(nil)

// var _ prometheus.Collector = (*grpcMetricsProxyServer)(nil)
var _ prometheus.Gatherer = (*grpcMetricsProxyServer)(nil)

func newGRPCMetricsProxyServer(l *logrus.Logger) *grpcMetricsProxyServer {
	return &grpcMetricsProxyServer{l: l, metrics: make(map[string]*scrapeStatus)}
}

// MetricsPoller collects metrics from caches and surfaces them to Prometheus
// To minimise the attack area, complex processing is deferred to actual scrapes by prometheus.
func (s *grpcMetricsProxyServer) MetricsPoller(srv metrics.Metrics_MetricsPollerServer) error {
	s.l.Debug("New MetricsPoller stream")
	ctx := srv.Context()
	// TODO: read the cache public key
	pubkey, err := srv.Recv()
	if err != nil {
		if err != io.EOF {
			s.l.Debugf("Failed to read public key %s", err)
		}
		return err
	}
	if pubkey.PublicKey.Keytype != metrics.KeyType_ED25519 {
		s.l.Debugf("Bad key type %s", pubkey.PublicKey.Keytype)
		return errors.New("Only ED25519 keys are supported")
	}
	clientkey := pubkey.PublicKey.PublicKey
	if len(clientkey) != 32 {
		s.l.Debugf("Bad key length %d", len(clientkey))
		return errors.New("invalid ed25519 key")
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		// Scrape every 5 seconds for now;
		// future work can involve:
		// - state based encodings
		// - differential encodings
		// polling when prometheus does
		case <-time.After(5 * time.Second):
		}
		err := srv.Send(&empty.Empty{})
		if err == io.EOF {
			return nil
		}
		if err != nil {
			s.l.Debugf("Err requesting scrape %s", err)
			return err
		}
		scrape, err := srv.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			s.l.Debugf("Err receiving scrape %s", err)
			return err
		}
		func() {
			s.lock.Lock()
			defer s.lock.Unlock()
			status, ok := s.metrics[(string)(clientkey)]
			if !ok {
				status = &scrapeStatus{}
				s.metrics[(string)(clientkey)] = status
			}
			status.generation = s.generation
			status.scrape = scrape
		}()
	}
}

// func (s *grpcMetricsProxyServer) Collect(output chan<- prometheus.Metric) {
func (s *grpcMetricsProxyServer) Gather() ([]*dto.MetricFamily, error) {
	var result []*dto.MetricFamily
	generation := func() int64 {
		s.lock.Lock()
		defer s.lock.Unlock()
		s.generation++
		return s.generation - 1
	}()

	for clientkey, scrape := range s.metrics {
		if generation-scrape.generation > 10 {
			// Discard very old scrapes to save memory
			// We could preserve the clientkey to report up=0 metrics
			if func() bool {
				s.lock.Lock()
				defer s.lock.Unlock()
				if generation-s.metrics[clientkey].generation <= 10 {
					return false
				}
				delete(s.metrics, clientkey)
				return true
			}() {
				continue
			}
		}
		key := (ed25519.PublicKey)(clientkey)
		hexkey := hex.EncodeToString([]byte(clientkey[:]))
		labelname := "clientkey"
		if !ed25519.Verify(key, scrape.scrape.Data, scrape.scrape.Signature) {
			// incorrectly signed scrape.
			// TODO: track this in a metric
			continue
		}
		dec := expfmt.NewDecoder(bytes.NewBuffer(scrape.scrape.Data), expfmt.FmtProtoDelim)
		var err error
		for {
			d := dto.MetricFamily{}
			if err = dec.Decode(&d); err != nil {
				break
			}
			name := d.GetName()
			if len(name) == 0 {
				continue
			}
			// TODO: build a whitelist of metric features (name-types, label characteristics)that we want to collect)
			// TODO: consider rejecting caches not participating in the network
			// add an enforced label identifying the client to all the things
			// TODO: add an 'up' series for caches where is within 1? 2? of generation
			for _, m := range d.Metric {
				m.Label = append(m.Label, &dto.LabelPair{Name: &labelname, Value: &hexkey})
			}
			result = append(result, &d)
		}

		if err != io.EOF {
			// TODO: should we send this case back to the client somewhere or track it as a metric?
			continue
		}

	}
	return result, nil
}

// Describe the metrics we collect. As this is dynmaic, at least for now, we punt per the docs:
// ---
// Sending no descriptor at all marks the Collector as “unchecked”,
// i.e. no checks will be performed at registration time, and the
// Collector may yield any Metric it sees fit in its Collect method.
// ---
// func (s *grpcMetricsProxyServer) Describe(chan<- *prometheus.Desc) {
// }
