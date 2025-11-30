package ea

import (
	"fmt"
	"math"
	"math/rand/v2"
	"sync"
	"time"
)

type Population[G any] struct {
	genomes       []G
	fitnesses     []float64
	evaluate      func(G) float64
	crossover     func(G, G) (G, G)
	mutate        func(G) G
	selector      func([]float64, int) []int
	crossoverRate float64
	mutationRate  float64
	eliteCount    int
	numWorkers    int

	cache      map[string]float64
	cacheMutex sync.RWMutex
	toKey      func(G) string
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
	toKey func(G) string,
) *Population[G] {
	genomes := make([]G, size)
	for i := range genomes {
		genomes[i] = create()
	}
	numWorkers := min(size, 8)
	return &Population[G]{
		genomes:       genomes,
		fitnesses:     make([]float64, size),
		evaluate:      evaluate,
		crossover:     crossover,
		mutate:        mutate,
		selector:      selector,
		crossoverRate: crossoverRate,
		mutationRate:  mutationRate,
		eliteCount:    eliteCount,
		numWorkers:    numWorkers,
		cache:         make(map[string]float64),
		toKey:         toKey,
	}
}

func (p *Population[G]) evaluateAll() {
	jobs := make(chan int, len(p.genomes))
	var wg sync.WaitGroup

	for w := 0; w < p.numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				key := p.toKey(p.genomes[idx])

				// Check cache first (read lock)
				p.cacheMutex.RLock()
				fitness, exists := p.cache[key]
				p.cacheMutex.RUnlock()

				if exists {
					p.fitnesses[idx] = fitness
				} else {
					// Evaluate and store in cache (write lock)
					fitness = p.evaluate(p.genomes[idx])
					p.fitnesses[idx] = fitness

					p.cacheMutex.Lock()
					p.cache[key] = fitness
					p.cacheMutex.Unlock()
				}
			}
		}()
	}

	for i := range p.genomes {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
}

func (p *Population[G]) Evolve(generations int) {
	timeList := make([]time.Duration, 0, generations)

	for generation := 0; generation < generations; generation++ {
		start := time.Now()
		fmt.Printf("Generation %d\n", generation)

		p.evaluateAll()

		parentIndices := p.selector(p.fitnesses, len(p.genomes))

		offspring := make([]G, len(p.genomes))

		type job struct {
			idx  int
			idx1 int
			idx2 int
		}

		jobs := make(chan job, len(parentIndices)/2)
		var wg sync.WaitGroup

		for w := 0; w < p.numWorkers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				localRng := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), uint64(w)))

				for j := range jobs {
					c1, c2 := p.crossover(p.genomes[j.idx1], p.genomes[j.idx2])

					if localRng.Float64() < p.mutationRate {
						c1 = p.mutate(c1)
					}
					if localRng.Float64() < p.mutationRate {
						c2 = p.mutate(c2)
					}

					offspring[j.idx] = c1
					offspring[j.idx+1] = c2
				}
			}()
		}

		for i := 0; i < len(parentIndices)-1; i += 2 {
			jobs <- job{
				idx:  i,
				idx1: parentIndices[i],
				idx2: parentIndices[i+1],
			}
		}

		close(jobs)
		wg.Wait()

		totalFitness := 0.0
		bestFitness := -math.MaxFloat64

		for _, f := range p.fitnesses {
			if !math.IsInf(f, 0) && !math.IsNaN(f) {
				totalFitness += f
				if f > bestFitness {
					bestFitness = f
				}
			}
		}
		fmt.Printf("\t\tBest fitness: %0.2f, ", bestFitness)
		fmt.Printf("\t\tTotal fitness: %0.2f, ", totalFitness)
		fmt.Printf("\t\t\tAverage fitness: %0.2f\n", totalFitness/float64(len(p.fitnesses)))

		if p.eliteCount > 0 {
			elite := p.getElite()
			offspring = offspring[:len(offspring)-p.eliteCount]
			offspring = append(offspring, elite...)
		}

		p.genomes = offspring

		elapsed := time.Since(start)
		timeList = append(timeList, elapsed)
	}

	p.evaluateAll()

	var sum time.Duration
	for _, t := range timeList {
		sum += t
	}
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
	bestIdx := -1
	bestFit := -100000000.0

	// fmt.Println(len(p.fitnesses))

	for i, f := range p.fitnesses {
		fmt.Println("Fitness", i, ":", f)
		if math.IsInf(f, 0) {
			continue
		}
		if f > bestFit {
			bestIdx = i
			bestFit = f
		}
	}
	return p.genomes[bestIdx], bestFit
}
