package common

import (
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
)

// GRPCDial creates a client connection to the given target.
func GRPCDial(target string, insecure bool, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts = append([]grpc.DialOption{
		grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor),
		grpc.WithStatsHandler(&ocgrpc.ClientHandler{})},
		opts...)
	if insecure || true { // disabled until config is rolled out
		opts = append(opts, grpc.WithInsecure())
	}
	return grpc.Dial(target, opts...)
}

func NewGRPCServer(opt ...grpc.ServerOption) *grpc.Server {
	return grpc.NewServer(
		append([]grpc.ServerOption{
			grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
			grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
			grpc.StatsHandler(&ocgrpc.ServerHandler{})},
			opt...)...)
}
