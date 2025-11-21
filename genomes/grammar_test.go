package genomes_test

import (
	"testing"

	"github.com/danielkennedy1/sieve/genomes"
)

func TestLectureExampleGrammarAndGenome(t *testing.T) {
	lectureExampleGrammar := genomes.NewTestLectureExampleGrammar()

	genotype := genomes.Genotype{
		Genes: []uint8{220, 149, 147, 220, 144, 55, 36, 170},
	}

	want := "a + 0.2"
	got := genotype.MapToGrammar(lectureExampleGrammar, 100).String()

	if got != want {
		t.Errorf("Got unexpected string from grammar. got '%s', want '%s'", got, want)
	}
}
