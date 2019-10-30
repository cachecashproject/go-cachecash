package ledger

// TODO: Test that copy() works well enough that calling Apply() does not modify the original state.

// Test
// - delete first/last/middle items in list
// - interaction of deletes & insertions
// - mid-list insertions

// Test error-case behavior
// - Insert (index>0) into list that doesn't exist
// - Insert past end of list
// - Delete of element that doesn't exist

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type GlobalConfigTestSuite struct {
	suite.Suite

	l *logrus.Logger
}

func TestGlobalConfigTestSuite(t *testing.T) {
	suite.Run(t, new(GlobalConfigTestSuite))
}

func (suite *GlobalConfigTestSuite) SetupTest() {
	t := suite.T()

	l := logrus.New()
	suite.l = l

	_ = t
}

func (suite *GlobalConfigTestSuite) TestScalarInsert() {
	t := suite.T()

	st := NewGlobalConfigState()
	st2, err := st.Apply(&GlobalConfigTransaction{
		ScalarUpdates: []GlobalConfigScalarUpdate{
			{Key: "abc", Value: []byte("def")},
		},
		ListUpdates: []GlobalConfigListUpdate{},
	})

	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, []byte("def"), st2.Scalars["abc"])
}

func (suite *GlobalConfigTestSuite) TestScalarUpsert() {
	t := suite.T()

	st := NewGlobalConfigState()
	st.Scalars["abc"] = []byte("def")

	st2, err := st.Apply(&GlobalConfigTransaction{
		ScalarUpdates: []GlobalConfigScalarUpdate{
			{Key: "abc", Value: []byte("ghi")},
		},
		ListUpdates: []GlobalConfigListUpdate{},
	})

	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, []byte("ghi"), st2.Scalars["abc"])
}

func (suite *GlobalConfigTestSuite) TestScalarDelete() {
	t := suite.T()

	st := NewGlobalConfigState()
	st.Scalars["abc"] = []byte("def")

	st2, err := st.Apply(&GlobalConfigTransaction{
		ScalarUpdates: []GlobalConfigScalarUpdate{
			{Key: "abc", Value: []byte{}},
		},
		ListUpdates: []GlobalConfigListUpdate{},
	})

	if !assert.Nil(t, err) {
		return
	}

	_, ok := st2.Scalars["abc"]
	assert.False(t, ok)
}

func (suite *GlobalConfigTestSuite) TestListInsertNew() {
	t := suite.T()

	st := NewGlobalConfigState()
	st2, err := st.Apply(&GlobalConfigTransaction{
		ScalarUpdates: []GlobalConfigScalarUpdate{},
		ListUpdates: []GlobalConfigListUpdate{
			{Key: "FuzzyWombats", Insertions: []GlobalConfigListInsertion{
				{Index: 0, Value: []byte("alpha")},
			}},
		},
	})

	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, [][]byte{
		[]byte("alpha"),
	}, st2.Lists["FuzzyWombats"])
}

func (suite *GlobalConfigTestSuite) makeListTestState() *GlobalConfigState {
	st := NewGlobalConfigState()
	st.Lists["FuzzyWombats"] = [][]byte{
		[]byte("red"),
		[]byte("blue"),
		[]byte("green"),
	}
	return st
}

func (suite *GlobalConfigTestSuite) TestListInsertPrepends() {
	t := suite.T()

	st := suite.makeListTestState()
	st2, err := st.Apply(&GlobalConfigTransaction{
		ScalarUpdates: []GlobalConfigScalarUpdate{},
		ListUpdates: []GlobalConfigListUpdate{
			{Key: "FuzzyWombats", Insertions: []GlobalConfigListInsertion{
				{Index: 0, Value: []byte("dog")},
				{Index: 0, Value: []byte("cat")},
			}},
		},
	})

	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, [][]byte{
		[]byte("cat"),
		[]byte("dog"),
		[]byte("red"),
		[]byte("blue"),
		[]byte("green"),
	}, st2.Lists["FuzzyWombats"])
}

