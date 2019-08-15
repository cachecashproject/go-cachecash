package ledger

import (
	"crypto/sha256"
	"encoding/binary"
	"math"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ed25519"
)

type BlockHeader struct {
	Version       uint32
	PreviousBlock BlockID // CanonicalDigest of previous block (32 bytes)
	MerkleRoot    []byte  // 32 bytes
	Timestamp     uint32
	// Bits          uint32
	// Nonce         uint32

	// Signature is a signature over the canonical digest of the block header.  It is produced by the centralized ledger
	// authority.
	Signature []byte
}

type Block struct {
	Header       *BlockHeader
	Transactions []*Transaction
}

func NewBlock(sigKey ed25519.PrivateKey, previousBlock BlockID, txs []*Transaction) (*Block, error) {
	b := &Block{
		Header: &BlockHeader{
			Version:       0,
			PreviousBlock: previousBlock,
			Timestamp:     0, // XXX: Populate this correctly.
		},
		Transactions: txs,
	}

	var err error
	b.Header.MerkleRoot, err = b.MerkleRoot()
	if err != nil {
		return nil, errors.Wrap(err, "failed to compute merkle root")
	}

	bid := b.BlockID()
	b.Header.Signature = ed25519.Sign(sigKey, bid[:])

	return b, nil
}

func (block *Block) Marshal() ([]byte, error) {
	s := block.Size()
	data := make([]byte, s)
	n, err := block.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	if n != len(data) {
		return nil, errors.New("unexpected data length in Block.Marshal()")
	}
	return data, nil
}

func (block *Block) Size() int {
	var n int

	n += 4
	n += len(block.Header.PreviousBlock)
	n += len(block.Header.MerkleRoot)
	n += len(block.Header.Signature)
	n += 4

	for _, tx := range block.Transactions {
		n += 4 + tx.Size()
	}

	return n
}

func (block *Block) MarshalTo(data []byte) (int, error) {
	var n int

	binary.LittleEndian.PutUint32(data[n:], block.Header.Version)
	n += 4

	n += copy(data[n:], block.Header.PreviousBlock[:])
	a := copy(data[n:], block.Header.MerkleRoot)
	if a != 32 {
		// XXX: MerkleRoot shouldn't be dynamic length
		return 0, errors.New("MerkleRoot didn't write 32 bytes")
	}
	n += a
	a = copy(data[n:], block.Header.Signature)
	if a != ed25519.SignatureSize {
		// XXX: Signature shouldn't be dynamic length
		return 0, errors.New("Signature didn't write 64 bytes")
	}
	n += a

	binary.LittleEndian.PutUint32(data[n:], block.Header.Timestamp)
	n += 4

	for _, tx := range block.Transactions {
		txBytes, err := tx.Marshal()
		if err != nil {
			return 0, err
		}
		binary.LittleEndian.PutUint32(data[n:], uint32(len(txBytes)))
		n += 4
		n += copy(data[n:], txBytes)
	}

	return n, nil
}

func (block *Block) Unmarshal(data []byte) error {
	_, err := block.UnmarshalFrom(data)
	return err
}

// N.B.: This is not strictly required for the protobuf interface, but it's useful for test code to be able to tell how
// many bytes were consumed.
func (block *Block) UnmarshalFrom(data []byte) (int, error) {
	var n int

	block.Header = &BlockHeader{
		MerkleRoot: make([]byte, 32),
		Signature:  make([]byte, ed25519.SignatureSize),
	}

	if len(data[n:]) < 4 {
		return 0, errors.New("incomplete Version field")
	}
	block.Header.Version = binary.LittleEndian.Uint32(data[n:])
	n += 4

	if len(data[n:]) < 32 {
		return 0, errors.New("incomplete PreviousBlock field")
	}
	n += copy(block.Header.PreviousBlock[:], data[n:n+32])

	if len(data[n:]) < 32 {
		return 0, errors.New("incomplete MerkleRoot field")
	}
	n += copy(block.Header.MerkleRoot, data[n:n+32])

	if len(data[n:]) < ed25519.SignatureSize {
		return 0, errors.New("incomplete Signature field")
	}
	n += copy(block.Header.Signature, data[n:n+ed25519.SignatureSize])

	if len(data[n:]) < 4 {
		return 0, errors.New("incomplete Timestamp field")
	}
	block.Header.Timestamp = binary.LittleEndian.Uint32(data[n:])
	n += 4

	for len(data[n:]) > 0 {
		if len(data[n:]) < 4 {
			return 0, errors.New("incomplete tx length field")
		}
		b := int(binary.LittleEndian.Uint32(data[n:]))
		n += 4

		if len(data) < b {
			return 0, errors.New("transaction length field exceeds remaining data")
		}

		tx := Transaction{}
		err := tx.Unmarshal(data[n : n+b])
		if err != nil {
			return 0, errors.Wrap(err, "failed to unmarshal transaction")
		}
		n += b

		block.Transactions = append(block.Transactions, &tx)
	}

	return n, nil
}

func (block *Block) BlockID() BlockID {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, block.Header.Version)
	buf = append(buf, block.Header.PreviousBlock[:]...)
	buf = append(buf, block.Header.MerkleRoot...)
	a := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, block.Header.Timestamp)
	buf = append(buf, a...)

	d := sha256.Sum256(buf)
	d = sha256.Sum256(d[:])
	return BlockID(d)
}

func (block *Block) CanonicalDigest() []byte {
	var n int
	buf := make([]byte, 4+len(block.Header.PreviousBlock)+len(block.Header.MerkleRoot)+4)

	binary.LittleEndian.PutUint32(buf[n:], block.Header.Version)
	n += 4
	n += copy(buf[n:], block.Header.PreviousBlock[:])
	n += copy(buf[n:], block.Header.MerkleRoot)
	binary.LittleEndian.PutUint32(buf[n:], block.Header.Timestamp)

	d := sha256.Sum256(buf)
	d = sha256.Sum256(d[:])
	return d[:]
}

func (block *Block) MerkleRoot() ([]byte, error) {
	txs := block.Transactions

	if len(txs) == 0 {
		return nil, errors.New("transaction list is empty")
	}

	dd := make([][]byte, len(txs))
	for i := 0; i < len(txs); i++ {
		d, err := txs[i].TXID()
		if err != nil {
			return nil, errors.Wrap(err, "failed to compute TXID")
		}
		dd[i] = d[:]
	}

	for len(dd) > 1 {
		next := make([][]byte, halfCeil(len(dd)))
		for i := 0; i < len(next); i++ {
			bi := i*2 + 1
			if bi >= len(dd) {
				bi = i * 2
			}
			hi := append(dd[i*2], dd[bi]...)

			d := sha256.Sum256(hi)
			d = sha256.Sum256(d[:])
			next[i] = d[:]
		}
		dd = next
	}

	return dd[0], nil
}

func halfCeil(x int) int {
	return int(math.Ceil((float64)(x) / 2.0))
}
