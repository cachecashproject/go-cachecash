package testutil

import (
	"crypto/rand"

	"github.com/pkg/errors"
)

func RandBytes(n int) []byte {
	x := make([]byte, n)
	if _, err := rand.Read(x); err != nil {
		panic(errors.Wrap(err, "failed to generate random digest"))
	}
	return x
}
