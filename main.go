package main

import (
	"bufio"
	"fmt"
	"math/rand/v2"
	"os"
	"time"

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

	fmt.Println(target.String())
	r := rand.New(rand.NewPCG(0, 0))
	s := bufio.NewScanner(f)
	g := grammar.Parse(*s)
	g.BuildRuleMap()

	targetExpressionString := "(a + (a * a) + (a * a * a)) + (a * a * a * a)"
	numSamplesToGenerate := 50
	initialVariables := []float64{0.0}

	samples, err := grammar.GenerateSamples(targetExpressionString, numSamplesToGenerate, initialVariables, g)

	if err != nil {
		fmt.Printf("Error generating samples: %v\n", err)
		return
	}

	population := ea.NewPopulation(
		500,
		0.1,
		0.7,
		15,
		genomes.NewCreateGenotype(48, r),
		grammar.NewRMSE(samples, g),
		genomes.NewCrossoverGenotype(r),
		genomes.NewMutateGenotype(r, 0.1),
		ea.Tournament(7),
	)

	start := time.Now()

	population.Evolve(100)

	elapsed := time.Since(start)

	best, fitness := population.Best()
	fmt.Printf("Best fitness: %.2f\n", fitness)
	fmt.Println("Best: ", best.MapToGrammar(g, 100).String())
	fmt.Printf("Elapsed time: %s\n", elapsed)

}