func (suite *GlobalConfigTestSuite) TestListInsertAppends() {
	t := suite.T()

	st := suite.makeListTestState()
	st2, err := st.Apply(&GlobalConfigTransaction{
		ScalarUpdates: []GlobalConfigScalarUpdate{},
		ListUpdates: []GlobalConfigListUpdate{
			{Key: "FuzzyWombats", Insertions: []GlobalConfigListInsertion{
				{Index: 3, Value: []byte("dog")},
				{Index: 4, Value: []byte("cat")},
			}},
		},
	})

	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, [][]byte{
		[]byte("red"),
		[]byte("blue"),
		[]byte("green"),
		[]byte("dog"),
		[]byte("cat"),
	}, st2.Lists["FuzzyWombats"])
}

func (suite *GlobalConfigTestSuite) TestNoOperations() {
	t := suite.T()

	st := suite.makeListTestState()

	st2, err := st.Apply(&GlobalConfigTransaction{})
	assert.Nil(t, err)
	assert.Equal(t, st, st2)

	st2, err = st.Apply(&GlobalConfigTransaction{ScalarUpdates: nil, ListUpdates: nil})
	assert.Nil(t, err)
	assert.Equal(t, st, st2)

	st2, err = st.Apply(&GlobalConfigTransaction{ScalarUpdates: []GlobalConfigScalarUpdate{}, ListUpdates: []GlobalConfigListUpdate{}})
	assert.Nil(t, err)
	assert.Equal(t, st, st2)
}

func (suite *GlobalConfigTestSuite) TestListNoop() {
	t := suite.T()

	st := suite.makeListTestState()
	_, err := st.Apply(&GlobalConfigTransaction{
		ListUpdates: []GlobalConfigListUpdate{
			{Key: "FuzzyWombats"}, // No insertions or deletions; this is not valid.
		},
	})

	assert.NotNil(t, err)
}

func (suite *GlobalConfigTestSuite) TestListDeletionErrors() {
	t := suite.T()

	for i, tt := range []struct {
		lu GlobalConfigListUpdate
	}{
		// Deletions past the end of the list (when the list is empty).
		{lu: GlobalConfigListUpdate{Key: "nonexistent", Deletions: []uint64{1}}},
		// Invalid sequence of deletion indices: out of order.
		{lu: GlobalConfigListUpdate{Key: "FuzzyWombats", Deletions: []uint64{2, 1, 0}}},
		// Invalid sequence of deletion indices: duplicates.
		{lu: GlobalConfigListUpdate{Key: "FuzzyWombats", Deletions: []uint64{2, 2, 2}}},
	} {
		st := suite.makeListTestState()
		st2, err := st.Apply(&GlobalConfigTransaction{ListUpdates: []GlobalConfigListUpdate{tt.lu}})

		assert.NotNil(t, err, fmt.Sprintf("[case %v] no error, but expected one", i))
		assert.Nil(t, st2, fmt.Sprintf("[case %v] got return value, but expected nil", i))
	}
}

func (suite *GlobalConfigTestSuite) TestListDeletion() {
	t := suite.T()

	for i, tt := range []struct {
		lu       GlobalConfigListUpdate
		expected [][]byte
	}{
		{
			lu:       GlobalConfigListUpdate{Key: "FuzzyWombats", Deletions: []uint64{0, 1}},
			expected: [][]byte{[]byte("green")},
		},
		{
			lu:       GlobalConfigListUpdate{Key: "FuzzyWombats", Deletions: []uint64{1, 2}},
			expected: [][]byte{[]byte("red")},
		},
		{
			lu:       GlobalConfigListUpdate{Key: "FuzzyWombats", Deletions: []uint64{0, 2}},
			expected: [][]byte{[]byte("blue")},
		},
		{
			lu:       GlobalConfigListUpdate{Key: "FuzzyWombats", Deletions: []uint64{1}},
			expected: [][]byte{[]byte("red"), []byte("green")},
		},
	} {
		st := suite.makeListTestState()
		st2, err := st.Apply(&GlobalConfigTransaction{ListUpdates: []GlobalConfigListUpdate{tt.lu}})

		assert.Nil(t, err, fmt.Sprintf("[case %v] got error, but expected nil", i))
		assert.Equal(t, tt.expected, st2.Lists["FuzzyWombats"],
			fmt.Sprintf("[case %v] unexpected value for list after applying update", i))
	}
}
