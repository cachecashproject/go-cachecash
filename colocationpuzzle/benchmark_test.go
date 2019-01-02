package colocationpuzzle

import (
	"testing"
)

func makeBenchSuite() *ColocationPuzzleTestSuite {
	suite := &ColocationPuzzleTestSuite{
		params: Parameters{
			Rounds:      2,
			StartOffset: 0,
			StartRange:  0,
		},
	}
	suite.SetupTest()
	return suite
}

func BenchmarkGenerate(b *testing.B) {
	suite := makeBenchSuite()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		puzzle, err := Generate(suite.params, suite.obj, suite.blockIndices, suite.innerKeys, suite.innerIVs)

		_ = puzzle
		_ = err
	}
}

func BenchmarkSolve(b *testing.B) {
	suite := makeBenchSuite()
	puzzle, err := Generate(suite.params, suite.obj, suite.blockIndices, suite.innerKeys, suite.innerIVs)
	if err != nil {
		b.Fatal("failed to generate puzzle")
		return
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		secret, offset, err := Solve(suite.params, suite.ciphertextBlocks, puzzle.Goal)

		_ = secret
		_ = offset
		_ = err
	}
}
