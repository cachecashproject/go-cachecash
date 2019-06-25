package common

import (
	"contrib.go.opencensus.io/exporter/jaeger"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/trace"
)

// Flushable is a wrapper for jaeger.Exporter to permit clean code in callers
// while permitting tracing to be disabled.
type Flushable struct {
	exporter *jaeger.Exporter
}

// Flush blocks until all recent spans are flushed or a hard error occurs
// flushing (which is then swallowed).
func (f *Flushable) Flush() {
	if f == nil {
		return
	}
	f.exporter.Flush()
}

// SetupTracing configures OpenCensus to export to Jaeger.
func SetupTracing(traceAPI string, serviceName string, l *logrus.Logger) *Flushable {
	if len(traceAPI) == 0 {
		return nil
	}
	collectorEndpointURI := traceAPI + "/api/traces"

	je, err := jaeger.NewExporter(jaeger.Options{
		CollectorEndpoint: collectorEndpointURI,
		ServiceName:       serviceName,
	})
	if err != nil {
		l.Fatalf("Failed to create the Jaeger exporter: %v", err)
	}

	trace.RegisterExporter(je)
	return &Flushable{exporter: je}
}
