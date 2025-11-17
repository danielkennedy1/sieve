package main

import (
	"fmt"
	//"math/rand/v2"

	//"github.com/danielkennedy1/sieve/ea"
	"github.com/danielkennedy1/sieve/genomes"
	"github.com/danielkennedy1/sieve/problems/grammar"
)

func main() {
	lectureExampleGrammar := grammar.NewLectureExampleGrammar()

	genotype := genomes.Genotype{
		Genes: []uint8{220, 149, 147, 220, 144, 55, 36, 170},
	}

	str := genotype.MapToGrammar(lectureExampleGrammar)

	fmt.Println(str.String())
}
