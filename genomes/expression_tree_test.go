package genomes_test

import (
	"testing"

	"github.com/danielkennedy1/sieve/genomes"
)

func TestExpressionTree(t *testing.T){
	expr := genomes.NonTerminal{genomes.Add, genomes.Primitive{3}, genomes.Variable{&[]float64{2}, 0}}

	if expr.GetValue() != 5 {
		t.Errorf("Expression is not expected value, expected 5, got %f\n", expr.GetValue())
	}
}
