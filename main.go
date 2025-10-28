package main

import (
	"math/rand"
	"fmt"
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

func GetFitness(individual []bool) float64 {
	length := len(individual)
	weight := 0
	for i := range(length) {
		if individual[i] { weight++ }
	}
	return float64(weight) / float64(length)
}

func SinglePointCrossover(a, b []bool) ([]bool, []bool) {
	half := len(a) / 2
	t := make([]bool, half)
	copy(t, a[half:])
	copy(a[half:], b[half:])
	copy(b[half:], t)
	return a, b
}

func main() {
	//population := MakePopulation(100, 8)

	//fmt.Println(population)

	//for i := range(len(population)) {
	//	fmt.Println(GetFitness(population[i]))
	//}

	a := []bool{true, true}
	b := []bool{false, false}

	a, b = SinglePointCrossover(a, b)

	fmt.Println(a)
	fmt.Println(b)
}
