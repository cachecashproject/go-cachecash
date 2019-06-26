package common

import (
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
)

// XXX: No transport security!
// GRPCDial creates a client connection to the given target.
func GRPCDial(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.Dial(target,
		append([]grpc.DialOption{
			grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
			grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor),
			grpc.WithStatsHandler(&ocgrpc.ClientHandler{}),
			grpc.WithInsecure()},
			opts...)...)
}

func NewGRPCServer(opt ...grpc.ServerOption) *grpc.Server {
	return grpc.NewServer(
		append([]grpc.ServerOption{
			grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
			grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
			grpc.StatsHandler(&ocgrpc.ServerHandler{})},
			opt...)...)
}
