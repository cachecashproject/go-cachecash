package ratelimit

import (
	"context"
	"encoding/binary"
	"time"

	"github.com/cachecashproject/go-cachecash/kv"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const maxCASRetries uint = 10

// Config is the configuration passed to the rate limiter.
type Config struct {
	// Logger for debug and info messages related to rate limiting.
	Logger *logrus.Logger
	// Maximum burst rate
	Cap int64
	// Refresh burst tokens over this interval
	RefreshInterval time.Duration
	// Refresh this amount / interval
	RefreshAmount int64
}

// RateLimiter is a simple API to rate limiting with the k/v library.
type RateLimiter struct {
	Config
	kvClient *kv.Client
}

// NewRateLimiter constructs a new RateLimiter
func NewRateLimiter(c Config, kvClient *kv.Client) *RateLimiter {
	return &RateLimiter{Config: c, kvClient: kvClient}
}

func encodeCount(count int64) ([]byte, error) {
	timeBuf := make([]byte, binary.MaxVarintLen64)
	countBuf := make([]byte, binary.MaxVarintLen64)
	if binary.PutVarint(timeBuf, time.Now().Unix()) == 0 {
		return nil, errors.New("short write in buffer encode")
	}

	if binary.PutVarint(countBuf, count) == 0 {
		return nil, errors.New("short write in buffer encode")
	}

	return append(timeBuf, countBuf...), nil
}

func decodeCount(buf []byte) (time.Time, int64, error) {
	timeBuf := buf[:binary.MaxVarintLen64]
	countBuf := buf[binary.MaxVarintLen64:]

	val, n := binary.Varint(timeBuf)
	if n == 0 {
		return time.Time{}, 0, errors.New("unmarshaling error decoding time")
	}

	theTime := time.Unix(val, 0)

	count, n := binary.Varint(countBuf)
	if n == 0 {
		return time.Time{}, 0, errors.New("unmarshaling error decoding count")
	}

	return theTime, count, nil
}

// CreateInitialLimit creates an initial limit based off the Cap value.
func (rl *RateLimiter) CreateInitialLimit(ctx context.Context, dataRateKey string) ([]byte, error) {
	countBuf, err := encodeCount(rl.Cap)
	if err != nil {
		return nil, err
	}

	nonce, err := rl.kvClient.CreateBytes(ctx, dataRateKey, countBuf)
	if err != nil && errors.Cause(err) != kv.ErrAlreadySet {
		return nil, err
	}

	return nonce, nil
}

// CASUpdate is a safe method to slam in data. Will return kv.ErrNotEqual if
// the values do not line up after a few attempts. If unset, it will set the
// initial value instead of doing the update.
func (rl *RateLimiter) CASUpdate(ctx context.Context, dataRateKey string, nonce, byt, newByt []byte) ([]byte, error) {
	var retries uint
retry:
	nonce, err := rl.kvClient.CASBytes(ctx, dataRateKey, nonce, byt, newByt)
	if err == kv.ErrNotEqual {
		if retries > maxCASRetries {
			return nil, err
		}
		retries++
		goto retry
	} else if err == kv.ErrUnsetValue {
		return rl.CreateInitialLimit(ctx, dataRateKey)
	} else if err != nil {
		return nil, err
	}
	return nonce, nil
}

// RateLimit rate limits stuff, provided a dataRateKey and a value. Returns
// ErrTooMuchData if too much data has tripped the limit. Returns other errors
// on systemic issues.
func (rl *RateLimiter) RateLimit(ctx context.Context, dataRateKey string, size int64) error {
	_, err := rl.CreateInitialLimit(ctx, dataRateKey)
	if err != nil {
		return err
	}

	byt, nonce, err := rl.kvClient.GetBytes(ctx, dataRateKey)
	if err != nil {
		return err
	}

	t, count, err := decodeCount(byt)
	if err != nil {
		return err
	}

	rl.Logger.Debugf("Initial tokens for %v: %d", dataRateKey, count)

	since := time.Since(t.Add(rl.RefreshInterval))
	if since > 0 {
		rl.Logger.Debugf("Increasing count for %v to %d", dataRateKey, count)
		count += rl.RefreshAmount * int64(time.Since(t)/rl.RefreshInterval)
		if count > rl.Cap {
			count = rl.Cap
		}

		newByt, err := encodeCount(count)
		if err != nil {
			return err
		}

		nonce, err = rl.CASUpdate(ctx, dataRateKey, nonce, byt, newByt)
		if err != nil {
			return err
		}

		byt = newByt
	}

	left := count - size
	if left < 0 {
		return ErrTooMuchData
	}

	rl.Logger.Debugf("Left over for %v: %d", dataRateKey, left)

	newByt, err := encodeCount(left)
	if err != nil {
		return err
	}

	_, err = rl.CASUpdate(ctx, dataRateKey, nonce, byt, newByt)
	return err
}
