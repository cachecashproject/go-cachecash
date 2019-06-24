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
}

var _ cacheConnection = (*cacheMock)(nil)

func newCacheMock(addr string, pubkey ed25519.PublicKey) *cacheMock {
	return &cacheMock{
		pubkey:      base64.StdEncoding.EncodeToString(pubkey),
		pubkeyBytes: pubkey,

		backlog: make(chan DownloadTask, 128),
	}
}

func (cc *cacheMock) Run(context.Context) {
	// empty
}

func (cc *cacheMock) QueueRequest(task DownloadTask) {
	cc.backlog <- task
}

func (cc *cacheMock) ExchangeTicketL2(context.Context, *ccmsg.ClientCacheRequest) error {
	panic("unimplemented(ExchangeTicketL2)")
}

func (cc *cacheMock) Close(context.Context) error {
	panic("unimplemented(Close)")
}

func (cc *cacheMock) BacklogLength() uint64 {
	return uint64(len(cc.backlog))
}

func (cc *cacheMock) PublicKey() string {
	return cc.pubkey
}

func (cc *cacheMock) PublicKeyBytes() []byte {
	return cc.pubkeyBytes
}
