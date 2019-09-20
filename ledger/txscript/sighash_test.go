package txscript

import (
	"testing"

	"github.com/cachecashproject/go-cachecash/keypair"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ed25519"
)

func TestP2WPHVerifyValid(t *testing.T) {
	kp, err := keypair.Generate()
	assert.Nil(t, err)
	pubKeyHash := Hash160Sum(kp.PublicKey)

	inScr, err := MakeP2WPKHOutputScript(pubKeyHash)
	assert.Nil(t, err)

	outScr, err := MakeP2WPKHInputScript(pubKeyHash)
	assert.Nil(t, err)

	txIdx := 0
	amount := int64(1234)

	tx := &DummySigHash{
		sighash: []byte{0, 1, 2, 3, 4},
	}

	sighash, err := tx.SigHash(inScr, txIdx, amount)
	assert.Nil(t, err)
	signature := ed25519.Sign(kp.PrivateKey, sighash)

	witness := [][]byte{
		signature,
		kp.PublicKey,
	}

	err = ExecuteVerify(inScr, outScr, witness, tx, txIdx, amount)
	assert.Nil(t, err)
}

func TestP2WPHVerifyWrongWitnessSig(t *testing.T) {
	kp1, err := keypair.Generate()
	assert.Nil(t, err)
	kp2, err := keypair.Generate()
	assert.Nil(t, err)

	pubKeyHash := Hash160Sum(kp1.PublicKey)

	inScr, err := MakeP2WPKHOutputScript(pubKeyHash)
	assert.Nil(t, err)

	outScr, err := MakeP2WPKHInputScript(pubKeyHash)
	assert.Nil(t, err)

	txIdx := 0
	amount := int64(1234)

	tx := &DummySigHash{
		sighash: []byte{0, 1, 2, 3, 4},
	}

	sighash, err := tx.SigHash(inScr, txIdx, amount)
	assert.Nil(t, err)
	signature := ed25519.Sign(kp2.PrivateKey, sighash) // sign with wrong privkey

	witness := [][]byte{
		signature,
		kp1.PublicKey,
	}

	err = ExecuteVerify(inScr, outScr, witness, tx, txIdx, amount)
	assert.NotNil(t, err)
}

func TestP2WPHVerifyWrongInput(t *testing.T) {
	kp1, err := keypair.Generate()
	assert.Nil(t, err)
	kp2, err := keypair.Generate()
	assert.Nil(t, err)

	pubKeyHash1 := Hash160Sum(kp1.PublicKey)
	pubKeyHash2 := Hash160Sum(kp2.PublicKey)

	inScr, err := MakeP2WPKHOutputScript(pubKeyHash2) // this is somebody elses public key
	assert.Nil(t, err)

	outScr, err := MakeP2WPKHInputScript(pubKeyHash1)
	assert.Nil(t, err)

	txIdx := 0
	amount := int64(1234)

	tx := &DummySigHash{
		sighash: []byte{0, 1, 2, 3, 4},
	}

	sighash, err := tx.SigHash(inScr, txIdx, amount)
	assert.Nil(t, err)
	signature := ed25519.Sign(kp1.PrivateKey, sighash)

	witness := [][]byte{
		signature,
		kp1.PublicKey,
	}

	err = ExecuteVerify(inScr, outScr, witness, tx, txIdx, amount)
	assert.NotNil(t, err)
}

func TestP2WPHVerifyWrongWitnessKey(t *testing.T) {
	kp1, err := keypair.Generate()
	assert.Nil(t, err)
	kp2, err := keypair.Generate()
	assert.Nil(t, err)

	pubKeyHash1 := Hash160Sum(kp1.PublicKey)
	pubKeyHash2 := Hash160Sum(kp2.PublicKey)

	inScr, err := MakeP2WPKHOutputScript(pubKeyHash1)
	assert.Nil(t, err)

	outScr, err := MakeP2WPKHInputScript(pubKeyHash2)
	assert.Nil(t, err)

	txIdx := 0
	amount := int64(1234)

	tx := &DummySigHash{
		sighash: []byte{0, 1, 2, 3, 4},
	}

	sighash, err := tx.SigHash(inScr, txIdx, amount)
	assert.Nil(t, err)
	signature := ed25519.Sign(kp1.PrivateKey, sighash)

	witness := [][]byte{
		signature,
		kp2.PublicKey, // this key doesn't match the input script
	}

	err = ExecuteVerify(inScr, outScr, witness, tx, txIdx, amount)
	assert.NotNil(t, err)
}

func TestMakeOutputScript(t *testing.T) {
	pubkey := ed25519.PublicKey([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31})
	script, err := MakeOutputScript(pubkey)
	assert.Nil(t, err)

	expected := []byte{0x0, 0x14, 0x2f, 0x28, 0x7b, 0x4d, 0x3d, 0x49, 0x10, 0xf6, 0xca, 0xda, 0x9e, 0x1b, 0xd1, 0xb4, 0x64, 0x80, 0x99, 0xe8, 0xc5, 0x2c}
	assert.Equal(t, expected, script)
}

func TestMakeInputScript(t *testing.T) {
	pubkey := ed25519.PublicKey([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31})
	script, err := MakeInputScript(pubkey)
	assert.Nil(t, err)

	expected := []byte{0x76, 0xa9, 0x14, 0x2f, 0x28, 0x7b, 0x4d, 0x3d, 0x49, 0x10, 0xf6, 0xca, 0xda, 0x9e, 0x1b, 0xd1, 0xb4, 0x64, 0x80, 0x99, 0xe8, 0xc5, 0x2c, 0x88, 0xac}
	assert.Equal(t, expected, script)
}
