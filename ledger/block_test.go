package ledger

import (
	"testing"

	"github.com/cachecashproject/go-cachecash/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	mockDataLen = 8
)

type mockTransaction struct {
	data []byte
}

var _ TransactionBody = (*mockTransaction)(nil)

func (tx *mockTransaction) Size() int {
	return len(tx.data)
}

func (tx *mockTransaction) TxType() TxType {
	return TxTypeUnknown
}

func (tx *mockTransaction) MarshalTo(data []byte) (int, error) {
	n := copy(data, tx.data)
	return n, nil
}

func (tx *mockTransaction) Unmarshal(data []byte) error {
	panic("not supported by mock: Unmarshal")
}

func (tx *mockTransaction) UnmarshalFrom(data []byte) (int, error) {
	panic("not supported by mock: UnmarshalFrom")
}

func (tx *mockTransaction) Inpoints() []Outpoint {
	panic("not supported by mock: Inpoints")
}

func (tx *mockTransaction) OutputCount() uint8 {
	panic("not supported by mock: OutputCount")
}

func (tx *mockTransaction) TxInputs() []TransactionInput {
	panic("not supported by mock: TxInputs")
}

func (tx *mockTransaction) TxOutputs() []TransactionOutput {
	panic("not supported by mock: TxOutputs")
}

func (tx *mockTransaction) TxWitnesses() []TransactionWitness {
	panic("not supported by mock: TxWitnesses")
}

type BlockTestSuite struct {
	suite.Suite

	l *logrus.Logger
}

func TestBlockTestSuite(t *testing.T) {
	suite.Run(t, new(BlockTestSuite))
}

func (suite *BlockTestSuite) makeBlock(txQty int) *Block {
	var txs []*Transaction

	for i := 0; i < txQty; i++ {
		data := make([]byte, mockDataLen)
		for j := 0; j < len(data); j++ {
			data[j] = (byte)(i)
		}

		txs = append(txs, &Transaction{
			Body: &mockTransaction{data: data},
		})
	}

	return &Block{
		Transactions: txs,
	}
}

func (suite *BlockTestSuite) SetupTest() {
	t := suite.T()

	l := logrus.New()
	suite.l = l

	_ = t
}

func (suite *BlockTestSuite) TestMarshal() {
	t := suite.T()

	block := &Block{
		Header: &BlockHeader{
			Version:       123,
			PreviousBlock: [32]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
			MerkleRoot:    []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
			Timestamp:     1234,
			Signature:     []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63},
		},
		Transactions: []*Transaction{
			{
				Version: 0x01,
				Flags:   0x0000,
				Body: &TransferTransaction{
					Inputs: []TransactionInput{
						{
							Outpoint: Outpoint{
								PreviousTx: mustDecodeTXID("deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
								Index:      0,
							},
							ScriptSig:  testutil.MustDecodeString("abc123"),
							SequenceNo: 0xFFFFFFFF,
						},
					},
					Witnesses: []TransactionWitness{
						{
							// A zero-item stack.
						},
					},
					Outputs: []TransactionOutput{
						{
							Value:        1234,
							ScriptPubKey: testutil.MustDecodeString("def456"),
						},
					},
				},
			},
		},
	}

	blockBytes := []byte{0x7b, 0x0, 0x0, 0x0, 0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f, 0xd2, 0x4, 0x0, 0x0, 0x38, 0x0, 0x0, 0x0, 0x1, 0x1, 0x0, 0x0, 0x1, 0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef, 0x0, 0x3, 0xab, 0xc1, 0x23, 0xff, 0xff, 0xff, 0xff, 0x1, 0xd2, 0x4, 0x0, 0x0, 0x3, 0xde, 0xf4, 0x56, 0x0}

	bytes, err := block.Marshal()
	assert.Nil(t, err)
	assert.Equal(t, blockBytes, bytes)

	block2 := &Block{}
	err = block2.Unmarshal(blockBytes)
	assert.Nil(t, err)
	assert.Equal(t, block, block2)
}

func (suite *BlockTestSuite) TestMerkleRoot() {
	t := suite.T()

	for _, p := range []struct {
		TxQty    int
		Expected []byte
	}{
		// {0, testutil.MustDecodeString("")}, // TODO: This should return an error.
		{1, testutil.MustDecodeString("2c910d9f228bd2bd6112c481c6a534a88833b8c9507eb4284530ed4976a39169")},
		{2, testutil.MustDecodeString("74bea43f5579b31a8d71c78988007bc6f8769c23266952f976fb013aca840bee")},
		{3, testutil.MustDecodeString("50bcb43a25f27ed0b964453a97ad1277395e379551e08945a7abd875e378806c")},
		{4, testutil.MustDecodeString("70aa5271e38881ab7cb6affa71f476e6747f183b48261540ea33bbedad985b79")},
		{5, testutil.MustDecodeString("cd39c515582186e29b62a314f33f1ce371171dc792b8287aa85d44563aaa3c44")},
		{6, testutil.MustDecodeString("659aca5e29066426ce3d1fe98236c0b9c960562183968a95b274f43a86818503")},
		{7, testutil.MustDecodeString("a22f0ddacf3af74c218b287927c7f27a12b83cbe3b98d6f1d87fb0f5d33f6503")},
		{8, testutil.MustDecodeString("26a11d3b865d6ea5df18ad9f4e3ee6f13b84ccedad2bf7a502fda83024c62d87")},
		{9, testutil.MustDecodeString("d30386515333f61a21a0f8bac0e3ad8074fa7c22728a92a9f84d7acc994ca96b")},
	} {
		blk := suite.makeBlock(p.TxQty)
		actual, err := blk.MerkleRoot()
		if !assert.Nil(t, err) {
			continue
		}
		assert.Equal(t, p.Expected, actual)
	}
}
