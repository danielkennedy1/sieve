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
	f, err := os.Open("data/lecture.bnf")
	if err != nil {
		fmt.Println("File not found")
		os.Exit(1)
	}
	defer f.Close()

}
