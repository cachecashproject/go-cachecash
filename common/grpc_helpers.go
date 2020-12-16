package common

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"io/ioutil"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/cachecashproject/go-cachecash/dbtx"
)

// GRPCDial creates a client connection to the given target.
func GRPCDial(target string, insecure bool, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts = append([]grpc.DialOption{
		grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor),
		grpc.WithStatsHandler(&ocgrpc.ClientHandler{})},
		opts...)
	if insecure {
		opts = append(opts, grpc.WithInsecure())
	} else {
		config := &tls.Config{
			InsecureSkipVerify: false,
		}
		ca, err := ioutil.ReadFile("/tls/ca.pem")
		if err == nil {
			certPool := x509.NewCertPool()
			if certPool.AppendCertsFromPEM(ca) {
				config.RootCAs = certPool
			}
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(config)))
	}
	return grpc.Dial(target, opts...)
}

// NewGRPCServer makes a stateless GRPC server preconfigured with tracing and
// monitoring middleware.
func NewGRPCServer(insecure bool, opt ...grpc.ServerOption) *grpc.Server {
	opts := append([]grpc.ServerOption{
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(grpc_prometheus.StreamServerInterceptor)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(grpc_prometheus.UnaryServerInterceptor)),
		grpc.StatsHandler(&ocgrpc.ServerHandler{})},
		opt...)
	if !insecure {
		creds, err := credentials.NewServerTLSFromFile("/tls/server-cert.pem", "/tls/server-key.pem")
		if err == nil {
			opts = append(opts, grpc.Creds(creds))
		}
	}
	return grpc.NewServer(opts...)
}

// New GRPCServer makes a DB enabled GRPC server preconfigured with tracing and
// monitoring middleware.
func NewDBGRPCServer(insecure bool, db *sql.DB, opt ...grpc.ServerOption) *grpc.Server {
	injector := dbtx.NewInjector(db)
	opts := append([]grpc.ServerOption{
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(grpc_prometheus.StreamServerInterceptor, injector.StreamServerInterceptor())),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(grpc_prometheus.UnaryServerInterceptor, injector.UnaryServerInterceptor())),
		grpc.StatsHandler(&ocgrpc.ServerHandler{})},
		opt...)
	if !insecure {
		creds, err := credentials.NewServerTLSFromFile("/tls/server-cert.pem", "/tls/server-key.pem")
		if err == nil {
			opts = append(opts, grpc.Creds(creds))
		}
	}
	return grpc.NewServer(opts...)
}
