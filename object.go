package cachecash

import (
	"crypto/aes"
	"crypto/rand"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

const (
	defaultBlockSize = 256 * 1024
)

type ContentObject interface {
	// Get metadata (number and sizes of blocks?)

	// BlockSize returns the size of a particular data block in bytes.
	// TODO: Do we really need this?
	BlockSize(dataBlockIdx uint32) (int, error)

	GetBlock(dataBlockIdx uint32) ([]byte, error)

	// GetCipherBlock returns an individual cipher block (aka "sub-block") of a particular data block (a protocol-level
	// block).  The return value will be aes.BlockSize bytes long (16 bytes).  ciperBlockIdx is taken modulo the number
	// of whole cipher blocks that exist in the data block.
	GetCipherBlock(dataBlockIdx, cipherBlockIdx uint32) ([]byte, error)

	// BlockCount returns the number of bloks in this object.
	BlockCount() int
}

// type contentFile struct {
// }

// var _ ContentObject = (*contentFile)(nil)

// contentBuffer is a ContentObject backed by in-memory data (i.e. a "buffer").
type contentBuffer struct {
	blocks [][]byte
}

var _ ContentObject = (*contentBuffer)(nil)

func NewContentBuffer(blocks [][]byte) ContentObject {
	return &contentBuffer{blocks: blocks}
}

// RandomContentBuffer generates and returns a contentBuffer containing blockQty blocks; each block will contain
// blockSize bytes of random data.
func RandomContentBuffer(blockQty, blockSize uint32) (ContentObject, error) {
	blocks := make([][]byte, blockQty)
	for i := uint32(0); i < blockQty; i++ {
		blocks[i] = make([]byte, blockSize)
		if _, err := rand.Read(blocks[i]); err != nil {
			return nil, errors.Wrap(err, "failed to generate random block data")
		}
	}
	return NewContentBuffer(blocks), nil
}

// TODO: This would probably look nicer if we didn't use ioutil.ReadAll.
func NewContentBufferFromFile(path string) (ContentObject, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}
	defer func() { _ = f.Close() }()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read file")
	}

	obj := &contentBuffer{}

	// Split the single data buffer up into fixed-size blocks.
	for offset := 0; offset < len(data); offset += defaultBlockSize {
		blockSize := defaultBlockSize
		if offset+blockSize > len(data) {
			blockSize = len(data) - offset
		}

		obj.blocks = append(obj.blocks, data[offset:offset+blockSize])
	}

	return obj, nil
}

func (o *contentBuffer) GetBlock(dataBlockIdx uint32) ([]byte, error) {
	if int(dataBlockIdx) >= len(o.blocks) {
		return nil, errors.New("data block index out of range")
	}

	return o.blocks[dataBlockIdx], nil
}

func (o *contentBuffer) GetCipherBlock(dataBlockIdx, cipherBlockIdx uint32) ([]byte, error) {
	if int(dataBlockIdx) >= len(o.blocks) {
		return nil, errors.New("data block index out of range")
	}
	dataBlock := o.blocks[dataBlockIdx]

	cipherBlockIdx = cipherBlockIdx % uint32(len(dataBlock)/aes.BlockSize)
	cipherBlock := dataBlock[cipherBlockIdx*aes.BlockSize : (cipherBlockIdx+1)*aes.BlockSize]
	return cipherBlock, nil
}

func (o *contentBuffer) BlockSize(dataBlockIdx uint32) (int, error) {
	return len(o.blocks[dataBlockIdx]), nil
}

func (o *contentBuffer) BlockCount() int {
	return len(o.blocks)
}
