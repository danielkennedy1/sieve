package genomes_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/danielkennedy1/sieve/genomes"
)

func TestExpressionTreeSingle(t *testing.T) {
	expr := genomes.NonTerminal{genomes.Add, genomes.Primitive{3}, genomes.Variable{&[]float64{2}, 0}}

	if expr.GetValue() != 5 {
		t.Errorf("Expression is not expected value, expected 5, got %f\n", expr.GetValue())
	}
}

func TestExpressionTreeParameterized(t *testing.T) {
	variables := make([]float64, 3)

	// x1 + ( 3 * x2 - x3 / 2 )
	expr := genomes.NonTerminal{
		genomes.Add,
		genomes.Variable{&variables, 0},
		genomes.NonTerminal{
			genomes.Subtract,
			genomes.NonTerminal{
				genomes.Multiply,
				genomes.Primitive{3},
				genomes.Variable{&variables, 1},
			},
			genomes.NonTerminal{
				genomes.Divide,
				genomes.Variable{&variables, 2},
				genomes.Primitive{2},
			},
		},
	}

	tests := []struct {
		sample []float64
		want   float64
	}{
		{[]float64{0, 0, 0}, 0},
		{[]float64{4, 2, 2}, 9},
		{[]float64{2, 1, 8}, 1},
		{[]float64{0, 3, 1}, 8.5},
	}

	for idx, in := range tests {
		t.Run(fmt.Sprintf("Tree %d", idx), func(t *testing.T) {
			variables = in.sample
			got := expr.GetValue()

			if got != in.want {
				t.Errorf("Got %f in test index %d, expected %f", got, idx, in.want)
			}
		})
	}
}

func TestRandomExpressionTree(t *testing.T) {
	r := rand.New(rand.NewPCG(0, 0))

	const numVars = 30
	maxDepth := 20
	variableValues := make([]float64, numVars)

	constants := []float64{0.1, 0.2, 0.3, 0.4}

	expr := genomes.RandomFormula(maxDepth, &variableValues, &constants, numVars, r)

	t.Logf("Generated Expression Tree: %+v\n", expr)

	variableValues[0] = 5.0
	variableValues[1] = 2.0
	t.Logf("Evaluating with x0=%.2f, x1=%.2f\n", variableValues[0], variableValues[1])

	result := expr.GetValue()
	t.Logf("Result: %f\n\n", result)
}
