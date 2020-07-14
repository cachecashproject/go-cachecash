package colocationpuzzle

// XXX: I don't like the name "offset" for "piece index", or "cipher block index".
//
// XXX: Is there any use in also having an IV for each block?

import (
	"bytes"
	"crypto/aes"
	"crypto/md5"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"hash/crc32"
	"math/rand"
	"time"
	"unsafe"

	"github.com/pkg/errors"

	"github.com/cachecashproject/go-cachecash/util"
)

const (
	IVSize  = aes.BlockSize
	KeySize = 16      // XXX: This shouldn't go here.
	dCrypto = "crc32" // XXX: This is only for testing which algorithm is faster.
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

type getBlockFnT func(chunkIdx, offset uint32) ([]byte, error)

func init() {
	// XXX: Is this the right place to put this?
	// TODO: check if we need secure randomness here
	rand.Seed(time.Now().UTC().UnixNano())
}

// params, raw-chunks, keys -> secret, goal, offset
// (offset is not shared with the client)
func Generate(params Parameters, chunks [][]byte, innerKeys [][]byte, innerIVs [][]byte) (*Puzzle, error) {
	if err := params.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid parameters")
	}
	if len(chunks) == 0 {
		return nil, errors.New("must have at least one chunk")
	}
	if len(chunks) != len(innerKeys) || len(chunks) != len(innerIVs) {
		return nil, errors.New("must have same numer of chunks, keys, and IVs")
	}

	if params.Rounds*uint32(len(chunks)) <= 1 {
		// XXX: Using a single ruond and a single cache is a silly idea, but `runPuzzle` will fail with those inputs.
		// With e.g. two rounds over a single cache, it won't fail, but will return an all-zero secret.
		return nil, errors.New("must use at least two puzzle iterations; increase number of rounds or caches")
	}

	// XXX: It doesn't look like we're really using chunkSize!  Can we remove it?  If we can't, we should probably
	//   rename it--"chunkCount", maybe?
	// Compute, in advance, how many cipher blocks each chunk contains.
	chunkSize := make([]int, len(chunks))
	for i := 0; i < len(chunks); i++ {
		chunkLen := len(chunks[i])
		// XXX: This is almost certainly a problem: what do we do when we get objects that are not a multiple of the
		// cipher block size?
		chunkSize[i] = chunkLen / aes.BlockSize
	}

	// Now let's pick a place to start.  We randomly select this value and don't tell the client what we choose; the
	// client having to perform a brute-force search for the correct value is what makes the puzzle "hard" in the sense
	// that we want.
	startOffset := uint32(rand.Intn(chunkSize[0]))

	getBlockFn := func(i, offset uint32) ([]byte, error) {
		chunkLen := len(chunks[i])
		offset = offset % uint32(chunkLen/aes.BlockSize)
		plaintext, err := getCipherBlock(chunks[i], offset)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get piece")
		}
		return util.EncryptCipherBlock(plaintext, innerKeys[i], innerIVs[i], offset)
	}

	goal, secret, err := runPuzzle(params.Rounds, uint32(len(chunks)), startOffset, getBlockFn, dCrypto)
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

// params, enc-chunks, goal -> secret, offset
// (offset is not necessary, but having it is useful for testing/debugging)
func Solve(params Parameters, chunks [][]byte, goal []byte) ([]byte, uint32, error) {
	debugResult := make([]byte, len(goal))

	if err := params.Validate(); err != nil {
		return nil, 0, errors.Wrap(err, "invalid parameters")
	}
	if len(chunks) == 0 {
		return nil, 0, errors.New("must have at least one chunk")
	}
	if len(goal) != sha512.Size384 {
		return nil, 0, errors.New("goal value must be a SHA-384 digest; its length is wrong")
	}
	if params.Rounds*uint32(len(chunks)) <= 1 {
		// XXX: Using a single round and a single cache is a silly idea, but `runPuzzle` will fail with those inputs.
		return nil, 0, errors.New("must use at least two puzzle iterations; increase number of rounds or caches")
	}

	getBlockFn := func(chunkIdx, offset uint32) ([]byte, error) {
		offset = offset % uint32(len(chunks[chunkIdx])/aes.BlockSize)
		return chunks[chunkIdx][offset*aes.BlockSize : (offset+1)*aes.BlockSize], nil
	}

	// Try all possible starting offsets and look for one that produces the result/goal value we're looking for.
	for offset := uint32(0); offset < uint32(len(chunks[0])/aes.BlockSize); offset++ {
		result, secret, err := runPuzzle(params.Rounds, uint32(len(chunks)), offset, getBlockFn, dCrypto)
		if err != nil {
			return nil, 0, errors.Wrap(err, "failed to run puzzle")
		}
		if bytes.Equal(goal, result) {
			return secret, offset, nil
		}
		debugResult = result
	}

	return nil, 0, errors.New("\n" + hex.Dump(goal) + "\n" + hex.Dump(debugResult))

}

