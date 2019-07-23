// Package txscript implements a limited subset of the Bitcoin script language.
//
// We support only standard transactions, and our definition of a standard transaction is more narrow than Bitcoin's.
// In particular, we support only native P2WPKH transactions.
//
package txscript

type VirtualMachine struct {
	// stack is the underlying data on the stack.  The 0th element of this slice is the bottom element on the stack.
	stack *ScriptStack
}

func (vm *VirtualMachine) Execute(scr *Script) error {
	return nil
}

func (vm *VirtualMachine) ExecuteWitness(version uint8) error {
	return nil
}
