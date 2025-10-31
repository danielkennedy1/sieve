package main

import (
	"fmt"
	"math/rand"
	"sort"
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

func clone(ind []bool) []bool {
	return append([]bool(nil), ind...)
}

func performTournament(population [][]bool, tournamentSize int, numSelected int) [][]bool {
	selected := make([][]bool, numSelected)

	for i := 0; i < numSelected; i++ {
		best := population[rand.Intn(len(population))]
		for j := 1; j < tournamentSize; j++ {
			competitor := population[rand.Intn(len(population))]
			if GetFitness(competitor) > GetFitness(best) {
				best = competitor
			}
		}
		selected[i] = best
	}

	return selected
}
func EvolveGeneration(population [][]bool, tournamentSize int, mutationRate float64, crossoverRate float64, elitism int) [][]bool {
	populationSize := len(population)
	geneLen := len(population[0])

	// Elitism
	indices := make([]int, populationSize)
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(i, j int) bool {
		return GetFitness(population[indices[i]]) > GetFitness(population[indices[j]])
	})
	elites := make([][]bool, 0, elitism)
	for i := 0; i < elitism && i < populationSize; i++ {
		elites = append(elites, clone(population[indices[i]]))
	}

	// Tournament
	selected := performTournament(population, tournamentSize, populationSize)

	next := make([][]bool, 0, populationSize)
	next = append(next, elites...)

	// Crossover and Mutation
	for len(next) < populationSize {
		p1 := selected[rand.Intn(len(selected))]
		p2 := selected[rand.Intn(len(selected))]

		var child1, child2 []bool
		if rand.Float64() < crossoverRate {
			point := rand.Intn(geneLen-1) + 1
			child1, child2 = SinglePointCrossover(p1, p2, point)
		} else {
			child1 = clone(p1)
			child2 = clone(p2)
			Mutate(child1, mutationRate)
			Mutate(child2, mutationRate)

		}

		next = append(next, child1)
		if len(next) < populationSize {
			next = append(next, child2)
		}
	}

	return next
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

func bestFitness(population [][]bool) (float64, int) {
	bestIndex := 0
	bestFitness := GetFitness(population[0])
	for i := 1; i < len(population); i++ {
		f := GetFitness(population[i])
		if f > bestFitness {
			bestFitness = f
			bestIndex = i
		}
	}
	return bestFitness, bestIndex
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

func Mutate(individual []bool, rate float64) {
	for i := range individual {
		if rand.Float64() < rate {
			individual[i] = !individual[i]
		}
	}
}

func main() {

	populationSize := 100
	geneLength := 200
	tournamentSize := 5
	crossoverRate := 0.8
	mutationRate := 0.01
	elitism := 2
	generations := 50

	population := MakePopulation(populationSize, geneLength)

	for generation := 0; generation < generations; generation++ {
		best, _ := bestFitness(population)
		fmt.Printf("Gen %d | Best fitness: %.4f\n", generation, best)
		population = EvolveGeneration(population, tournamentSize, mutationRate, crossoverRate, elitism)
	}

	finalBest, index := bestFitness(population)
	fmt.Printf("Final best fitness: %.4f\n", finalBest)
	fmt.Println("Best individual:", population[index])

}
