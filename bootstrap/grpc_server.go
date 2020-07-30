package bootstrap

import (
	"context"

	"github.com/cachecashproject/go-cachecash/ccmsg"
)

type grpcBootstrapServer struct {
	bootstrap *Bootstrapd
}

var _ ccmsg.NodeBootstrapdServer = (*grpcBootstrapServer)(nil)

func (s *grpcBootstrapServer) AnnounceCache(ctx context.Context, req *ccmsg.CacheAnnounceRequest) (*ccmsg.CacheAnnounceResponse, error) {
	return s.bootstrap.HandleCacheAnnounceRequest(ctx, req)
}

func (s *grpcBootstrapServer) FetchCaches(ctx context.Context, req *ccmsg.CacheFetchRequest) (*ccmsg.CacheFetchResponse, error) {
	return s.bootstrap.HandleCacheFetchRequest(ctx, req)
}
