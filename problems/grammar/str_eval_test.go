package grammar

import (
	"testing"

	"github.com/danielkennedy1/sieve/genomes"
)

func TestRMSEfromGrammar(t *testing.T) {
	// variables is passed into RMSE to match your existing API,
	// but the expression string uses x0, x1, x2 ...

	grammar := genomes.NewTestLectureExampleGrammar()

	genotype := genomes.Genotype{
		Genes: []uint8{220, 149, 147, 220, 144, 55, 36, 170},
	}
	// a + 0.2
	exprStr := genotype.MapToGrammar(grammar, 1000)

	samples := []Sample{
		{Variables: []float64{0, 0}, Output: 0.2},
		{Variables: []float64{4, 0}, Output: 4.2},
		{Variables: []float64{2, 0}, Output: 2.2},
		{Variables: []float64{5, 0}, Output: 5.2},
	}

	got := RMSE(exprStr.String(), grammar, samples)

	want := 0.0

	if got != want {
		t.Errorf("Got RMSE %f, want %f", got, want)
	}
}
