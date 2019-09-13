package faucet

import (
	"context"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/ledger"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
)

type grpcFaucetServer struct {
	faucet *Faucet
}

var _ ccmsg.FaucetServer = (*grpcFaucetServer)(nil)

func (s *grpcFaucetServer) GetCoins(ctx context.Context, req *ccmsg.GetCoinsRequest) (*empty.Empty, error) {
	target, err := ledger.ParseAddress(req.Address)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode address")
	}

	err = s.faucet.SendCoins(ctx, target, 1337)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send coins")
	}

	return &empty.Empty{}, nil
}
