package main

import (
	"bufio"
	"fmt"
	"os"

	"math/rand/v2"

	"github.com/danielkennedy1/sieve/ea"
	"github.com/danielkennedy1/sieve/genomes"
	"github.com/danielkennedy1/sieve/problems/grammar"
	// "github.com/danielkennedy1/sieve/problems/grammar"
	// "github.com/danielkennedy1/sieve/problems/grammar"
)

func main() {
	f, err := os.Open("data/lecture.bnf")
	if err != nil {
		fmt.Println("File not found")
		os.Exit(1)
	}
	defer f.Close()

	samples := []grammar.Sample{
		{Variables: []float64{0, 0}, Output: 0.2},
		{Variables: []float64{4, 0}, Output: 4.2},
		{Variables: []float64{2, 0}, Output: 2.2},
		{Variables: []float64{5, 0}, Output: 5.2},
	}

	// /func NewPopulation[G any](
	//     size int,
	//     mutationRate float64,
	//     eliteCount int,
	//     create func() G,
	//     evaluate func(G) float64,
	//     crossover func(G, G) (G, G),
	//     mutate func(G) G,
	//     selector func([]float64, int) []int,
	// ) *Population[G] {
	// population for grammar problem

	r := rand.New(rand.NewPCG(0, 0))
	s := bufio.NewScanner(f)
	g := grammar.Parse(*s)

	population := ea.NewPopulation(
		100,
		0.1,
		2,
		genomes.NewCreateGenotype(8, r),
		grammar.NewRMSE(samples, g),
		genomes.NewCrossoverGenotype(r),
		genomes.NewMutateGenotype(r, 0.1),
		ea.Tournament(25),
	)

	population.Evolve(20)

	best, fitness := population.Best()
	fmt.Printf("Best fitness: %.2f\n", fitness)
	fmt.Println("Best: ", best.MapToGrammar(g, 1000))

}
