package main

import (
	"math/rand"
	"fmt"
	"unsafe"
)

func MakePopulation(size, gene_length int) [][]bool {
	population := make([][]bool, size)
	for i := range(size) {
		population[i] = make([]bool, gene_length)
		for j := range(gene_length) {
			population[i][j] = rand.Intn(2) == 1
		}
	}
	return population
}

func main() {
	population := MakePopulation(100, 8)

	fmt.Println(population)
}
