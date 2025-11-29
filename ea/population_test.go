package ea

import (
	"bufio"
	"math/rand/v2"
	"os"
	"testing"

	"github.com/danielkennedy1/sieve/genomes"
	"github.com/danielkennedy1/sieve/problems/grammar"
)

// Load grammar once for all tests
var testSamples []grammar.Sample
var testGrammar genomes.Grammar

func init() {
	// Load the grammar file
	f, err := os.Open("../data/lecture.bnf")
	if err != nil {
		panic("Could not open grammar file: " + err.Error())
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	testGrammar = grammar.Parse(*s)
	testGrammar.BuildRuleMap()

	// Define test samples
	testSamples = []grammar.Sample{
		{Variables: []float64{0, 0}, Output: 0.2},
		{Variables: []float64{4, 0}, Output: 4.2},
		{Variables: []float64{2, 0}, Output: 2.2},
		{Variables: []float64{5, 0}, Output: 5.2},
	}
}

// Benchmark with your actual grammar problem - small population
func BenchmarkGrammarEvolveSmall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := rand.New(rand.NewPCG(0, 0))

		pop := NewPopulation(
			100,
			0.1,
			0.7,
			2,
			genomes.NewCreateGenotype(8, r),
			grammar.NewRMSE(testSamples, testGrammar),
			genomes.NewCrossoverGenotype(r),
			genomes.NewMutateGenotype(r, 0.1),
			Tournament(25),
		)
		pop.Evolve(10) // fewer generations for benchmarking
	}
}

// Benchmark with your actual parameters from main.go
func BenchmarkGrammarEvolveFull(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := rand.New(rand.NewPCG(0, 0))

		pop := NewPopulation(
			500,
			0.1,
			0.7,
			2,
			genomes.NewCreateGenotype(8, r),
			grammar.NewRMSE(testSamples, testGrammar),
			genomes.NewCrossoverGenotype(r),
			genomes.NewMutateGenotype(r, 0.1),
			Tournament(25),
		)
		pop.Evolve(50) // Use fewer generations than your full 400 for benchmarking
	}
}

// Benchmark just the evaluation phase
func BenchmarkGrammarEvaluateAll(b *testing.B) {
	r := rand.New(rand.NewPCG(0, 0))

	pop := NewPopulation(
		500,
		0.1,
		0.7,
		2,
		genomes.NewCreateGenotype(8, r),
		grammar.NewRMSE(testSamples, testGrammar),
		genomes.NewCrossoverGenotype(r),
		genomes.NewMutateGenotype(r, 0.1),
		Tournament(25),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pop.evaluateAll()
	}
}

// Benchmark different population sizes
func BenchmarkGrammarPopSize50(b *testing.B) {
	benchmarkWithPopSize(b, 50)
}

func BenchmarkGrammarPopSize100(b *testing.B) {
	benchmarkWithPopSize(b, 100)
}

func BenchmarkGrammarPopSize500(b *testing.B) {
	benchmarkWithPopSize(b, 500)
}

func BenchmarkGrammarPopSize1000(b *testing.B) {
	benchmarkWithPopSize(b, 1000)
}

func benchmarkWithPopSize(b *testing.B, popSize int) {
	for i := 0; i < b.N; i++ {
		r := rand.New(rand.NewPCG(0, 0))

		pop := NewPopulation(
			popSize,
			0.1,
			0.7,
			2,
			genomes.NewCreateGenotype(8, r),
			grammar.NewRMSE(testSamples, testGrammar),
			genomes.NewCrossoverGenotype(r),
			genomes.NewMutateGenotype(r, 0.1),
			Tournament(25),
		)
		pop.Evolve(10)
	}
}
