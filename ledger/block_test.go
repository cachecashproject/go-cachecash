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
	var txs []Transaction

	for i := 0; i < txQty; i++ {
		data := make([]byte, mockDataLen)
		for j := 0; j < len(data); j++ {
			data[j] = (byte)(i)
		}

		txs = append(txs, Transaction{
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
