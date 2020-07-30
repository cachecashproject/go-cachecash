package client

import (
	"context"
	"encoding/base64"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"golang.org/x/crypto/ed25519"
)

type cacheMock struct {
	pubkey      string
	pubkeyBytes []byte

	backlog uint64
	chunks  [][]byte
}

var _ cacheConnection = (*cacheMock)(nil)

func newCacheMock(addr string, pubkey ed25519.PublicKey, chunks [][]byte) *cacheMock {
	return &cacheMock{
		pubkey:      base64.StdEncoding.EncodeToString(pubkey),
		pubkeyBytes: pubkey,

		chunks: chunks,
	}
}

// SubmitRequest isn't actually queueing in this test double implementation
func (cc *cacheMock) SubmitRequest(ctx context.Context, clientNotify chan DownloadResult, chunkRequest *chunkRequest) {
	if len(cc.chunks) == 0 {
		return
	}

	if len(cc.chunks) > 0 {
		chunkRequest.encData, cc.chunks = cc.chunks[0], cc.chunks[1:]
	}
	clientNotify <- DownloadResult{
		resp:  chunkRequest,
		cache: cc,
	}
}

func (cc *cacheMock) ExchangeTicketL2(context.Context, *ccmsg.ClientCacheRequest) {}

func (cc *cacheMock) Close(context.Context) error {
	return nil
}

func (cc *cacheMock) GetStatus() ccmsg.ContentRequest_ClientCacheStatus {
	return ccmsg.ContentRequest_ClientCacheStatus{
		BacklogDepth: cc.backlog,
	}
}

func (cc *cacheMock) PublicKey() string {
	return cc.pubkey
}

func (cc *cacheMock) PublicKeyBytes() []byte {
	return cc.pubkeyBytes
}
