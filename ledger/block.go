package ledger

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/pkg/errors"
)

type BlockHeader struct {
	Version       uint32
	PreviousBlock []byte
	MerkleRoot    []byte
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
		next := make([][]byte, len(txs)<<1)
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
