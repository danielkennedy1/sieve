package grammar

import (
	"testing"

	"github.com/danielkennedy1/sieve/genomes"
)

func TestRMSEfromGrammar(t *testing.T) {

	grammar := genomes.NewTestLectureExampleGrammar()

	genotype := genomes.Genotype{
		Genes: []uint8{220, 149, 147, 220, 144, 55, 36, 170},
	}
	// a + 0.2
	samples := []Sample{
		{Variables: []float64{0, 0}, Output: 0.2},
		{Variables: []float64{4, 0}, Output: 4.2},
		{Variables: []float64{2, 0}, Output: 2.2},
		{Variables: []float64{5, 0}, Output: 5.2},
	}

	rmseFunc := NewRMSE(samples, grammar, 0.001, 100)
	got := rmseFunc(genotype)

	want := -0.007

	if got != want {
		t.Errorf("Got RMSE %f, want %f", got, want)
	}
}
