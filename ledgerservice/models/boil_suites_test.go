// +build sqlboiler_test

// Code generated by SQLBoiler (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import "testing"

// This test suite runs each operation test in parallel.
// Example, if your database has 3 tables, the suite will run:
// table1, table2 and table3 Delete in parallel
// table1, table2 and table3 Insert in parallel, and so forth.
// It does NOT run each operation group in parallel.
// Separating the tests thusly grants avoidance of Postgres deadlocks.
func TestParent(t *testing.T) {
	t.Run("Blocks", testBlocks)
	t.Run("MempoolTransactions", testMempoolTransactions)
}

func TestDelete(t *testing.T) {
	t.Run("Blocks", testBlocksDelete)
	t.Run("MempoolTransactions", testMempoolTransactionsDelete)
}

func TestQueryDeleteAll(t *testing.T) {
	t.Run("Blocks", testBlocksQueryDeleteAll)
	t.Run("MempoolTransactions", testMempoolTransactionsQueryDeleteAll)
}

func TestSliceDeleteAll(t *testing.T) {
	t.Run("Blocks", testBlocksSliceDeleteAll)
	t.Run("MempoolTransactions", testMempoolTransactionsSliceDeleteAll)
}

func TestExists(t *testing.T) {
	t.Run("Blocks", testBlocksExists)
	t.Run("MempoolTransactions", testMempoolTransactionsExists)
}

func TestFind(t *testing.T) {
	t.Run("Blocks", testBlocksFind)
	t.Run("MempoolTransactions", testMempoolTransactionsFind)
}

func TestBind(t *testing.T) {
	t.Run("Blocks", testBlocksBind)
	t.Run("MempoolTransactions", testMempoolTransactionsBind)
}

func TestOne(t *testing.T) {
	t.Run("Blocks", testBlocksOne)
	t.Run("MempoolTransactions", testMempoolTransactionsOne)
}

func TestAll(t *testing.T) {
	t.Run("Blocks", testBlocksAll)
	t.Run("MempoolTransactions", testMempoolTransactionsAll)
}

func TestCount(t *testing.T) {
	t.Run("Blocks", testBlocksCount)
	t.Run("MempoolTransactions", testMempoolTransactionsCount)
}

func TestHooks(t *testing.T) {
	t.Run("Blocks", testBlocksHooks)
	t.Run("MempoolTransactions", testMempoolTransactionsHooks)
}

func TestInsert(t *testing.T) {
	t.Run("Blocks", testBlocksInsert)
	t.Run("Blocks", testBlocksInsertWhitelist)
	t.Run("MempoolTransactions", testMempoolTransactionsInsert)
	t.Run("MempoolTransactions", testMempoolTransactionsInsertWhitelist)
}

// TestToOne tests cannot be run in parallel
// or deadlocks can occur.
func TestToOne(t *testing.T) {}

// TestOneToOne tests cannot be run in parallel
// or deadlocks can occur.
func TestOneToOne(t *testing.T) {}

// TestToMany tests cannot be run in parallel
// or deadlocks can occur.
func TestToMany(t *testing.T) {}

// TestToOneSet tests cannot be run in parallel
// or deadlocks can occur.
func TestToOneSet(t *testing.T) {}

// TestToOneRemove tests cannot be run in parallel
// or deadlocks can occur.
func TestToOneRemove(t *testing.T) {}

// TestOneToOneSet tests cannot be run in parallel
// or deadlocks can occur.
func TestOneToOneSet(t *testing.T) {}

// TestOneToOneRemove tests cannot be run in parallel
// or deadlocks can occur.
func TestOneToOneRemove(t *testing.T) {}

// TestToManyAdd tests cannot be run in parallel
// or deadlocks can occur.
func TestToManyAdd(t *testing.T) {}

// TestToManySet tests cannot be run in parallel
// or deadlocks can occur.
func TestToManySet(t *testing.T) {}

// TestToManyRemove tests cannot be run in parallel
// or deadlocks can occur.
func TestToManyRemove(t *testing.T) {}

func TestReload(t *testing.T) {
	t.Run("Blocks", testBlocksReload)
	t.Run("MempoolTransactions", testMempoolTransactionsReload)
}

func TestReloadAll(t *testing.T) {
	t.Run("Blocks", testBlocksReloadAll)
	t.Run("MempoolTransactions", testMempoolTransactionsReloadAll)
}

func TestSelect(t *testing.T) {
	t.Run("Blocks", testBlocksSelect)
	t.Run("MempoolTransactions", testMempoolTransactionsSelect)
}

func TestUpdate(t *testing.T) {
	t.Run("Blocks", testBlocksUpdate)
	t.Run("MempoolTransactions", testMempoolTransactionsUpdate)
}

func TestSliceUpdateAll(t *testing.T) {
	t.Run("Blocks", testBlocksSliceUpdateAll)
	t.Run("MempoolTransactions", testMempoolTransactionsSliceUpdateAll)
}
