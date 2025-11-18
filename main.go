package main

import (
	"bufio"
	"fmt"
	"os"

	//"math/rand/v2"

	//"github.com/danielkennedy1/sieve/ea"
	//"github.com/danielkennedy1/sieve/genomes"
	"github.com/danielkennedy1/sieve/problems/grammar"
	//"github.com/danielkennedy1/sieve/problems/grammar"
)

func main() {
	//lectureExampleGrammar := grammar.NewLectureExampleGrammar()

	//genotype := genomes.Genotype{
	//	Genes: []uint8{220, 149, 147, 220, 144, 55, 36, 170},
	//}

	//str := genotype.MapToGrammar(lectureExampleGrammar)

	//fmt.Println(str.String())

	f, err := os.Open("data/lecture.bnf")
	if err != nil {
		fmt.Println("File not found")
		os.Exit(1)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	grammar := grammar.Parse(*scanner)

	fmt.Println(grammar)
}
