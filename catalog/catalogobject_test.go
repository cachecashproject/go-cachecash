package catalog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockCount(t *testing.T) {
	assert.Equal(t, 2, BlockCount(20, 10))
	assert.Equal(t, 3, BlockCount(25, 10))
	assert.Equal(t, 1, BlockCount(1, 10))
	// do we care about empty objects?
	// assert.Equal(t, 0, BlockCount(0, 10))
}

func TestChunkIntoBlocks(t *testing.T) {
	policy := ObjectPolicy{
		BlockSize: 2,
	}
	assert.Equal(t, [][]byte{[]byte("ab"), []byte("cd"), []byte("ef"), []byte("gh")}, policy.ChunkIntoBlocks([]byte("abcdefgh")))
	assert.Equal(t, [][]byte{[]byte("ab")}, policy.ChunkIntoBlocks([]byte("ab")))
	assert.Equal(t, [][]byte{[]byte("ab"), []byte("c")}, policy.ChunkIntoBlocks([]byte("abc")))
}
