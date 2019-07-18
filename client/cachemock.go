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

	backlog chan DownloadTask
	chunks  [][]byte
}

var _ cacheConnection = (*cacheMock)(nil)

func newCacheMock(addr string, pubkey ed25519.PublicKey, chunks [][]byte) *cacheMock {
	return &cacheMock{
		pubkey:      base64.StdEncoding.EncodeToString(pubkey),
		pubkeyBytes: pubkey,

		backlog: make(chan DownloadTask, 128),
		chunks:  chunks,
	}
}

func (cc *cacheMock) Run(context.Context) {
	if len(cc.chunks) == 0 {
		return
	}

	for task := range cc.backlog {
		chunkRequest := task.req
		if len(cc.chunks) > 0 {
			chunkRequest.encData, cc.chunks = cc.chunks[0], cc.chunks[1:]
		}
		task.clientNotify <- DownloadResult{
			resp:  chunkRequest,
			cache: cc,
		}
	}
}

func (cc *cacheMock) QueueRequest(task DownloadTask) {
	cc.backlog <- task
}

func (cc *cacheMock) ExchangeTicketL2(context.Context, *ccmsg.ClientCacheRequest) error {
	return nil
}

func (cc *cacheMock) Close(context.Context) error {
	return nil
}

func (cc *cacheMock) GetStatus() ccmsg.ContentRequest_ClientCacheStatus {
	return ccmsg.ContentRequest_ClientCacheStatus{
		BacklogDepth: uint64(len(cc.backlog)),
	}
}

func (cc *cacheMock) PublicKey() string {
	return cc.pubkey
}

func (cc *cacheMock) PublicKeyBytes() []byte {
	return cc.pubkeyBytes
}
