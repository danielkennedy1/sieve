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

	variables := []float64{1.0, 2.0, 3.0}

	//
	target := genomes.NonTerminal{
		Operator: genomes.Add,
		Left: genomes.NonTerminal{
			Operator: genomes.Add,
			Left: genomes.NonTerminal{
				Operator: genomes.Multiply,
				Left: genomes.NonTerminal{
					Operator: genomes.Multiply,
					Left:     genomes.Variable{Variables: &variables, Index: 0},
					Right:    genomes.Variable{Variables: &variables, Index: 0},
				},
				Right: genomes.NonTerminal{
					Operator: genomes.Multiply,
					Left:     genomes.Variable{Variables: &variables, Index: 0},
					Right:    genomes.Variable{Variables: &variables, Index: 0},
				},
			},
			Right: genomes.NonTerminal{
				Operator: genomes.Multiply,
				Left: genomes.NonTerminal{
					Operator: genomes.Multiply,
					Left:     genomes.Variable{Variables: &variables, Index: 0},
					Right:    genomes.Variable{Variables: &variables, Index: 0},
				},
				Right: genomes.Variable{Variables: &variables, Index: 0},
			},
		},
		Right: genomes.NonTerminal{
			Operator: genomes.Add,
			Left: genomes.NonTerminal{
				Operator: genomes.Multiply,
				Left:     genomes.Variable{Variables: &variables, Index: 0},
				Right:    genomes.Variable{Variables: &variables, Index: 0},
			},
			Right: genomes.Variable{Variables: &variables, Index: 0},
		},
	}

	samples := make([]grammar.Sample, 0)

	for i := -100.0; i < 100.0; i += 1 {
		variables[0] = i
		samples = append(
			samples,
			grammar.Sample{
				Variables: []float64{i},
				Output:    target.GetValue(),
			},
		)
	}

	r := rand.New(rand.NewPCG(0, 0))
	s := bufio.NewScanner(f)
	g := grammar.Parse(*s)

	population := ea.NewPopulation(
		250,
		0.1,
		2,
		genomes.NewCreateGenotype(8, r),
		grammar.NewRMSE(samples, g),
		genomes.NewCrossoverGenotype(r),
		genomes.NewMutateGenotype(r, 0.1),
		ea.Tournament(25),
	)

	population.Evolve(100)

	best, fitness := population.Best()
	fmt.Printf("Best fitness: %.2f\n", fitness)
	fmt.Println("Best: ", best.MapToGrammar(g, 1000))
}
