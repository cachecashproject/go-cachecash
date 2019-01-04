package provider

import (
	"context"

	"github.com/kelleyk/go-cachecash/ccmsg"
)

type grpcClientProviderServer struct {
	provider *ContentProvider
}

var _ ccmsg.ClientProviderServer = (*grpcClientProviderServer)(nil)

func (s *grpcClientProviderServer) GetContent(ctx context.Context, req *ccmsg.ContentRequest) (*ccmsg.ContentResponse, error) {
	bundle, err := s.provider.HandleContentRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	// TODO: XXX: The sequence number(s) are used by some of the cryptography.  We can't just completely ignore those
	// fields after our move to gRPC.
	return &ccmsg.ContentResponse{
		// RequestSequenceNo: ... -- no longer necessary, since gRPC is handling RPC stuff for us
		Bundle: bundle,
	}, nil
}

type grpcCacheProviderServer struct {
	provider *ContentProvider
}

var _ ccmsg.CacheProviderServer = (*grpcCacheProviderServer)(nil)

func (s *grpcCacheProviderServer) CacheMiss(ctx context.Context, req *ccmsg.CacheMissRequest) (*ccmsg.CacheMissResponse, error) {
	return s.provider.CacheMiss(ctx, req)
}
