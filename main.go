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

	samples := []grammar.Sample{
		{Variables: []float64{0}, Output: 0},
		{Variables: []float64{1}, Output: 3},
		{Variables: []float64{2}, Output: 14},
		{Variables: []float64{3}, Output: 39},
		{Variables: []float64{4}, Output: 84},
		{Variables: []float64{5}, Output: 155},
		{Variables: []float64{6}, Output: 258},
		{Variables: []float64{7}, Output: 399},
		{Variables: []float64{8}, Output: 584},
		{Variables: []float64{9}, Output: 819},
		{Variables: []float64{10}, Output: 1110},
		{Variables: []float64{11}, Output: 1473},
		{Variables: []float64{12}, Output: 1914},
		{Variables: []float64{13}, Output: 2439},
		{Variables: []float64{14}, Output: 3054},
		{Variables: []float64{15}, Output: 3765},
		{Variables: []float64{16}, Output: 4578},
		{Variables: []float64{17}, Output: 5499},
		{Variables: []float64{18}, Output: 6534},
		{Variables: []float64{19}, Output: 7695},
		{Variables: []float64{20}, Output: 8880},
		{Variables: []float64{21}, Output: 10101},
		{Variables: []float64{22}, Output: 11454},
		{Variables: []float64{23}, Output: 12939},
		{Variables: []float64{24}, Output: 14562},
	}

	r := rand.New(rand.NewPCG(0, 0))
	s := bufio.NewScanner(f)
	g := grammar.Parse(*s)
	g.BuildRuleMap()

	population := ea.NewPopulation(
		400,
		0.1,
		2,
		genomes.NewCreateGenotype(8, r),
		grammar.NewRMSE(samples, g),
		genomes.NewCrossoverGenotype(r),
		genomes.NewMutateGenotype(0.1),
		ea.Tournament(7),
	)

	population.Evolve(400)

	best, fitness := population.Best()
	fmt.Printf("Best fitness: %.2f\n", fitness)
	fmt.Println("Best: ", best.MapToGrammar(g, 1000))

}
