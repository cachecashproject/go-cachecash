package faucet

import (
	"context"
	"time"

	"github.com/cachecashproject/go-cachecash/wallet"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ed25519"
)

type Faucet struct {
	l            *logrus.Logger
	wallet       *wallet.Wallet
	syncInterval time.Duration
}

func NewFaucet(l *logrus.Logger, wallet *wallet.Wallet) (*Faucet, error) {
	return &Faucet{
		l:            l,
		wallet:       wallet,
		syncInterval: 10 * time.Second,
	}, nil
}

func (f *Faucet) SyncChain() {
	ctx := context.Background()
	for {
		err := f.FetchBlocks(ctx)
		if err != nil {
			f.l.Error(err)
		}
		time.Sleep(f.syncInterval)
	}
}

func (f *Faucet) FetchBlocks(ctx context.Context) error {
	return f.wallet.FetchBlocks(ctx)
}

func (f *Faucet) SendCoins(ctx context.Context, target ed25519.PublicKey, amount uint32) error {
	return f.wallet.SendCoins(ctx, target, amount)
}
