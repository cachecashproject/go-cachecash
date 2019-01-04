package cache

import (
	"context"

	"github.com/kelleyk/go-cachecash/ccmsg"
	"github.com/pkg/errors"
)

type grpcClientCacheServer struct {
	cache *Cache
}

var _ ccmsg.ClientCacheServer = (*grpcClientCacheServer)(nil)

func (s *grpcClientCacheServer) GetBlock(ctx context.Context, req *ccmsg.ClientCacheRequest) (*ccmsg.ClientCacheResponseData, error) {
	resp, err := s.cache.HandleRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	// XXX: Should refactor to eliminate the need for this type conversion.
	switch resp := resp.Msg.(type) {
	case *ccmsg.ClientCacheResponse_DataResponse:
		return resp.DataResponse, nil
	case *ccmsg.ClientCacheResponse_Error:
		return nil, errors.New(resp.Error.Message)
	default:
		return nil, errors.New("unexpected response type")
	}
}

func (s *grpcClientCacheServer) ExchangeTicketL1(ctx context.Context, req *ccmsg.ClientCacheRequest) (*ccmsg.ClientCacheResponseL1, error) {
	resp, err := s.cache.HandleRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	// XXX: Should refactor to eliminate the need for this type conversion.
	switch resp := resp.Msg.(type) {
	case *ccmsg.ClientCacheResponse_L1Response:
		return resp.L1Response, nil
	case *ccmsg.ClientCacheResponse_Error:
		return nil, errors.New(resp.Error.Message)
	default:
		return nil, errors.New("unexpected response type")
	}
}

func (s *grpcClientCacheServer) ExchangeTicketL2(ctx context.Context, req *ccmsg.ClientCacheRequest) (*ccmsg.ClientCacheResponseL2, error) {
	resp, err := s.cache.HandleRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	// XXX: Should refactor to eliminate the need for this type conversion.
	// N.B.: This looks unlike the others because there is no return message.  We should remove the message and change the gRPC service definition.
	switch resp := resp.Msg.(type) {
	case *ccmsg.ClientCacheResponse_Error:
		return nil, errors.New(resp.Error.Message)
	default:
		return &ccmsg.ClientCacheResponseL2{}, nil
	}
}
