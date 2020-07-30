package colocationpuzzle

/*
Microbenchmarks for the colocation puzzle implementation.

Run with e.g.
  $ go test -v -bench=. -benchtime=30s ./colocationpuzzle/...

We generate qtyBenchmarkPuzzles different random inputs, since different puzzles may be inherently easier or harder to
solve.  The execution time and number of distinct inputs should be chosen so that the results are roughly stable from
run to run.

The Go benchmark framework will output results in ns/op; we usually convert this to puzzles/sec (Hz).
*/

import (
	"testing"
)

const (
	qtyBenchmarkPuzzles = 128
)

func makeBenchSuites() []*ColocationPuzzleTestSuite {
	suites := make([]*ColocationPuzzleTestSuite, 0, qtyBenchmarkPuzzles)
	for i := 0; i < cap(suites); i++ {
		s := &ColocationPuzzleTestSuite{
			params: Parameters{
				Rounds:      2,
				StartOffset: 0,
				StartRange:  0,
			},
		}
		s.SetupSuite()
		suites = append(suites, s)
	}

	return suites
}

func BenchmarkGenerate(b *testing.B) {
	suites := makeBenchSuites()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		suite := suites[i%len(suites)]
		puzzle, err := Generate(suite.params, suite.plaintextBlocks, suite.innerKeys, suite.innerIVs)

		_ = puzzle
		_ = err
	}
}

func BenchmarkSolve(b *testing.B) {
	suites := makeBenchSuites()
	puzzles := make([]*Puzzle, 0, len(suites))
	for i := 0; i < cap(puzzles); i++ {
		suite := suites[i]
		p, err := Generate(suite.params, suite.plaintextBlocks, suite.innerKeys, suite.innerIVs)
		if err != nil {
			b.Fatal("failed to generate puzzle")
			return
		}
		puzzles = append(puzzles, p)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		suite := suites[i%len(suites)]
		puzzle := puzzles[i%len(suites)]

		secret, offset, err := Solve(suite.params, suite.ciphertextBlocks, puzzle.Goal)

		_ = secret
		_ = offset
		_ = err
	}
}
