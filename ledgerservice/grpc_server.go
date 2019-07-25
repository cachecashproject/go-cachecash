package ledgerservice

import (
	"context"

	"github.com/cachecashproject/go-cachecash/ccmsg"
)

type grpcLedgerServer struct {
	ledgerService *LedgerService
}

var _ ccmsg.LedgerServer = (*grpcLedgerServer)(nil)

func (s *grpcLedgerServer) PostTransaction(ctx context.Context, req *ccmsg.PostTransactionRequest) (*ccmsg.PostTransactionResponse, error) {
	return s.ledgerService.PostTransaction(ctx, req)
}

func (s *grpcLedgerServer) GetBlocks(ctx context.Context, req *ccmsg.GetBlocksRequest) (*ccmsg.GetBlocksResponse, error) {
	return s.ledgerService.GetBlocks(ctx, req)
}
