package ledger

import (
	"errors"
	"fmt"
)

const (
	BlockSizeLimit = 1000 // XXX: this is an arbitrary value
)

type UTXOSet struct {
	utxos map[OutpointKey]struct{}
}

func NewUTXOSet() *UTXOSet {
	return &UTXOSet{utxos: make(map[OutpointKey]struct{})}
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
	spent map[OutpointKey]struct{}
	txs   []*Transaction
	size  int
}

func NewSpendingState() *SpendingState {
	return &SpendingState{
		spent: map[OutpointKey]struct{}{},
		txs:   []*Transaction{},
		size:  0,
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
		_, alreadySpent := s.spent[ipk]
		if alreadySpent {
			return errors.New("input has been spent already")
		}
	}

	// validation was successful, mark inpoints as spent
	for _, ip := range tx.Inpoints() {
		ipk := ip.Key()
		s.spent[ipk] = struct{}{}
	}

	// add tx to backlog
	s.txs = append(s.txs, tx)
	s.size += tx.Size()

	return nil
}

func (s *SpendingState) AcceptedTransactions() []*Transaction {
	return s.txs
}

func (s *SpendingState) SpentUTXOs() []OutpointKey {
	spent := make([]OutpointKey, 0, len(s.spent))
	for k := range s.spent {
		spent = append(spent, k)
	}
	return spent
}

func (s *SpendingState) Size() int {
	return s.size
}
