package main

import (
	"fmt"
	"math/rand"
)

func MakePopulation(size, gene_length int) [][]bool {
	population := make([][]bool, size)
	for i := range size {
		population[i] = make([]bool, gene_length)
		for j := range gene_length {
			population[i][j] = rand.Intn(2) == 1
		}
	}
	return population
}

func CreateMutation(parent_gene []bool, percentage_chance float64) []bool {
	mutated_gene := make([]bool, len(parent_gene))
	copy(mutated_gene, parent_gene)
	for i := range parent_gene {
		chance := rand.Float64()
		if chance > percentage_chance {
			mutated_gene[i] = !parent_gene[i]
		}
	}
	return mutated_gene
}

func GetFitness(individual []bool) float64 {
	length := len(individual)
	weight := 0
	for i := range length {
		if individual[i] {
			weight++
		}
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

	for i := range len(population) {
		fmt.Println(GetFitness(population[i]))
	}

}
