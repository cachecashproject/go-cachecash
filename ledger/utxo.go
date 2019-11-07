package ledger

import (
	"errors"
	"fmt"
)

const (
	BlockSizeLimit = 1024 * 1024 // XXX: this is an arbitrary value
)

type UTXOSet struct {
	utxos map[OutpointKey]struct{}
}

func NewUTXOSet() *UTXOSet {
	return &UTXOSet{utxos: make(map[OutpointKey]struct{})}
}

func (us *UTXOSet) Length() int {
	return len(us.utxos)
}

func (us *UTXOSet) Update(tx *Transaction) error {
	// Mark inpoints as spent.
	for _, ip := range tx.Inpoints() {
		ipk := ip.Key()
		_, ok := us.utxos[ipk]
		if !ok {
			return fmt.Errorf("inpoint not in UTXO set: %v", ip)
		}
		delete(us.utxos, ipk)
	}

	// Add outpoints to unspent set.
	for _, op := range tx.Outpoints() {
		us.utxos[op.Key()] = struct{}{}
	}

	return nil
}

// SpendingState is a temporary state when crafting a new TransferTransaction
// Transaction in the mempool may conflict with each other, like spending the
// same UTXO twice. Transactions added to the SpendingState are guaranteed to
// be conflict free.
type SpendingState struct {
	TXs        []*Transaction
	spentUTXOs map[OutpointKey]struct{}
	newUTXOs   map[OutpointKey]TransactionOutput
	size       int
}

func NewSpendingState() *SpendingState {
	return &SpendingState{
		TXs:        []*Transaction{},
		spentUTXOs: map[OutpointKey]struct{}{},
		newUTXOs:   map[OutpointKey]TransactionOutput{},
		size:       0,
	}
}

func (s *SpendingState) IsNewUnspent(key OutpointKey) *TransactionOutput {
	utxo, ok := s.newUTXOs[key]
	if ok {
		return &utxo
	} else {
		return nil
	}
}

func (s *SpendingState) AddTx(tx *Transaction) error {
	// check if we can still fit this tx in the block
	if s.Size()+tx.Size() > BlockSizeLimit {
		return errors.New("not enough remaining space in block")
	}

	// validate inpoints aren't spent twice in this block
	for _, ip := range tx.Inpoints() {
		ipk := ip.Key()
		_, alreadySpent := s.spentUTXOs[ipk]
		if alreadySpent {
			return errors.New("input has been spent already")
		}
	}

	// add tx to backlog
	s.AcceptTransaction(tx)

	return nil
}

func (s *SpendingState) AcceptTransaction(tx *Transaction) {
	txid, err := tx.TXID()
	if err != nil {
		panic(err) // XXX: We should change TXID() so that it doesn't return an error.
	}

	// mark inpoints as spent
	for _, ip := range tx.Inpoints() {
		ipk := ip.Key()
		s.spentUTXOs[ipk] = struct{}{}
	}

	// add outputs as spendable
	for i, output := range tx.Outputs() {
		op := Outpoint{
			PreviousTx: txid,
			Index:      uint8(i),
		}
		opk := op.Key()
		s.newUTXOs[opk] = output
	}

	// add to backlog
	s.TXs = append(s.TXs, tx)
	s.size += tx.Size()
}

func (s *SpendingState) AcceptedTransactions() []*Transaction {
	return s.TXs
}

func (s *SpendingState) SpentUTXOs() []OutpointKey {
	spent := make([]OutpointKey, 0, len(s.spentUTXOs))
	for k := range s.spentUTXOs {
		spent = append(spent, k)
	}
	return spent
}

func (s *SpendingState) Size() int {
	return s.size
}
