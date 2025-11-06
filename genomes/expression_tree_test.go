package genomes_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/danielkennedy1/sieve/genomes"
)

func TestExpressionTree(t *testing.T) {
	expr := genomes.NonTerminal{genomes.Add, genomes.Primitive{3}, genomes.Variable{&[]float64{2}, 0}}

	if expr.GetValue() != 5 {
		t.Errorf("Expression is not expected value, expected 5, got %f\n", expr.GetValue())
	}
}

func TestRandomExpressionTree(t *testing.T) {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	const numVars = 2
	maxDepth := 3
	variableValues := make([]float64, numVars)

	expr := genomes.RandomFormula(maxDepth, &variableValues, numVars, r)

	t.Logf("Generated Expression Tree: %+v\n", expr)

	variableValues[0] = 5.0
	variableValues[1] = 2.0
	t.Logf("Evaluating with x0=%.2f, x1=%.2f\n", variableValues[0], variableValues[1])

	result := expr.GetValue()
	t.Logf("Result: %f\n\n", result)
}
