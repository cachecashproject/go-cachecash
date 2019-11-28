package common

import (
	"database/sql"

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
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(nil)))
	}
	return grpc.Dial(target, opts...)
}

// NewGRPCServer makes a stateless GRPC server preconfigured with tracing and
// monitoring middleware.
func NewGRPCServer(opt ...grpc.ServerOption) *grpc.Server {
	return grpc.NewServer(
		append([]grpc.ServerOption{
			grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(grpc_prometheus.StreamServerInterceptor)),
			grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(grpc_prometheus.UnaryServerInterceptor)),
			grpc.StatsHandler(&ocgrpc.ServerHandler{})},
			opt...)...)
}

// New GRPCServer makes a DB enabled GRPC server preconfigured with tracing and
// monitoring middleware.
func NewDBGRPCServer(db *sql.DB, opt ...grpc.ServerOption) *grpc.Server {
	injector := dbtx.NewInjector(db)
	return grpc.NewServer(
		append([]grpc.ServerOption{
			grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(grpc_prometheus.StreamServerInterceptor, injector.StreamServerInterceptor())),
			grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(grpc_prometheus.UnaryServerInterceptor, injector.UnaryServerInterceptor())),
			grpc.StatsHandler(&ocgrpc.ServerHandler{})},
			opt...)...)
}
