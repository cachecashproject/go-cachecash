package colocationpuzzle

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha512"
	"testing"

	cachecash "github.com/kelleyk/go-cachecash"
	"github.com/kelleyk/go-cachecash/testutil"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ColocationPuzzleTestSuite struct {
	suite.Suite

	obj              cachecash.ContentObject
	plaintextBlocks  [][]byte
	ciphertextBlocks [][]byte
	innerKeys        [][]byte
	innerIVs         [][]byte
	params           Parameters
	blockIndices     []uint32
}

const (
	BlockQty  = 4
	BlockSize = aes.BlockSize * 1024
)

func TestColocationPuzzleTestSuite(t *testing.T) {
	suite.Run(t, new(ColocationPuzzleTestSuite))
}

// XXX: TODO: Add a test case that covers the single-block case; the implementation fails!

func (suite *ColocationPuzzleTestSuite) SetupTest() {
	t := suite.T()

	suite.params = Parameters{
		Rounds:      2,
		StartOffset: 0, // XXX: Not used yet.
		StartRange:  0,
	}

	for i := 0; i < BlockQty; i++ {
		k := testutil.RandBytes(16)
		suite.innerKeys = append(suite.innerKeys, k)

		iv := testutil.RandBytes(aes.BlockSize)
		suite.innerIVs = append(suite.innerIVs, iv)

		// TODO: The blocks need not all be the same size.
		b := testutil.RandBytes(BlockSize)
		suite.plaintextBlocks = append(suite.plaintextBlocks, b)

		// Set up our AES-CTR encryption gizmo.
		block, err := aes.NewCipher(k)
		if err != nil {
			t.Fatal(errors.Wrap(err, "failed to create block cipher"))
		}
		stream := cipher.NewCTR(block, iv)

		// Encrypt the plaintext block.
		cb := make([]byte, len(b))
		stream.XORKeyStream(cb, b)
		suite.ciphertextBlocks = append(suite.ciphertextBlocks, cb)

		suite.blockIndices = append(suite.blockIndices, uint32(i))
	}

	// Wrap the plaintext blocks in a ContentObject.
	suite.obj = cachecash.NewContentBuffer(suite.plaintextBlocks)
}

func (suite *ColocationPuzzleTestSuite) TestGenerateAndSolve() {
	t := suite.T()

	puzzle, err := Generate(suite.params, suite.obj, suite.blockIndices, suite.innerKeys, suite.innerIVs)
	if !assert.Nil(t, err, "failed to generate puzzle") {
		return
	}

	assert.Equal(t, sha512.Size384, len(puzzle.Secret), "unexpected secret length")
	assert.Equal(t, sha512.Size384, len(puzzle.Goal), "unxpected goal length")

	secret, offset, err := Solve(suite.params, suite.ciphertextBlocks, puzzle.Goal)
	if !assert.Nil(t, err, "failed to solve puzzle") {
		return
	}

	// It's only actually important that the secret be the same on both sides, but it would be very odd if multiple
	// offsets led to correct solutions to the puzzle.
	assert.Equal(t, puzzle.Offset, offset, "generator and solevr do not agree on starting offset")
	assert.Equal(t, puzzle.Secret, secret, "generator and solver do not agree on secret")
}

func (suite *ColocationPuzzleTestSuite) TestKeyAndIVFromSecret() {
	t := suite.T()

	puzzle, err := Generate(suite.params, suite.obj, suite.blockIndices, suite.innerKeys, suite.innerIVs)
	if !assert.Nil(t, err, "failed to generate puzzle") {
		return
	}

	key := puzzle.Key()
	iv := puzzle.IV()

	assert.Equal(t, KeySize, len(key))
	assert.Equal(t, IVSize, len(iv))
}
