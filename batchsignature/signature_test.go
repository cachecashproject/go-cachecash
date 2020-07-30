package batchsignature

import (
	"crypto/rand"
	"crypto/sha512"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DigestTreeTestSuite struct {
	suite.Suite
}

const (
	MaxTestSize = 33
)

func TestDigestTreeTestSuite(t *testing.T) {
	suite.Run(t, new(DigestTreeTestSuite))
}

func (suite *DigestTreeTestSuite) randomDigests(n uint) [][]byte {
	t := suite.T()

	dd := make([][]byte, n)
	for i := uint(0); i < n; i++ {
		dd[i] = make([]byte, sha512.Size384)
		if _, err := rand.Read(dd[i]); err != nil {
			t.Fatalf("failed to generate random digest: %v", err)
		}
	}

	return dd
}

func (suite *DigestTreeTestSuite) TestEmptyInput() {
	t := suite.T()

	var leaves [][]byte
	rootDigest, trees, err := NewDigestTree(leaves)
	assert.Nil(t, err)
	assert.Nil(t, rootDigest)
	assert.Equal(t, 0, len(trees))
}

// TODO: Once I have internet access, look up how to make this separate tests instead of a for-loop.
func (suite *DigestTreeTestSuite) TestDigestTree() {
	t := suite.T()

	for n := uint(1); n <= MaxTestSize; n++ {
		leaves := suite.randomDigests(n)

		rootDigest, trees, err := NewDigestTree(leaves)
		_, _, _ = rootDigest, trees, err

		if !assert.Nil(t, err) {
			break
		}

		ok := assert.Equal(t, int(n), len(trees))
		ok = ok && assert.Equal(t, sha512.Size384, len(rootDigest))
		if !ok {
			break
		}

		for i, tree := range trees {
			recomputedRootDigest := tree.RootDigest()
			ok = ok && assert.Equal(t, rootDigest, recomputedRootDigest, fmt.Sprintf("path %v produced incorrect root digest", i))
		}
		if !ok {
			break
		}
	}
}
