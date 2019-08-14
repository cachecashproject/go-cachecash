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
	PreviousBlock []byte // CanonicalDigest of previous block (32 bytes)
	MerkleRoot    []byte // 32 bytes
	Timestamp     uint32
	// Bits          uint32
	// Nonce         uint32

	// Signature is a signature over the canonical digest of the block header.  It is produced by the centralized ledger
	// authority.
	Signature []byte
}

type Block struct {
	Header       *BlockHeader
	Transactions []Transaction
}

func NewBlock(sigKey ed25519.PrivateKey, previousBlock []byte, txs []Transaction) (*Block, error) {
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

	cd := b.CanonicalDigest()
	b.Header.Signature = ed25519.Sign(sigKey, cd)

	return b, nil
}

func (block *Block) CanonicalDigest() []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, block.Header.Version)
	buf = append(buf, block.Header.PreviousBlock...)
	buf = append(buf, block.Header.MerkleRoot...)
	a := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, block.Header.Timestamp)
	buf = append(buf, a...)

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
		dd[i] = d
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
