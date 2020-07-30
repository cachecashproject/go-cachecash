package kv

import (
	"context"
	"encoding/binary"
	"errors"
	"math"
)

var (
	// ErrUnmarshal is returned when unmarshaling to a native type failed.
	ErrUnmarshal = errors.New("unmarshaling failed")
	// ErrMarshal is returned when marshaling from a native type failed.
	ErrMarshal = errors.New("marshaling failed")
)

func marshalBytesInt64(out int64) []byte {
	byt := make([]byte, 8)
	binary.LittleEndian.PutUint64(byt, uint64(out))
	return byt
}

func marshalBytesUint64(out uint64) []byte {
	byt := make([]byte, binary.MaxVarintLen64)
	binary.LittleEndian.PutUint64(byt, out)
	return byt
}

// CreateUint64 creates a new key as uint64.
func (c *Client) CreateUint64(ctx context.Context, key string, out uint64) ([]byte, error) {
	return c.driver.Create(ctx, c.member, key, marshalBytesUint64(out))
}

// GetUint64 retrieves the marshaled data for key and then converts it to uint64.
func (c *Client) GetUint64(ctx context.Context, key string) (uint64, []byte, error) {
	out, nonce, err := c.driver.Get(ctx, c.member, key)
	if err != nil {
		return 0, nonce, err
	}

	u := binary.LittleEndian.Uint64(out)
	return u, nonce, nil
}

// SetUint64 sets a uint64 to key.
func (c *Client) SetUint64(ctx context.Context, key string, out uint64) ([]byte, error) {
	return c.driver.Set(ctx, c.member, key, marshalBytesUint64(out))
}

// CASUint64 compares and swaps uint64 values.
func (c *Client) CASUint64(ctx context.Context, key string, nonce []byte, origValue, value uint64) ([]byte, error) {
	return c.driver.CAS(ctx, c.member, key, nonce, marshalBytesUint64(origValue), marshalBytesUint64(value))
}

// CreateInt64 creates a new key as int64.
func (c *Client) CreateInt64(ctx context.Context, key string, out int64) ([]byte, error) {
	return c.driver.Create(ctx, c.member, key, marshalBytesInt64(out))
}

// GetInt64 retrieves the marshaled data for key and then converts it to int64.
func (c *Client) GetInt64(ctx context.Context, key string) (int64, []byte, error) {
	out, nonce, err := c.driver.Get(ctx, c.member, key)
	if err != nil {
		return 0, nonce, err
	}

	v := binary.LittleEndian.Uint64(out)
	return int64(v), nonce, nil
}

// SetInt64 sets a int64 to key.
func (c *Client) SetInt64(ctx context.Context, key string, out int64) ([]byte, error) {
	return c.driver.Set(ctx, c.member, key, marshalBytesInt64(out))
}

// CASInt64 compares and swaps uint64 values.
func (c *Client) CASInt64(ctx context.Context, key string, nonce []byte, origValue, value int64) ([]byte, error) {
	return c.driver.CAS(ctx, c.member, key, nonce, marshalBytesInt64(origValue), marshalBytesInt64(value))
}

// CreateString creates a string where there wasn't one before.
func (c *Client) CreateString(ctx context.Context, key, value string) ([]byte, error) {
	return c.driver.Create(ctx, c.member, key, []byte(value))
}

// GetString retrieves the marshaled data for key and then converts it to string.
func (c *Client) GetString(ctx context.Context, key string) (string, []byte, error) {
	out, nonce, err := c.driver.Get(ctx, c.member, key)
	if err != nil {
		return "", nonce, err
	}

	return string(out), nonce, nil
}

// SetString sets a string to a key.
func (c *Client) SetString(ctx context.Context, key, value string) ([]byte, error) {
	return c.driver.Set(ctx, c.member, key, []byte(value))
}

// CASString compares and swaps strings.
func (c *Client) CASString(ctx context.Context, key string, nonce []byte, origValue, value string) ([]byte, error) {
	return c.driver.CAS(ctx, c.member, key, nonce, []byte(origValue), []byte(value))
}

// CreateFloat64 creates a float64 where there wasn't one before.
func (c *Client) CreateFloat64(ctx context.Context, key string, out float64) ([]byte, error) {
	return c.CreateUint64(ctx, key, math.Float64bits(out))
}

// GetFloat64 retrieves the marshaled data for key and then converts it to float64.
func (c *Client) GetFloat64(ctx context.Context, key string) (float64, []byte, error) {
	u, nonce, err := c.GetUint64(ctx, key)
	if err != nil {
		return 0, nonce, err
	}

	return math.Float64frombits(u), nonce, nil
}

// SetFloat64 sets a float64 to a key.
func (c *Client) SetFloat64(ctx context.Context, key string, value float64) ([]byte, error) {
	return c.SetUint64(ctx, key, math.Float64bits(value))
}

// CASFloat64 compares and swaps float64s.
func (c *Client) CASFloat64(ctx context.Context, key string, nonce []byte, origValue, value float64) ([]byte, error) {
	return c.CASUint64(ctx, key, nonce, math.Float64bits(origValue), math.Float64bits(value))
}
