package ledger

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cachecashproject/go-cachecash/keypair"
	"github.com/cachecashproject/go-cachecash/ledger/txscript"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ed25519"
)

type SigHashType uint32

const (
	SigHashAll SigHashType = 0x1
)

func (tx *Transaction) SigHash(script *txscript.Script, txIdx int, inputAmount int64) ([]byte, error) {
	// Generate the signature hash based on the signature hash type.
	sigHashes, err := NewTransactionHashes(tx)
	if err != nil {
		return nil, err
	}

	return calcWitnessSignatureHash(script, sigHashes, SigHashAll,
		tx, txIdx, inputAmount)
}

func (tx *Transaction) GenerateWitnesses(kp *keypair.KeyPair, prevOutputs []TransactionOutput) error {
	body, ok := tx.Body.(*TransferTransaction)
	if !ok {
		return errors.New("sighash only supports TransferTransaction")
	}

	body.Witnesses = []TransactionWitness{}

	for txIdx, input := range prevOutputs {
		script, err := txscript.ParseScript(input.ScriptPubKey)
		if err != nil {
			return errors.Wrap(err, "failed to parse script")
		}

		sighash, err := tx.SigHash(script, txIdx, int64(input.Value))
		if err != nil {
			return errors.Wrap(err, "failed to calculate sighash")
		}
		signature := ed25519.Sign(kp.PrivateKey, sighash)

		body.Witnesses = append(body.Witnesses, TransactionWitness{
			Data: [][]byte{
				signature,
				kp.PublicKey,
			},
		})
	}

	return nil
}

type TransactionHashes struct {
	HashPrevOuts chainhash.Hash
	HashSequence chainhash.Hash
	HashOutputs  chainhash.Hash
}

func NewTransactionHashes(tx *Transaction) (*TransactionHashes, error) {
	HashPrevOuts, err := calcHashPrevOuts(tx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to hash previous outputs")
	}
	HashSequence, err := calcHashSequence(tx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to hash sequence numbers")
	}
	HashOutputs, err := calcHashOutputs(tx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to hash outputs")
	}

	return &TransactionHashes{
		HashPrevOuts: *HashPrevOuts,
		HashSequence: *HashSequence,
		HashOutputs:  *HashOutputs,
	}, nil
}

// calcHashPrevOuts calculates a single hash of all the previous outputs
// (txid:index) referenced within the passed transaction. This calculated hash
// can be re-used when validating all inputs spending segwit outputs, with a
// signature hash type of SigHashAll. This allows validation to re-use previous
// hashing computation, reducing the complexity of validating SigHashAll inputs
// from  O(N^2) to O(N).
func calcHashPrevOuts(tx *Transaction) (*chainhash.Hash, error) {
	var b bytes.Buffer

	body, ok := tx.Body.(*TransferTransaction)
	if !ok {
		return nil, errors.New("transaction is not a transfer transaction")
	}

	for _, in := range body.Inputs {
		// First write out the 32-byte transaction ID one of whose
		// outputs are being referenced by this input.
		b.Write(in.PreviousTx[:])

		// Next, we'll encode the index of the referenced output as a
		// little endian integer.
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], uint32(in.Index)) // TODO: do we want to use 4 bytes while Index is only uint8?
		b.Write(buf[:])
	}

	hash := chainhash.DoubleHashH(b.Bytes())
	return &hash, nil
}

// calcHashSequence computes an aggregated hash of each of the sequence numbers
// within the inputs of the passed transaction. This single hash can be re-used
// when validating all inputs spending segwit outputs, which include signatures
// using the SigHashAll sighash type. This allows validation to re-use previous
// hashing computation, reducing the complexity of validating SigHashAll inputs
// from O(N^2) to O(N).
func calcHashSequence(tx *Transaction) (*chainhash.Hash, error) {
	var b bytes.Buffer

	body, ok := tx.Body.(*TransferTransaction)
	if !ok {
		return nil, errors.New("transaction is not a transfer transaction")
	}

	for _, in := range body.Inputs {
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], in.SequenceNo)
		b.Write(buf[:])
	}

	hash := chainhash.DoubleHashH(b.Bytes())
	return &hash, nil
}

// calcHashOutputs computes a hash digest of all outputs created by the
// transaction encoded using the wire format. This single hash can be re-used
// when validating all inputs spending witness programs, which include
// signatures using the SigHashAll sighash type. This allows computation to be
// cached, reducing the total hashing complexity from O(N^2) to O(N).
func calcHashOutputs(tx *Transaction) (*chainhash.Hash, error) {
	var b bytes.Buffer

	body, ok := tx.Body.(*TransferTransaction)
	if !ok {
		return nil, errors.New("transaction is not a transfer transaction")
	}

	for _, out := range body.Outputs {
		err := WriteTxOut(&b, 0, 0, out)
		if err != nil {
			return nil, errors.Wrap(err, "failed to add output to hash")
		}
	}

	hash := chainhash.DoubleHashH(b.Bytes())
	return &hash, nil
}

