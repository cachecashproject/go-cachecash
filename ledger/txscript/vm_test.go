package txscript

import (
	"testing"

	"golang.org/x/crypto/ed25519"

	"github.com/cachecashproject/go-cachecash/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type VMTestSuite struct {
	suite.Suite

	l *logrus.Logger
}

func TestVMTestSuite(t *testing.T) {
	suite.Run(t, new(VMTestSuite))
}

func (suite *VMTestSuite) SetupTest() {
	t := suite.T()

	l := logrus.New()
	suite.l = l

	_ = t
}

func (suite *VMTestSuite) TestP2WPKHOutput_StandardOutput() {
	t := suite.T()

	// ------------
	// Setup
	// ------------

	_, pubKey, err := ed25519.GenerateKey(nil)
	pubKeyHash := hash160Sum(pubKey)

	scriptPubKey, err := MakeP2WPKHOutputScript(pubKeyHash)
	if err != nil {
		t.Fatalf("failed to create scriptPubKey: %v", err)
	}

	scriptSig, err := MakeP2WPKHInputScript(pubKeyHash)
	if err != nil {
		t.Fatalf("failed to create scriptSig: %v", err)
	}

	witData := [][]byte{
		testutil.MustDecodeString("cafebabe"), // XXX: Once we have sighash, we'll need an actual signature here.
		pubKey,
	}

	vm := NewVirtualMachine()

	// ------------
	// Execution
	// ------------

	if err := vm.Execute(scriptPubKey); err != nil {
		t.Fatalf("failed to execute scriptPubKey: %v", err)
	}

	// XXX: This should be better-encapsulated.  These two values are consumed during the process where scriptSig is
	// generated (which the VM knows to do because the address is a P2WPKH address).
	_, _ = vm.stack.PopBytes()
	_, _ = vm.stack.PopBytes()

	vm.PushWitnessData(witData)
	if err := vm.Execute(scriptSig); err != nil {
		t.Fatalf("failed to execute scriptSig: %v", err)
	}
	if err := vm.Verify(); err != nil {
		t.Fatalf("failed to verify after execution: %v", err)
	}
	assert.Equal(t, 0, vm.stack.Size())
}
