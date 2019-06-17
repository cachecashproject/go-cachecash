package catalog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockCount(t *testing.T) {
	assert.Equal(t, 2, ChunkCount(20, 10))
	assert.Equal(t, 3, ChunkCount(25, 10))
	assert.Equal(t, 1, ChunkCount(1, 10))
	// do we care about empty objects?
	// assert.Equal(t, 0, ChunkCount(0, 10))
}

func TestSplitIntoChunks(t *testing.T) {
	policy := ObjectPolicy{
		ChunkSize: 2,
	}
	assert.Equal(t, [][]byte{[]byte("ab"), []byte("cd"), []byte("ef"), []byte("gh")}, policy.SplitIntoChunks([]byte("abcdefgh")))
	assert.Equal(t, [][]byte{[]byte("ab")}, policy.SplitIntoChunks([]byte("ab")))
	assert.Equal(t, [][]byte{[]byte("ab"), []byte("c")}, policy.SplitIntoChunks([]byte("abc")))
}
