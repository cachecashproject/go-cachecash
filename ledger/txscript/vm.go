// Package txscript implements a limited subset of the Bitcoin script language.
//
// We support only standard transactions, and our definition of a standard transaction is more narrow than Bitcoin's.
// In particular, we support only native P2WPKH transactions.
//
package txscript

type VirtualMachine struct {
	stack *ScriptStack
}

// ScriptStack represents a Bitcoin script stack.  This type offers utility functions for interpreting values that will
// be pushed or popped as various types.
type ScriptStack struct {
	data [][]byte
}

// TODO: Need to push/pop/peek []byte, various types of ints, etc.
