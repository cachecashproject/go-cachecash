package txscript

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// An instruction consists of an opcode (which is always a single byte), possibly followed by immediate(s).  Right now
// there is never more than one immediate, and the length of the immediate is always constant for a particular opcode.
type instruction struct {
	opcode     *opcode
	immediates [][]byte
}

func (ins *instruction) Marshal() ([]byte, error) {
	d := []byte{ins.opcode.code}
	for _, imm := range ins.immediates {
		d = append(d, imm...)
	}
	return d, nil
}

func (ins *instruction) PrettyPrint() (string, error) {
	s := ins.opcode.name
	for _, imm := range ins.immediates {
		s += " 0x" + hex.EncodeToString(imm)
	}
	return s, nil
}

type Script struct {
	// We don't support even the limited kinds of conditional execution that Bitcoin does at the moment, so this
	// representation is very simple: a list of parsed instructions.
	ast []*instruction
}

func ParseScript(buf []byte) (*Script, error) {
	return nil, nil
}

func (scr *Script) Marshal() ([]byte, error) {
	var d []byte
	for _, ins := range scr.ast {
		insd, err := ins.Marshal()
		if err != nil {
			return nil, err
		}
		d = append(d, insd...)
	}
	return d, nil
}

// StandardOutput returns nil iff the transaction is standard; otherwise, it returns a descriptive error.
func (scr *Script) StandardOutput() error {
	if len(scr.ast) != 2 {
		return errors.New("unexpected script length")
	}
	if scr.ast[0].opcode.code != OP_0 {
		return errors.New("script does not begin with OP_0")
	}
	if len(scr.ast[0].immediates) != 0 {
		return errors.New("OP_0 must have 0 immediate(s)")
	}
	if scr.ast[1].opcode.code != OP_DATA_20 {
		return errors.New("script does not end with OP_DATA_20")
	}
	if len(scr.ast[1].immediates) != 1 {
		return fmt.Errorf("OP_DATA_20 must have 1 immediate(s); found %v", len(scr.ast[1].immediates))
	}
	if len(scr.ast[1].immediates[0]) != 20 {
		return errors.New("OP_DATA_20 immediate #0 must have length 20")
	}
	return nil
}

func (scr *Script) PrettyPrint() (string, error) {
	var ss []string
	for _, ins := range scr.ast {
		s, err := ins.PrettyPrint()
		if err != nil {
			return "", err
		}
		ss = append(ss, s)
	}
	return strings.Join(ss, " "), nil
}

// TODO: When we implement better transaction-building/signing helpers, this might want to go live with them.
func MakeP2WPKHOutputScript(pubKeyHash []byte) (*Script, error) {
	if len(pubKeyHash) != 20 { // XXX: Magic number!
		return nil, errors.New("bad length for pubkeyhash")
	}
	return &Script{
		ast: []*instruction{
			{opcode: OPCODES[OP_0]},
			{opcode: OPCODES[OP_DATA_20], immediates: [][]byte{pubKeyHash}},
		},
	}, nil
}

func MakeP2WPKHInputScript(pubKeyHash []byte) (*Script, error) {
	if len(pubKeyHash) != 20 { // XXX: Magic number!
		return nil, errors.New("bad length for pubkeyhash")
	}
	return &Script{
		ast: []*instruction{
			{opcode: OPCODES[OP_DUP]},
			{opcode: OPCODES[OP_HASH160]},
			{opcode: OPCODES[OP_DATA_20], immediates: [][]byte{pubKeyHash}},
			{opcode: OPCODES[OP_EQUALVERIFY]},
			{opcode: OPCODES[OP_CHECKSIG]},
		},
	}, nil
}
