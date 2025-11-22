package ea

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

type Population[G any] struct {
	genomes   []G
	fitnesses []float64

	evaluate  func(G) float64
	crossover func(G, G) (G, G)
	mutate    func(G) G
	selector  func([]float64, int) []int

	mutationRate float64
	eliteCount   int
}

func NewPopulation[G any](
	size int,
	mutationRate float64,
	eliteCount int,
	create func() G,
	evaluate func(G) float64,
	crossover func(G, G) (G, G),
	mutate func(G) G,
	selector func([]float64, int) []int,
) *Population[G] {
	genomes := make([]G, size)
	for i := range genomes {
		genomes[i] = create()
	}

	return &Population[G]{
		genomes:      genomes,
		fitnesses:    make([]float64, size),
		evaluate:     evaluate,
		crossover:    crossover,
		mutate:       mutate,
		selector:     selector,
		mutationRate: mutationRate,
		eliteCount:   eliteCount,
	}
}
func (p *Population[G]) Evolve(generations int) {
	time_list := []time.Duration{}
	for generation := range generations {
		start := time.Now()
		fmt.Printf("Generation %d\n", generation)

		var wg sync.WaitGroup
		for i, g := range p.genomes {
			wg.Add(1)
			go func(idx int, genome G) {
				defer wg.Done()
				p.fitnesses[idx] = p.evaluate(genome)
			}(i, g)
		}
		wg.Wait()

		parentIndices := p.selector(p.fitnesses, len(p.genomes))

		offspring := make([]G, len(p.genomes))
		for i := 0; i < len(parentIndices)-1; i += 2 {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				idx1, idx2 := parentIndices[idx], parentIndices[idx+1]
				c1, c2 := p.crossover(p.genomes[idx1], p.genomes[idx2])

				if rand.Float64() < p.mutationRate {
					c1 = p.mutate(c1)
				}
				if rand.Float64() < p.mutationRate {
					c2 = p.mutate(c2)
				}

				offspring[idx] = c1
				offspring[idx+1] = c2
			}(i)
		}
		wg.Wait()

		if p.eliteCount > 0 {
			elite := p.getElite()
			offspring = offspring[:len(offspring)-p.eliteCount]
			offspring = append(offspring, elite...)
		}

		p.genomes = offspring

		elapsed := time.Since(start)
		time_list = append(time_list, elapsed)
	}
	sum := 0 * time.Nanosecond
	for i := 0; i < len(time_list); i++ {
		sum += (time_list[i])
	}

	avg_time := sum / time.Duration(len(time_list))
	fmt.Printf("Average time: %s\n", avg_time)
}

func (p *Population[G]) getElite() []G {
	indices := make([]int, len(p.genomes))
	for i := range indices {
		indices[i] = i
	}

	for i := range indices {
		for j := i + 1; j < len(indices); j++ {
			if p.fitnesses[indices[j]] > p.fitnesses[indices[i]] {
				indices[i], indices[j] = indices[j], indices[i]
			}
		}
	}

	elite := make([]G, p.eliteCount)
	for i := 0; i < p.eliteCount; i++ {
		elite[i] = p.genomes[indices[i]]
	}
	return elite
}

func (p *Population[G]) Best() (G, float64) {
	bestIdx := 0
	bestFit := p.fitnesses[0]

	for i, f := range p.fitnesses {
		if f > bestFit {
			bestIdx = i
			bestFit = f
		}
	}

	return p.genomes[bestIdx], bestFit
}
