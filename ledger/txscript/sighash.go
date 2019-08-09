package txscript

type SigHashable interface {
	SigHash(script *Script, txIdx int, inputAmount int64) ([]byte, error)
}
