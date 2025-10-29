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

func SinglePointCrossover(a, b []bool, point int) ([]bool, []bool) {
	if point == 0 {
		return a, b
	}
	t := make([]bool, point)
	copy(t, a[point:])
	copy(a[point:], b[point:])
	copy(b[point:], t)
	return a, b
}

func Mutate(individual []bool, point int) {
	individual[point] = !individual[point]
}

func main() {
	population := MakePopulation(100, 30)

	fmt.Println(population)

	for i := range(len(population)) {
		fmt.Println(GetFitness(population[i]))
	}

}
