package ledgerservice

import (
	"context"
	"errors"

	"github.com/cachecashproject/go-cachecash/ccmsg"
)

type grpcLedgerServer struct {
	ledgerService *LedgerService
}

var _ ccmsg.LedgerServer = (*grpcLedgerServer)(nil)

func (s *grpcLedgerServer) PostTransaction(ctx context.Context, req *ccmsg.PostTransactionRequest) (*ccmsg.PostTransactionResponse, error) {
	return nil, errors.New("not implemented")
}

func (s *grpcLedgerServer) GetBlocks(ctx context.Context, req *ccmsg.GetBlocksRequest) (*ccmsg.GetBlocksResponse, error) {
	return nil, errors.New("not implemented")
}
