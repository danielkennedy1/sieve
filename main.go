package main

import (
	"fmt"
	"math/rand"
	"sort"
)

func MakePopulation(size, geneLength int) [][]bool {
	population := make([][]bool, size)
	for i := range size {
		population[i] = make([]bool, geneLength)
		for j := range geneLength {
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
	geneLength := len(population[0])

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

	nextGeneration := make([][]bool, 0, populationSize)
	nextGeneration = append(nextGeneration, elites...)

	// Crossover and Mutation
	for len(nextGeneration) < populationSize {
		p1 := selected[rand.Intn(len(selected))]
		p2 := selected[rand.Intn(len(selected))]

		var child1, child2 []bool
		if rand.Float64() < crossoverRate {
			point := rand.Intn(geneLength-1) + 1
			child1, child2 = SinglePointCrossover(p1, p2, point)
		} else {
			child1 = clone(p1)
			child2 = clone(p2)
			Mutate(child1, mutationRate)
			Mutate(child2, mutationRate)

		}

		nextGeneration = append(nextGeneration, child1)
		if len(nextGeneration) < populationSize {
			nextGeneration = append(nextGeneration, child2)
		}
	}

	return nextGeneration
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

func GetBestFitness(population [][]bool) (float64, int) {
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

func GetMetrics(prevPopulation [][]bool, newPopulation [][]bool) ([]float64, float64, float64) {

	fitness := EvaluateFitnessPopulation(newPopulation)
	diversity := EvaluateDiversityPopulation(newPopulation)
	improvement := EvaluateImprovementPopulation(prevPopulation, newPopulation)

	return fitness, diversity, improvement
}

func main() {

	populationSize := 500
	geneLength := 1000
	tournamentSize := 10
	crossoverRate := 0.7
	mutationRate := 0.01
	elitism := 20
	generations := 10
	fitnessMetricsList := make([][]float64, 0, generations)
	diversityMetricsList := make([]float64, 0, generations)
	improvementMetricsList := make([]float64, 0, generations)

	population := MakePopulation(populationSize, geneLength)
	prevPopulation := population

	for generation := 0; generation < generations; generation++ {
		best, _ := GetBestFitness(population)
		fmt.Printf("Gen %d | Best fitness: %.4f\n", generation, best)

		newPopulation := EvolveGeneration(population, tournamentSize, mutationRate, crossoverRate, elitism)

		fitness, diversity, improvement := GetMetrics(prevPopulation, newPopulation)

		fitnessMetricsList = append(fitnessMetricsList, fitness)
		diversityMetricsList = append(diversityMetricsList, diversity)
		improvementMetricsList = append(improvementMetricsList, improvement)

		// Move to next generation
		prevPopulation = population
		population = newPopulation
	}

	finalBest, index := GetBestFitness(population)
	fmt.Printf("Final best fitness: %.4f\n", finalBest)
	fmt.Println("Best individual:", population[index])

	// Charts

	CreateChart(fitnessMetricsList, diversityMetricsList, improvementMetricsList)

}
