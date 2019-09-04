package faucet

import (
	"context"
	"encoding/base64"
	"strings"
	"time"

	"github.com/cachecashproject/go-cachecash/kv"
	"github.com/cachecashproject/go-cachecash/kv/ratelimit"
	"github.com/cachecashproject/go-cachecash/wallet"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ed25519"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type Faucet struct {
	l            *logrus.Logger
	wallet       *wallet.Wallet
	syncInterval time.Duration
	rateLimiter  *ratelimit.RateLimiter
}

func NewFaucet(l *logrus.Logger, wallet *wallet.Wallet) (*Faucet, error) {
	kv := kv.NewClient("faucet", kv.NewMemoryDriver(nil))
	rateLimiter := ratelimit.NewRateLimiter(ratelimit.Config{
		Cap:             2000,
		RefreshInterval: 5 * time.Minute,
		RefreshAmount:   10,
	}, kv)

	return &Faucet{
		l:            l,
		wallet:       wallet,
		syncInterval: 10 * time.Second,
		rateLimiter:  rateLimiter,
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

func (f *Faucet) rateLimit(ctx context.Context, target ed25519.PublicKey, amount uint32) error {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return status.Errorf(codes.FailedPrecondition, "failed to get grpc peer from ctx")
	}

	// ratelimit by ip
	incomingIP := peer.Addr.String()[:strings.LastIndex(peer.Addr.String(), ":")]
	dataRateKey := strings.Join([]string{"rate-limit", "ipaddr", incomingIP}, "/")

	err := f.rateLimiter.RateLimit(dataRateKey, int64(amount))
	if err != nil {
		return errors.Wrap(err, "ip ratelimit failed")
	}

	// ratelimit by addr
	targetAddr := base64.StdEncoding.EncodeToString(target)
	dataRateKey = strings.Join([]string{"rate-limit", "pubkey", targetAddr}, "/")

	err = f.rateLimiter.RateLimit(dataRateKey, int64(amount))
	if err != nil {
		return errors.Wrap(err, "pubkey ratelimit failed")
	}

	return nil
}

func (f *Faucet) SendCoins(ctx context.Context, target ed25519.PublicKey, amount uint32) error {
	l := f.l.WithFields(logrus.Fields{
		"amount": amount,
	})
	l.Info("got coin request")
	if err := f.rateLimit(ctx, target, amount); err != nil {
		l.Info("rejected coins request: rate limit exceeded: ", err)
		return err
	}
	l.Info("sending coins from wallet")
	return f.wallet.SendCoins(ctx, target, amount)
}