func VerifySolution(params Parameters, chunks [][]byte, goal []byte, offset uint32) ([]byte, uint32, error) {
	if err := params.Validate(); err != nil {
		return nil, 0, errors.Wrap(err, "invalid parameters")
	}
	if len(chunks) == 0 {
		return nil, 0, errors.New("must have at least one chunk")
	}
	if len(goal) != sha512.Size384 {
		return nil, 0, errors.New("goal value must be a SHA-384 digest; its length is wrong")
	}
	if params.Rounds*uint32(len(chunks)) <= 1 {
		// XXX: Using a single ruond and a single cache is a silly idea, but `runPuzzle` will fail with those inputs.
		return nil, 0, errors.New("must use at least two puzzle iterations; increase number of rounds or caches")
	}

	getBlockFn := func(chunkIdx, offset uint32) ([]byte, error) {
		offset = offset % uint32(len(chunks[chunkIdx])/aes.BlockSize)
		return chunks[chunkIdx][offset*aes.BlockSize : (offset+1)*aes.BlockSize], nil
	}

	result, secret, err := runPuzzle(params.Rounds, uint32(len(chunks)), offset, getBlockFn, dCrypto)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to run puzzle")
	}
	if secret == nil {
		return nil, 0, errors.New("Secret Unsolved")
	}
	if bytes.Equal(goal, result) {
		if secret == nil {
			return nil, 0, errors.New("invalid result")
		}
		return secret, offset, nil
	}
	// XXX: some code needs to be cleaned up after all the stuff work
	return nil, 0, errors.New("no solution found, goal and result are: " + "\n" + hex.Dump(goal) + "\n" + hex.Dump(result))

}

func runPuzzle(rounds, chunkQty, offset uint32, getBlockFn getBlockFnT, method string) ([]byte, []byte, error) {
	// XXX: Do we actually need to compute this 'curLoc'?  I think that we can just check that rounds and len(chunks)
	//   are large enough that we'll never return it as prevLoc.
	// XXX: Is 'location' the best name for curLoc/prevLoc?
	var curLoc, prevLoc []byte // = hash(startblock, startoffset)

	// Initializing this to all zeroes allows this code to function with two rounds and a single cache.  Obviously, in
	// that situation the puzzle does not do anything anyhow, so having a predictable secret does not hurt us.

	curLoc = make([]byte, sha512.Size384) //Using length of 32 no matter what the method is
	if method == "sha384" {
		for i := uint32(0); i < uint32((rounds*chunkQty)-1); i++ {
			chunkIdx := i % chunkQty
			piece, err := getBlockFn(chunkIdx, offset)
			if err != nil {
				return nil, nil, errors.Wrap(err, "failed to get piece")
			}
			// piece := p.chunks[chunkIdx].data[offset*aes.BlockSize : (offset+1)*aes.BlockSize]
			// assert len(piece) == aes.BlockSize

			prevLoc = curLoc
			digest := sha512.Sum384(append(curLoc, piece...))
			curLoc = digest[:]
			offset = binary.LittleEndian.Uint32(curLoc[len(curLoc)-4:])
		}
		return curLoc, prevLoc, nil
	}

	if method == "md5" {
		for i := uint32(0); i < uint32((rounds*chunkQty)-1); i++ {
			var md5Digest []byte

			chunkIdx := i % chunkQty
			piece, err := getBlockFn(chunkIdx, offset)
			if err != nil {
				return nil, nil, errors.Wrap(err, "failed to get piece")
			}

			prevLoc = curLoc
			hashSum := md5.Sum(append(curLoc, piece...))
			hashSumS := hashSum[:]
			for j := uint32(0); j < uint32(3); j++ {
				md5Digest = append(md5Digest, hashSumS...)
			}
			curLoc = md5Digest
			offset = binary.LittleEndian.Uint32(curLoc[len(curLoc)-4:])
			//DEBUG START
			//println("DEBUG DATA"+ "\n" + hex.EncodeToString(curLoc) + "\n" + hex.EncodeToString(prevLoc) + "\n")
			if i > 0 && prevLoc == nil {
				return nil, nil, errors.New("Something goes wrong")
			}
			if curLoc == nil {
				return nil, nil, errors.New("Something goes wrong")
			}
			//DEBUG END
		}
		return curLoc, prevLoc, nil
	}

	if method == "crc32" {
		for i := uint32(0); i < uint32((rounds*chunkQty)-1); i++ {
			var crcDigest []byte
			chunkIdx := i % chunkQty
			piece, err := getBlockFn(chunkIdx, offset)
			if err != nil {
				return nil, nil, errors.Wrap(err, "failed to get piece")
			}
			// piece := p.chunks[chunkIdx].data[offset*aes.BlockSize : (offset+1)*aes.BlockSize]
			// assert len(piece) == aes.BlockSize

			prevLoc = curLoc
			hashSum := crc32.ChecksumIEEE(append(curLoc, piece...))
			hashSumS := (*[4]byte)(unsafe.Pointer(&hashSum))[:] //crc32 CheckSum returns uint32 type
			for j := uint32(0); j < uint32(12); j++ {           //crc32's checksum has a size of 4
				crcDigest = append(crcDigest, hashSumS...)
			}
			curLoc = crcDigest
			offset = binary.LittleEndian.Uint32(curLoc[len(curLoc)-4:])
		}
		return curLoc, prevLoc, nil
	}

	return nil, nil, errors.New("Method not defined")
}

func getCipherBlock(chunk []byte, cipherBlockIdx uint32) ([]byte, error) {
	cipherBlockIdx = cipherBlockIdx % uint32(len(chunk)/aes.BlockSize)
	cipherBlock := chunk[cipherBlockIdx*aes.BlockSize : (cipherBlockIdx+1)*aes.BlockSize]
	return cipherBlock, nil
}
