package main

import (
	"fmt"
	"math/rand/v2"

	"github.com/danielkennedy1/sieve/ea"
	"github.com/danielkennedy1/sieve/genomes"
	"github.com/danielkennedy1/sieve/problems/expression_tree"
)


func main() {
	constants := []float64{0.1, 0.2, 0.3, 0.4}
	variables := []float64{1.0, 2.0, 3.0}

	// x1 + ( 3 * x2 - x3 / 2 )
	target := genomes.NonTerminal{
		Operator: genomes.Add,
		Left: genomes.Variable{Variables: &variables, Index: 0},
		Right: genomes.NonTerminal{
			Operator: genomes.Subtract,
			Left: genomes.NonTerminal{
				Operator: genomes.Multiply,
				Left: genomes.Primitive{Value: constants[2]},
				Right: genomes.Variable{Variables: &variables, Index: 1},
			},
			Right: genomes.NonTerminal{
				Operator: genomes.Divide,
				Left: genomes.Variable{Variables: &variables, Index: 2},
				Right: genomes.Primitive{Value: constants[1]},
			},
		},
	}

	samples := make([]expression_tree.Sample, 0)

	for i := -6.0; i < 6.0; i += 0.1 {
		for j := -6.0; j < 6.0; j += 0.1 {
			for k := -6.0; k < 6.0; k += 0.1 {
				variables[0] = i
				variables[1] = j
				variables[2] = k
				samples = append(
					samples,
					expression_tree.Sample{
						Variables: []float64{i, j, k},
						Output: target.GetValue(),
					},
				)
			}
		}
	}

    const genomeSize = 200

	const maxDepth = 10

	r := rand.New(rand.NewPCG(42, 42))

	rmseFitness := expression_tree.NewRootMeanSquaredError(&variables, &samples)
	mutate := genomes.NewMutateExpression(constants)
    
    pop := ea.NewPopulation(
        100, // Population size
        0.05, // Mutation rate
        2, // Elite count
        func() genomes.Expression{return genomes.RandomFormula(maxDepth, &variables, &constants, len(variables), r)},
		rmseFitness,
		//genomes.SinglePointCrossover, // TODO:
        mutate,
        ea.Tournament(3),
    )
    
    pop.Evolve(100)
    
    _, fitness := pop.Best()
    fmt.Printf("Best fitness: %.2f\n", fitness)
}