func WriteTxOut(w io.Writer, pver uint32, version int32, to TransactionOutput) error {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(to.Value))
	_, err := w.Write(buf[:])
	if err != nil {
		return err
	}

	return WriteVarBytes(w, pver, to.ScriptPubKey)
}

// WriteVarInt serializes val to w using a variable number of bytes depending
// on its value.
func WriteVarInt(w io.Writer, pver uint32, val uint64) error {
	if val < 0xfd {
		_, err := w.Write([]byte{uint8(val)})
		return err
	}

	if val <= math.MaxUint16 {
		_, err := w.Write([]byte{0xfd})
		if err != nil {
			return err
		}

		var buf [2]byte
		binary.LittleEndian.PutUint16(buf[:], uint16(val))
		_, err = w.Write(buf[:])
		return err
	}

	if val <= math.MaxUint32 {
		_, err := w.Write([]byte{0xfe})
		if err != nil {
			return err
		}

		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], uint32(val))
		_, err = w.Write(buf[:])
		return err
	}

	_, err := w.Write([]byte{0xff})
	if err != nil {
		return err
	}

	// TODO: pull this array to the top and use slices for uint16/uint32
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(val))
	_, err = w.Write(buf[:])
	return err
}

func WriteVarBytes(w io.Writer, pver uint32, bytes []byte) error {
	slen := uint64(len(bytes))
	err := WriteVarInt(w, pver, slen)
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)
	return err
}

func calcWitnessSignatureHash(script *txscript.Script, sigHashes *TransactionHashes,
	hashType SigHashType, tx *Transaction, idx int, amt int64) ([]byte, error) {

	body, ok := tx.Body.(*TransferTransaction)
	if !ok {
		return nil, errors.New("transaction is not a transfer")
	}

	// As a sanity check, ensure the passed input index for the transaction
	// is valid.
	if idx > len(body.Inputs)-1 {
		return nil, errors.Errorf("idx %d but %d txins", idx, len(body.Inputs))
	}

	// We'll utilize this buffer throughout to incrementally calculate
	// the signature hash for this transaction.
	var sigHash bytes.Buffer

	// First write out, then encode the transaction's version number.
	var bVersion [4]byte
	binary.LittleEndian.PutUint32(bVersion[:], uint32(tx.Version))
	sigHash.Write(bVersion[:])

	// If anyone can pay isn't active, then we can use the cached
	// hashPrevOuts, otherwise we just write zeroes for the prev outs.
	sigHash.Write(sigHashes.HashPrevOuts[:])

	sigHash.Write(sigHashes.HashSequence[:])

	txIn := body.Inputs[idx]

	// Next, write the outpoint being spent.
	sigHash.Write(txIn.PreviousTx[:])
	var bIndex [4]byte
	binary.LittleEndian.PutUint32(bIndex[:], txIn.SequenceNo)
	sigHash.Write(bIndex[:])

	// The script code for a p2wkh is a length prefix varint for
	// the next 25 bytes, followed by a re-creation of the original
	// p2pkh pk script.
	sigHash.Write([]byte{0x19})
	sigHash.Write([]byte{txscript.OP_DUP})
	sigHash.Write([]byte{txscript.OP_HASH160})
	sigHash.Write([]byte{txscript.OP_DATA_20})

	scriptBytes, err := script.Marshal()
	if err != nil {
		return nil, err
	}

	sigHash.Write(scriptBytes)
	sigHash.Write([]byte{txscript.OP_EQUALVERIFY})
	sigHash.Write([]byte{txscript.OP_CHECKSIG})

	// Next, add the input amount, and sequence number of the input being
	// signed.
	var bAmount [8]byte
	binary.LittleEndian.PutUint64(bAmount[:], uint64(amt))
	sigHash.Write(bAmount[:])
	var bSequence [4]byte
	binary.LittleEndian.PutUint32(bSequence[:], txIn.SequenceNo)
	sigHash.Write(bSequence[:])

	sigHash.Write(sigHashes.HashOutputs[:])

	// Finally, write out the transaction's locktime, and the sig hash
	// type.
	var bLockTime [4]byte
	binary.LittleEndian.PutUint32(bLockTime[:], body.LockTime)
	sigHash.Write(bLockTime[:])
	var bHashType [4]byte
	binary.LittleEndian.PutUint32(bHashType[:], uint32(hashType))
	sigHash.Write(bHashType[:])

	return chainhash.DoubleHashB(sigHash.Bytes()), nil
}
