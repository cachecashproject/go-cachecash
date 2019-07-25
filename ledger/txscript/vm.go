// Package txscript implements a limited subset of the Bitcoin script language.
//
// We support only standard transactions, and our definition of a standard transaction is more narrow than Bitcoin's.
// In particular, we support only native P2WPKH transactions.
//
package txscript

import "errors"

type VirtualMachine struct {
	// stack is the underlying data on the stack.  The 0th element of this slice is the bottom element on the stack.
	stack *ScriptStack
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
