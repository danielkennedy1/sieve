package expression_tree_test

import (
	"encoding/csv"
	"strings"
	"testing"

	"github.com/danielkennedy1/sieve/genomes"
	"github.com/danielkennedy1/sieve/problems/expression_tree"
)

func TestLoadSamples(t *testing.T) {
	// Given

	csvData := `y,x0,x1
-1.6,-3.0,2.1
3.2,1.0,-0.5`

	want := []expression_tree.Sample{
		{[]float64{-3.0, 2.1}, -1.6},
		{[]float64{1.0, -0.5}, 3.2},
	}

	reader := csv.NewReader(strings.NewReader(csvData))

	// When

	samples, err := expression_tree.LoadSamples(reader)


	// Then

	if err != nil {
		t.Errorf("Reading samples failed: %s", err)
	}

	for i := range samples {
		for j := range samples[i].Variables {
			if samples[i].Variables[j] != want[i].Variables[j] {
				t.Errorf("Mismatch in variables: got %f, want %f", samples[i].Variables[j], want[i].Variables[j])
			}
		}
		if samples[i].Output != want[i].Output {
			t.Errorf("Mismatch in output: got %f want %f", samples[i].Output, want[i].Output)
		}
	}
}

func TestMeanSquaredError(t *testing.T) {
	variables := make([]float64, 3)

	// x1 + ( 3 * x2 - x3 / 2 )
	expr := genomes.NonTerminal{
		Operator: genomes.Add,
		Left: genomes.Variable{Variables: &variables, Index: 0},
		Right: genomes.NonTerminal{
			Operator: genomes.Subtract,
			Left: genomes.NonTerminal{
				Operator: genomes.Multiply,
				Left: genomes.Primitive{Value: 3},
				Right: genomes.Variable{Variables: &variables, Index: 1},
			},
			Right: genomes.NonTerminal{
				Operator: genomes.Divide,
				Left: genomes.Variable{Variables: &variables, Index: 2},
				Right: genomes.Primitive{Value: 2},
			},
		},
	}

	samples := []expression_tree.Sample{
		{[]float64{0, 0, 0}, 0},
		{[]float64{4, 2, 2}, 9},
		{[]float64{2, 1, 8}, 1},
		{[]float64{0, 3, 1}, 8.5},
	}

	mse := expression_tree.MeanSquaredError(expr, &variables, &samples)

	if mse != 0 {
		t.Errorf("Unexpected MSE, wanted 0, got %f", mse)
	}
}
