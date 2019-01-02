package colocationpuzzle

// XXX: I don't like the name "offset" for "sub-block index", or "cipher block index".
//
// XXX: Is there any use in also having an IV for each block?

import (
	"bytes"
	"crypto/aes"
	"crypto/sha512"
	"encoding/binary"
	"math/rand"
	"time"

	"github.com/pkg/errors"

	cachecash "github.com/kelleyk/go-cachecash"
	"github.com/kelleyk/go-cachecash/util"
)

const (
	IVSize  = aes.BlockSize
	KeySize = 16 // XXX: This shouldn't go here.
)

type Parameters struct {
	Rounds      uint32
	StartOffset uint32 // XXX: Not respected yet.
	StartRange  uint32
}

func (params *Parameters) Validate() error {
	if params.Rounds < 1 {
		return errors.New("puzzle must have at least one round")
	}
	return nil
}

type Puzzle struct {
	Secret []byte
	Goal   []byte // Goal is the only value that's shared with the client.
	Offset uint32
	Params Parameters
}

func (p *Puzzle) IV() []byte {
	return p.Secret[:IVSize]
}

func (p *Puzzle) Key() []byte {
	return p.Secret[IVSize : IVSize+KeySize]
}

type getBlockFnT func(blockIdx, offset uint32) ([]byte, error)

func init() {
	// XXX: Is this the right place to put this?
	rand.Seed(time.Now().UTC().UnixNano())
}

// params, raw-blocks, keys -> secret, goal, offset
// (offset is not shared with the client)
func Generate(params Parameters, obj cachecash.ContentObject, blocks []uint32, innerKeys [][]byte, innerIVs [][]byte) (*Puzzle, error) {
	if err := params.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid parameters")
	}
	if len(blocks) == 0 {
		return nil, errors.New("must have at least one data block")
	}
	if params.Rounds*uint32(len(blocks)) <= 1 {
		// XXX: Using a single ruond and a single cache is a silly idea, but `runPuzzle` will fail with those inputs.
		return nil, errors.New("must use at least two puzzle iterations; increase number of rounds or caches")
	}

	// XXX: It doesn't look like we're really using blockSize!  Can we remove it?  If we can't, we should probably
	//   rename it--"blockCount", maybe?
	// Compute, in advance, how many cipher blocks each data block contains.
	blockSize := make([]int, len(blocks))
	for i := 0; i < len(blocks); i++ {
		blockLen, err := obj.BlockSize(blocks[i])
		if err != nil {
			return nil, errors.Wrap(err, "failed to get block size")
		}
		if blockLen%aes.BlockSize != 0 {
			// XXX: Is this actually a problem?
			return nil, errors.New("input block size is not a multiple of cipher block size")
		}
		blockSize[i] = blockLen / aes.BlockSize
	}

	// Now let's pick a place to start.  We randomly select this value and don't tell the client what we choose; the
	// client having to perform a brute-force search for the correct value is what makes the puzzle "hard" in the sense
	// that we want.
	startOffset := uint32(rand.Intn(blockSize[0]))

	getBlockFn := func(blockIdx, offset uint32) ([]byte, error) {
		blockLen, err := obj.BlockSize(blockIdx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get block size")
		}
		offset = offset % uint32(blockLen/aes.BlockSize)
		plaintext, err := obj.GetCipherBlock(blockIdx, offset)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get sub-block")
		}
		return util.EncryptBlock(plaintext, innerKeys[blockIdx], innerIVs[blockIdx], offset)
	}

	goal, secret, err := runPuzzle(params.Rounds, uint32(len(blocks)), startOffset, getBlockFn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate outputs of colocation puzzle")
	}

	return &Puzzle{
		Params: params,
		Secret: secret,
		Goal:   goal,
		Offset: startOffset,
	}, nil
}

