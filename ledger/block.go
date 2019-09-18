package ledger

import (
	"crypto/sha256"
	"encoding/binary"
	"math"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ed25519"
)

const (
	merkleRootByteLength = 32
	blockIDByteLength    = 32
)

// <---- BlockHeader ----

// BlockHeader of a block containing metadata and a signature
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

func (header BlockHeader) Marshal() ([]byte, error) {
	s := header.Size()
	data := make([]byte, s)
	n, err := header.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	if n != len(data) {
		return nil, errors.New("unexpected data length in BlockHeader.Marshal()")
	}
	return data, nil
}

func (header *BlockHeader) MarshalTo(data []byte) (int, error) {
	var n int

	binary.LittleEndian.PutUint32(data[n:], header.Version)
	n += 4

	n += copy(data[n:], header.PreviousBlock[:])
	a := copy(data[n:], header.MerkleRoot)
	if a != merkleRootByteLength {
		// XXX: MerkleRoot shouldn't be dynamic length
		return 0, errors.New("MerkleRoot didn't write 32 bytes")
	}
	n += a
	a = copy(data[n:], header.Signature)
	if a != ed25519.SignatureSize {
		// XXX: Signature shouldn't be dynamic length
		return 0, errors.New("Signature didn't write 64 bytes")
	}
	n += a

	binary.LittleEndian.PutUint32(data[n:], header.Timestamp)
	n += 4
	return n, nil
}

func (header *BlockHeader) Unmarshal(data []byte) error {
	_, err := header.UnmarshalFrom(data)
	return err
}

func (header *BlockHeader) UnmarshalFrom(data []byte) (int, error) {
	var n int

	header.MerkleRoot = make([]byte, merkleRootByteLength)
	header.Signature = make([]byte, ed25519.SignatureSize)

	if len(data[n:]) < 4 {
		return 0, errors.New("incomplete Version field")
	}
	header.Version = binary.LittleEndian.Uint32(data[n:])
	n += 4

	if len(data[n:]) < blockIDByteLength {
		return 0, errors.New("incomplete PreviousBlock field")
	}
	n += copy(header.PreviousBlock[:], data[n:n+blockIDByteLength])

	if len(data[n:]) < merkleRootByteLength {
		return 0, errors.New("incomplete MerkleRoot field")
	}
	n += copy(header.MerkleRoot, data[n:n+merkleRootByteLength])

	if len(data[n:]) < ed25519.SignatureSize {
		return 0, errors.New("incomplete Signature field")
	}
	n += copy(header.Signature, data[n:n+ed25519.SignatureSize])

	if len(data[n:]) < 4 {
		return 0, errors.New("incomplete Timestamp field")
	}
	header.Timestamp = binary.LittleEndian.Uint32(data[n:])
	n += 4
	return n, nil
}

func (header *BlockHeader) Size() int {
	var n int
	n += 4                     // version uint32
	n += BlockIDSize           // len(block.Header.PreviousBlock)
	n += merkleRootByteLength  // len(block.Header.MerkleRoot)
	n += ed25519.SignatureSize // len(block.Header.Signature)
	n += 4
	return n
}

// ---- BlockHeader ---->

// NewBlock creates a new block containing the given transactions and sign it
func NewBlock(sigKey ed25519.PrivateKey, previousBlock BlockID, txs []*Transaction) (*Block, error) {
	b := &Block{
		Header: &BlockHeader{
			Version:       0,
			PreviousBlock: previousBlock,
			Timestamp:     0, // XXX: Populate this correctly.
		},
		Transactions: &Transactions{Transactions: txs},
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

// BlockID of the block header
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

// CanonicalDigest of the block header
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

// MerkleRoot  returns a hash of all transactions
func (block *Block) MerkleRoot() ([]byte, error) {
	txs := block.Transactions

	if len(txs.Transactions) == 0 {
		return nil, errors.New("transaction list is empty")
	}

	dd := make([][]byte, len(txs.Transactions))
	for i := 0; i < len(txs.Transactions); i++ {
		d, err := txs.Transactions[i].TXID()
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
