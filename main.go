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

	fmt.Println(genomes.ValidateGrammar(lectureExampleGrammar))
}
