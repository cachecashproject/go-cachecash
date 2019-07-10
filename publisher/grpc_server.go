package publisher

import (
	"context"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"go.opencensus.io/trace"
)

type grpcPublisherServer struct {
	publisher *ContentPublisher
}

var _ ccmsg.ClientPublisherServer = (*grpcPublisherServer)(nil)

func (s *grpcPublisherServer) GetContent(ctx context.Context, req *ccmsg.ContentRequest) (*ccmsg.ContentResponse, error) {
	ctx, span := trace.StartSpan(ctx, "cachecash.com/Publisher/GetContent")
	defer span.End()
	bundles, err := s.publisher.HandleContentRequest(ctx, req)
	if err != nil {
		s.publisher.l.WithError(err).Error("content request failed")
		return nil, err
	}

	// TODO: XXX: The sequence number(s) are used by some of the cryptography.  We can't just completely ignore those
	// fields after our move to gRPC.
	return &ccmsg.ContentResponse{
		// RequestSequenceNo: ... -- no longer necessary, since gRPC is handling RPC stuff for us
		Bundles: bundles,
	}, nil
}

var _ ccmsg.CachePublisherServer = (*grpcPublisherServer)(nil)

func (s *grpcPublisherServer) CacheMiss(ctx context.Context, req *ccmsg.CacheMissRequest) (*ccmsg.CacheMissResponse, error) {
	return s.publisher.cacheMiss(ctx, req)
}
