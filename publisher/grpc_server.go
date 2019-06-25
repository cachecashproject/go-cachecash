package publisher

import (
	"context"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"go.opencensus.io/trace"
)

type grpcClientPublisherServer struct {
	publisher *ContentPublisher
}

var _ ccmsg.ClientPublisherServer = (*grpcClientPublisherServer)(nil)

func (s *grpcClientPublisherServer) GetContent(ctx context.Context, req *ccmsg.ContentRequest) (*ccmsg.ContentResponse, error) {
	ctx, span := trace.StartSpan(ctx, "cachecash.com/Publisher/GetContent")
	defer span.End()
	bundle, err := s.publisher.HandleContentRequest(ctx, req)
	if err != nil {
		s.publisher.l.WithError(err).Error("content request failed")
		return nil, err
	}

	// TODO: XXX: The sequence number(s) are used by some of the cryptography.  We can't just completely ignore those
	// fields after our move to gRPC.
	return &ccmsg.ContentResponse{
		// RequestSequenceNo: ... -- no longer necessary, since gRPC is handling RPC stuff for us
		Bundle: bundle,
	}, nil
}

type grpcCachePublisherServer struct {
	publisher *ContentPublisher
}

var _ ccmsg.CachePublisherServer = (*grpcCachePublisherServer)(nil)

func (s *grpcCachePublisherServer) CacheMiss(ctx context.Context, req *ccmsg.CacheMissRequest) (*ccmsg.CacheMissResponse, error) {
	return s.publisher.CacheMiss(ctx, req)
}
