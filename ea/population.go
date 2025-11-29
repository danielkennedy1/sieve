package ea

import (
	"fmt"
	"math"
	"math/rand/v2"
)

type Population[G any] struct {
    genomes      []G
    fitnesses    []float64
    
    evaluate     func(G) float64
    crossover    func(G, G) (G, G)
    mutate       func(G) G
    selector     func([]float64, int) []int
    
    mutationRate float64
    eliteCount   int
}

func NewPopulation[G any](
    size int,
    mutationRate float64,
	crossoverRate float64,
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
	for generation := range generations {
		fmt.Printf("Generation %d\n", generation)
		for i, g := range p.Genomes {
			p.Fitnesses[i] = p.Evaluate(g)
        }

		totalFitness := 0.0

		for _, f := range p.fitnesses {
			if (f != math.Inf(-1) && f != math.Inf(+1) && !math.IsNaN(f)){
				totalFitness += f
			}
		}

		fmt.Printf("\t\tTotal fitness: %0.2f, ", totalFitness)
		fmt.Printf("\t\t\tAverage fitness: %0.2f\n", totalFitness / float64(len(p.fitnesses)))
        
		parentIndices := p.selector(p.Fitnesses, len(p.Genomes))
		offspring := make([]G, 0, len(p.Genomes))
        
        for i := 0; i < len(parentIndices)-1; i += 2 {
            idx1, idx2 := parentIndices[i], parentIndices[i+1]

			c1, c2 := p.Genomes[idx1], p.Genomes[idx2]
            
			// apply crossover with probability p.crossoverRate
			if rand.Float64() < p.crossoverRate {
				c1, c2 = p.crossover(c1, c2)
			} else {
            if rand.Float64() < p.mutationRate {
                c1 = p.mutate(c1)
            }
            if rand.Float64() < p.mutationRate {
                c2 = p.mutate(c2)
            }
			}
            offspring = append(offspring, c1, c2)
        }
        
        if p.eliteCount > 0 {
            elite := p.getElite()
            offspring = offspring[:len(offspring)-p.eliteCount]
            offspring = append(offspring, elite...)
        }
        
		p.Genomes = offspring
    }
}

func (p *Population[G]) getElite() []G {
	indices := make([]int, len(p.Genomes))
    for i := range indices {
        indices[i] = i
    }
    
	for i := range indices {
        for j := i + 1; j < len(indices); j++ {
			if p.Fitnesses[indices[j]] > p.Fitnesses[indices[i]] {
                indices[i], indices[j] = indices[j], indices[i]
            }
        }
    }
    
    elite := make([]G, p.eliteCount)
    for i := 0; i < p.eliteCount; i++ {
		elite[i] = p.Genomes[indices[i]]
    }
    return elite
}

func (p *Population[G]) Best() (G, float64) {
	bestIdx := -1
	bestFit := -100000000.0

	// fmt.Println(len(p.Fitnesses))
    
	for i, f := range p.Fitnesses {
		fmt.Println("Fitness", i, ":", f)
		if math.IsInf(f, 0) {
			continue
		}
        if f > bestFit {
            bestIdx = i
            bestFit = f
        }
    }
	return p.Genomes[bestIdx], bestFit
}
