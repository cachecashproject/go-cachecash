package txscript

type opcode struct {
	code   byte
	length int // The number of bytes taken by the instruction, including the opcode itself and any immediate.
	name   string
	fn     func(*VirtualMachine, *instruction) error
}

const (
	OP_0           uint8 = 0x00
	OP_DATA_20           = 0x14
	OP_DUP               = 0x76
	OP_HASH160           = 0xa9 // XXX: Do we want to change this to something SHA256-related?
	OP_EQUALVERIFY       = 0x88
	OP_CHECKSIG          = 0xac
	// OP_CHECKSIGVERIFY            = 0xad
	// OP_CHECKMULTISIG             = 0xae
	// OP_CHECKMULTISIGVERIFY       = 0xaf
)

var OPCODES = map[uint8]*opcode{
	OP_0:           {OP_0, 1, "OP_0", opConst},
	OP_DATA_20:     {OP_DATA_20, 21, "OP_DATA_20", opPushData},
	OP_DUP:         {OP_DUP, 1, "OP_DUP", opDup},
	OP_HASH160:     {OP_HASH160, 1, "OP_HASH160", opHash160},
	OP_EQUALVERIFY: {OP_EQUALVERIFY, 1, "OP_EQUALVERIFY", opEqual},
	OP_CHECKSIG:    {OP_CHECKSIG, 1, "OP_CHECKSIG", opCheckSig},
	// OP_CHECKSIGVERIFY:      {OP_CHECKSIGVERIFY, 1, "OP_CHECKSIGVERIFY", opCheckSig},
	// OP_CHECKMULTISIG:       {OP_CHECKMULTISIG, 1, "OP_CHECKMULTISIG", opCheckSig},
	// OP_CHECKMULTISIGVERIFY: {OP_CHECKMULTISIGVERIFY, 1, "OP_CHECKMULTISIGVERIFY", opCheckSig},
}

func opConst(vm *VirtualMachine, inst *instruction) error {
	return nil
}

func opPushData(vm *VirtualMachine, inst *instruction) error {
	return nil
}

func opDup(vm *VirtualMachine, inst *instruction) error {
	return nil
}

func opHash160(vm *VirtualMachine, inst *instruction) error {
	return nil
}

func opEqual(vm *VirtualMachine, inst *instruction) error {
	return nil
}

func opCheckSig(vm *VirtualMachine, inst *instruction) error {
	return nil
}