// params, enc-blocks, goal -> secret, offset
// (offset is not necessary, but having it is useful for testing/debugging)
func Solve(params Parameters, blocks [][]byte, goal []byte) ([]byte, uint32, error) {
	if err := params.Validate(); err != nil {
		return nil, 0, errors.Wrap(err, "invalid parameters")
	}
	if len(blocks) == 0 {
		return nil, 0, errors.New("must have at least one data block")
	}
	if len(goal) != sha512.Size384 {
		return nil, 0, errors.New("goal value must be a SHA-384 digest; its length is wrong")
	}
	if params.Rounds*uint32(len(blocks)) <= 1 {
		// XXX: Using a single ruond and a single cache is a silly idea, but `runPuzzle` will fail with those inputs.
		return nil, 0, errors.New("must use at least two puzzle iterations; increase number of rounds or caches")
	}

	getBlockFn := func(blockIdx, offset uint32) ([]byte, error) {
		offset = offset % uint32(len(blocks[blockIdx])/aes.BlockSize)
		return blocks[blockIdx][offset*aes.BlockSize : (offset+1)*aes.BlockSize], nil
	}

	// Try all possible starting offsets and look for one that produces the result/goal value we're looking for.
	for offset := uint32(0); offset < uint32(len(blocks[0])/aes.BlockSize); offset++ {
		result, secret, err := runPuzzle(params.Rounds, uint32(len(blocks)), offset, getBlockFn)
		if err != nil {
			return nil, 0, errors.Wrap(err, "failed to run puzzle")
		}
		if bytes.Equal(goal, result) {
			return secret, offset, nil
		}
	}

	return nil, 0, errors.New("no solution found")

}

func VerifySolution(params Parameters, blocks [][]byte, goal []byte, offset uint32) ([]byte, uint32, error) {
	if err := params.Validate(); err != nil {
		return nil, 0, errors.Wrap(err, "invalid parameters")
	}
	if len(blocks) == 0 {
		return nil, 0, errors.New("must have at least one data block")
	}
	if len(goal) != sha512.Size384 {
		return nil, 0, errors.New("goal value must be a SHA-384 digest; its length is wrong")
	}
	if params.Rounds*uint32(len(blocks)) <= 1 {
		// XXX: Using a single ruond and a single cache is a silly idea, but `runPuzzle` will fail with those inputs.
		return nil, 0, errors.New("must use at least two puzzle iterations; increase number of rounds or caches")
	}

	getBlockFn := func(blockIdx, offset uint32) ([]byte, error) {
		offset = offset % uint32(len(blocks[blockIdx])/aes.BlockSize)
		return blocks[blockIdx][offset*aes.BlockSize : (offset+1)*aes.BlockSize], nil
	}

	result, secret, err := runPuzzle(params.Rounds, uint32(len(blocks)), offset, getBlockFn)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to run puzzle")
	}
	if bytes.Equal(goal, result) {
		return secret, offset, nil
	}

	return nil, 0, errors.New("no solution found")

}

func runPuzzle(rounds, blockQty, offset uint32, getBlockFn getBlockFnT) ([]byte, []byte, error) {
	// XXX: Do we actually need to compute this 'curLoc'?  I think that we can just check that rounds and len(blocks)
	//   are large enough that we'll never return it as prevLoc.
	// XXX: Is 'location' the best name for curLoc/prevLoc?
	var curLoc, prevLoc []byte // = hash(startblock, startoffset)

	for i := uint32(0); i < uint32((rounds*blockQty)-1); i++ {
		blockIdx := i % blockQty
		subblock, err := getBlockFn(blockIdx, offset)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to get data sub-block")
		}
		// subblock := p.blocks[blockIdx].data[offset*aes.BlockSize : (offset+1)*aes.BlockSize]
		// assert len(subblock) == aes.BlockSize

		prevLoc = curLoc
		digest := sha512.Sum384(append(curLoc, subblock...))
		curLoc = digest[:]
		offset = binary.LittleEndian.Uint32(curLoc[len(curLoc)-4:])
	}

	return curLoc, prevLoc, nil
}
