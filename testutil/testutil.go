package testutil

import (
	"encoding/hex"
	"math/rand"

	"github.com/pkg/errors"
)

func RandBytes(n int) []byte {
	x := make([]byte, n)
	if _, err := rand.Read(x); err != nil {
		panic(errors.Wrap(err, "failed to generate random digest"))
	}
	return x
}

func RandBytesFromSource(src rand.Source, n int) []byte {
	x := make([]byte, n)
	if _, err := rand.New(src).Read(x); err != nil {
		panic(errors.Wrap(err, "failed to generate random digest"))
	}
	return x
}

func MustDecodeString(s string) []byte {
	d, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return d
}
