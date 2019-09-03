package faucet

import (
	"context"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ed25519"
)

type grpcFaucetServer struct {
	faucet *Faucet
}

var _ ccmsg.FaucetServer = (*grpcFaucetServer)(nil)

func (s *grpcFaucetServer) GetCoins(ctx context.Context, req *ccmsg.GetCoinsRequest) (*ccmsg.GetCoinsResponse, error) {
	/*
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
	*/

	target := ed25519.PublicKey(req.PublicKey.PublicKey)
	err := s.faucet.SendCoins(ctx, target, 1337)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send coins")
	}

	return &ccmsg.GetCoinsResponse{}, nil
}
