package main

import (
	"fmt"
	"math/rand/v2"

	"github.com/danielkennedy1/sieve/ea"
	"github.com/danielkennedy1/sieve/genomes"
	"github.com/danielkennedy1/sieve/problems/expression_tree"
)

func main() {
	constants := []float64{1.0, 2.0, 3.0}
	variables := []float64{1.0, 2.0, 3.0}

	// x0 + ( 3 * x1 - x2 / 2 )
	target := genomes.NonTerminal{
		Operator: genomes.Add,
		Left:     genomes.Variable{Variables: &variables, Index: 0},
		Right: genomes.NonTerminal{
			Operator: genomes.Subtract,
			Left: genomes.NonTerminal{
				Operator: genomes.Multiply,
				Left:     genomes.Primitive{Value: constants[2]},
				Right:    genomes.Variable{Variables: &variables, Index: 1},
			},
			Right: genomes.NonTerminal{
				Operator: genomes.Divide,
				Left:     genomes.Variable{Variables: &variables, Index: 2},
				Right:    genomes.Primitive{Value: constants[1]},
			},
		},
	}

	samples := make([]expression_tree.Sample, 0)

	for i := -6.0; i < 6.0; i += 1 {
		for j := -6.0; j < 6.0; j += 1 {
			for k := -6.0; k < 6.0; k += 1 {
				variables[0] = i
				variables[1] = j
				variables[2] = k
				samples = append(
					samples,
					expression_tree.Sample{
						Variables: []float64{i, j, k},
						Output:    target.GetValue(),
					},
				)
			}
		}
	}

	const maxDepth = 5

	r := rand.New(rand.NewPCG(42, 42))

	pop := ea.NewPopulation(
		500,  // Population size
		0.05, // Mutation rate
		10,    // Elite count
		func() genomes.Expression {
			return genomes.RandomFormula(maxDepth, &variables, &constants, len(variables), r)
		},
		expression_tree.NewRootMeanSquaredError(&variables, &samples),
		genomes.NewCrossoverExpression(r, maxDepth),
		genomes.NewMutateExpression(constants),
		ea.Tournament(3),
	)

	fmt.Println("Evolving...")

	pop.Evolve(100)

	best, fitness := pop.Best()
	fmt.Printf("Best fitness: %.2f\n", fitness)
	fmt.Println("Best: ", best.String())
}
