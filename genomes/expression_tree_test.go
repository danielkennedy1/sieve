package genomes_test

import (
	"fmt"
	"math"
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

func TestRandomExpression(t *testing.T) {

	tests := []struct {
		numVar    int
		maxDepth  int
		varValues []float64
		constants []float64
		rand      *rand.Rand
	}{
		{numVar: 2,
			maxDepth:  5,
			varValues: []float64{1, 2},
			constants: []float64{0.1, 0.2, 0.3, 0.4},
			rand:      rand.New(rand.NewPCG(0, 0))},
		{numVar: 2,
			maxDepth:  10,
			varValues: []float64{1, 2},
			constants: []float64{0.1, 0.2, 0.3, 0.4},
			rand:      rand.New(rand.NewPCG(0, 0))},
		{numVar: 2,
			maxDepth:  20,
			varValues: []float64{1, 2},
			constants: []float64{0.1, 0.2, 0.3, 0.4},
			rand:      rand.New(rand.NewPCG(0, 0))},
	}

	for _, in := range tests {
		t.Run(fmt.Sprintf("Depth: %d", in.maxDepth), func(t *testing.T) {
			r := in.rand
			var numVars = in.numVar
			maxDepth := in.maxDepth
			variableValues := in.varValues
			constants := in.constants

			expr := genomes.RandomFormula(maxDepth, &variableValues, &constants, numVars, r)

			result := expr.GetValue()

			if math.IsNaN(result) || math.IsInf(result, 0) || result == math.MaxFloat64 {
				t.Errorf("Got %f as result which is not expected type", result)
			}
		})
	}
}

func TestMakeMutateExpression(t *testing.T) {
	constants := []float64{1.0, 2.0, 3.0, 4.0}
	variables := []float64{0.1, 0.2, 0.3, 0.4}

	// x0 + ( 3 * x1 - x2 / 2 )
	expr := genomes.NonTerminal{
		genomes.Add,
		genomes.Variable{&variables, 0},
		genomes.NonTerminal{
			genomes.Subtract,
			genomes.NonTerminal{
				genomes.Multiply,
				genomes.Primitive{constants[2]},
				genomes.Variable{&variables, 1},
			},
			genomes.NonTerminal{
				genomes.Divide,
				genomes.Variable{&variables, 2},
				genomes.Primitive{constants[1]},
			},
		},
	}
	MutateExpression := genomes.NewMutateExpression(constants)

	new_expr := MutateExpression(expr)

	if expr.Compare(new_expr) {
		t.Errorf("Expression did not change: %s == %s", expr, new_expr)
	}
}

func TestExpressionCrossover(t *testing.T) {
	expr1 := genomes.RandomFormula(20, &[]float64{1, 2, 3}, &[]float64{0.1, 0.2, 0.3}, 3, rand.New(rand.NewPCG(0, 1)))
	expr2 := genomes.RandomFormula(20, &[]float64{1, 2, 3}, &[]float64{0.1, 0.2, 0.3}, 3, rand.New(rand.NewPCG(0, 3)))

	c1, c2 := genomes.Crossover(expr1, expr2, rand.New(rand.NewPCG(2, 0)), 20)

	t.Logf("Parent 1: %s\n", expr1.String())
	t.Logf("Parent 2: %s\n", expr2.String())
	t.Logf("Child 1: %s\n", c1.String())
	t.Logf("Child 2: %s\n", c2.String())
}
