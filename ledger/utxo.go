package ledger

import "fmt"

type UTXOSet struct {
	utxos map[OutpointKey]struct{}
}

func NewUTXOSet() *UTXOSet {
	return &UTXOSet{utxos: make(map[OutpointKey]struct{})}
}

func (us *UTXOSet) Update(tx Transaction) error {
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
