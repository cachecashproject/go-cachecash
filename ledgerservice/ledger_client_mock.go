package ledgerservice

import (
	"context"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type LedgerClientMock struct {
	mock.Mock
}

var _ ccmsg.LedgerClient = (*LedgerClientMock)(nil)

func NewLedgerClientMock() *LedgerClientMock {
	return &LedgerClientMock{}
}

func (m *LedgerClientMock) PostTransaction(ctx context.Context, in *ccmsg.PostTransactionRequest, opts ...grpc.CallOption) (*ccmsg.PostTransactionResponse, error) {
	args := m.Called(ctx, in, opts)
	cr := args.Get(0).(*ccmsg.PostTransactionResponse)
	err := args.Error(1)
	return cr, err
}

func (m *LedgerClientMock) GetBlocks(ctx context.Context, in *ccmsg.GetBlocksRequest, opts ...grpc.CallOption) (*ccmsg.GetBlocksResponse, error) {
	args := m.Called(ctx, in, opts)
	cr := args.Get(0).(*ccmsg.GetBlocksResponse)
	err := args.Error(1)
	return cr, err
}
