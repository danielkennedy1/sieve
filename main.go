package main

import (
	"bufio"
	"fmt"
	"os"

	"math/rand/v2"

	"github.com/danielkennedy1/sieve/ea"
	"github.com/danielkennedy1/sieve/genomes"
	"github.com/danielkennedy1/sieve/problems/grammar"
)

func main() {
	f, err := os.Open("data/lecture.bnf")
	if err != nil {
		fmt.Println("File not found")
		os.Exit(1)
	}
	defer f.Close()

	// main.go

	samples := []grammar.Sample{
		// Note: Variables is {a, b, c}
		{Variables: []float64{0, 0, 0}, Output: 0},
		{Variables: []float64{1, 0, 0}, Output: 3},
		{Variables: []float64{2, 0, 0}, Output: 14},
		{Variables: []float64{3, 0, 0}, Output: 39},
		{Variables: []float64{4, 0, 0}, Output: 84},
		{Variables: []float64{5, 0, 0}, Output: 155},
		{Variables: []float64{6, 0, 0}, Output: 258},
		{Variables: []float64{7, 0, 0}, Output: 399},
		{Variables: []float64{8, 0, 0}, Output: 584},
		{Variables: []float64{9, 0, 0}, Output: 819},
		{Variables: []float64{10, 0, 0}, Output: 1110},
		// The outputs below differ from your original list
		{Variables: []float64{11, 0, 0}, Output: 1463},
		{Variables: []float64{12, 0, 0}, Output: 1884},
		{Variables: []float64{13, 0, 0}, Output: 2379},
		{Variables: []float64{14, 0, 0}, Output: 2954},
		{Variables: []float64{15, 0, 0}, Output: 3615},
		{Variables: []float64{16, 0, 0}, Output: 4368},
		{Variables: []float64{17, 0, 0}, Output: 5219},
		{Variables: []float64{18, 0, 0}, Output: 6174},
		{Variables: []float64{19, 0, 0}, Output: 7239},
		{Variables: []float64{20, 0, 0}, Output: 8420},
		{Variables: []float64{21, 0, 0}, Output: 9723},
		{Variables: []float64{22, 0, 0}, Output: 11154},
		{Variables: []float64{23, 0, 0}, Output: 12719},
		{Variables: []float64{24, 0, 0}, Output: 14424},
	}

	r := rand.New(rand.NewPCG(0, 0))
	s := bufio.NewScanner(f)
	g := grammar.Parse(*s)

	population := ea.NewPopulation(
		250,
		0.1,
		0.7,
		15,
		genomes.NewCreateGenotype(48, r),
		grammar.NewRMSE(samples, g),
		genomes.NewCrossoverGenotype(r),
		genomes.NewMutateGenotype(r, 0.1),
		ea.Tournament(7),
	)

	population.Evolve(250)

	best, fitness := population.Best()
	fmt.Printf("Best fitness: %.2f\n", fitness)
	fmt.Println("Best: ", best.MapToGrammar(g, 100).String())

}
