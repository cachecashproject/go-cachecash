// Package txscript implements a limited subset of the Bitcoin script language.
//
// We support only standard transactions, and our definition of a standard transaction is more narrow than Bitcoin's.
// In particular, we support only native P2WPKH transactions.
//
package txscript

import (
	"github.com/pkg/errors"
)

type VirtualMachine struct {
	// stack is the underlying data on the stack.  The 0th element of this slice is the bottom element on the stack.
	stack       *ScriptStack
	tx          SigHashable
	script      *Script
	txIdx       int
	inputAmount int64
}

func NewVirtualMachine() *VirtualMachine {
	return &VirtualMachine{
		stack: &ScriptStack{},
	}
}

func (vm *VirtualMachine) Execute(scr *Script) error {
	for _, ins := range scr.ast {
		if err := ins.opcode.fn(vm, ins); err != nil {
			return err
		}
	}
	return nil
}

func (vm *VirtualMachine) PushWitnessData(data [][]byte) {
	for _, d := range data {
		vm.stack.PushBytes(d)
	}
}

func (vm *VirtualMachine) Verify() error {
	v, err := vm.stack.PopBool()
	if err != nil {
		return err
	}
	if !v {
		return errors.New("OP_VERIFY failed; top stack element is not truthy")
	}
	return nil
}

func ExecuteVerify(inScr, outScr *Script, witData [][]byte, tx SigHashable, txIdx int, inputAmount int64) error {
	vm := NewVirtualMachine()
	vm.tx = tx
	vm.script = outScr
	vm.txIdx = txIdx
	vm.inputAmount = inputAmount

	if err := vm.Execute(inScr); err != nil {
		return errors.Wrap(err, "failed to execute input script (scriptPubKey)")
	}

	// XXX: This should be better-encapsulated.  These two values are consumed during the process where scriptSig is
	// generated (which the VM knows to do because the address is a P2WPKH address).
	keyHash, _ := vm.stack.PopBytes()
	_, _ = vm.stack.PopBytes()

	if len(witData) < 2 {
		return errors.New("witness data is missing public key")
	}

	witnessKeyHash := Hash160Sum(witData[1])
	if string(keyHash) != string(witnessKeyHash) {
		return errors.New("incorrect public key in witness data")
	}

	vm.PushWitnessData(witData)
	if err := vm.Execute(outScr); err != nil {
		return errors.Wrap(err, "failed to execute output script (scriptSig)")
	}

	if err := vm.Verify(); err != nil {
		return errors.Wrap(err, "failed to verify after execution")
	}
	return nil
}
