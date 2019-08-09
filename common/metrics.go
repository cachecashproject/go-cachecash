package common

import (
	"bytes"
	"context"
	"crypto"
	"io"
	"time"

	"github.com/cachecashproject/go-cachecash/keypair"
	"github.com/cachecashproject/go-cachecash/metrics"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"

	"github.com/sirupsen/logrus"
)

// MetricsPusher sends metrics to the CacheCash central metrics store
type MetricsPusher struct {
	l        *logrus.Logger
	endpoint string
	kp       *keypair.KeyPair

	client metrics.MetricsClient
	cancel context.CancelFunc
}

// NewMetricsPusher creates a new MetricsPusher
func NewMetricsPusher(l *logrus.Logger, endpoint string, kp *keypair.KeyPair) *MetricsPusher {
	return &MetricsPusher{endpoint: endpoint, l: l, kp: kp}
}

// Start sending metrics
func (mp *MetricsPusher) Start() error {
	if len(mp.endpoint) == 0 {
		return nil
	}
	mp.l.Infof("Pushing metrics to %s", mp.endpoint)
	// Lazy dial - subconnections will reconnect over time; this will error
	// synchronously if the endpoint is syntactically incorrect.
	conn, err := GRPCDial(mp.endpoint)
	if err != nil {
		return errors.Wrap(err, "failed to dial metrics service")
	}

	mp.client = metrics.NewMetricsClient(conn)
	var stopCtx context.Context
	stopCtx, mp.cancel = context.WithCancel(context.Background())
	go mp.pushMetrics(stopCtx)
	return nil
}

// Shutdown sending of metrics
func (mp *MetricsPusher) Shutdown(ctx context.Context) error {
	if len(mp.endpoint) == 0 {
		return nil
	}
	mp.cancel()
	// Don't bother waiting for shutdown - there's no valuable data to dequeue
	return nil
}

// push metrics to the endpoint - handles session setup, reconnecting etc.
func (mp *MetricsPusher) pushMetrics(ctx context.Context) {
	// Outer loop - until cancelled, try once every 5 seconds (arbitrary figure,
	// should add jitter and backoff)
	for {
		mp.pushMetricsOneSession(ctx)
		select {
		case <-ctx.Done():
			mp.l.Errorf("Context done %s", ctx.Err())
		case <-time.After(5 * time.Second):
			mp.l.Debug("Retrying metrics session")
		}
	}
}

// push metrics for one server session
func (mp *MetricsPusher) pushMetricsOneSession(ctx context.Context) {
	stream, err := mp.client.MetricsPoller(ctx)
	if err != nil {
		mp.l.Errorf("Failed to open metrics poller channel %s", err)
		return
	}
	// Disable retries - we don't want any backlogs or hidden queues.
	_ = stream.Context()

	err = stream.Send(&metrics.Scrape{PublicKey: &metrics.PublicKey{PublicKey: mp.kp.PublicKey, Keytype: metrics.KeyType_ED25519}})
	if err != nil {
		mp.l.Errorf("Failed to set cache public key %s", err)
		return
	}

	for {
		_, err = stream.Recv()
		if err != nil {
			if err != io.EOF {
				mp.l.Errorf("Error waiting for metrics poll %s", err)
			} else {
				mp.l.Debug("EOF in metrics poll")
			}
			return
		}
		scrape := mp.scrape()
		// N.B.: See the godoc for `crypto/ed25519` for a discussion of the parameters to this call.  Passing nil as the
		// first argument makes Sign use crypto/rand.Reader for entropy.
		signature, err := mp.kp.PrivateKey.Sign(nil, scrape, crypto.Hash(0))
		if err != nil {
			mp.l.Errorf("Error signing metrics scrape %s", err)
			return
		}
		err = stream.Send(&metrics.Scrape{Data: scrape, Signature: signature})
		if err != nil {
			if err != io.EOF {
				mp.l.Errorf("Error sending metrics scrape %s", err)
			} else {
				mp.l.Debug("EOF sending scrape")
			}
			return
		}
	}
}

// scrape prometheus
// TODO: generate and send synthetic error metrics if scraping fails
func (mp *MetricsPusher) scrape() []byte {
	reg := prometheus.DefaultGatherer
	mfs, err := reg.Gather()
	if err != nil {
		mp.l.Errorf("Error scraping prometheus metrics: %s", err)
		return nil
	}
	contentType := expfmt.FmtProtoDelim
	var b bytes.Buffer

	enc := expfmt.NewEncoder(&b, contentType)

	var lastErr error
	for _, mf := range mfs {
		if err := enc.Encode(mf); err != nil {
			lastErr = err
		}
	}

	if lastErr != nil {
		mp.l.Errorf("Error scraping prometheus metrics: %s", lastErr)
		return nil
	}
	return b.Bytes()
}
