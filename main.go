package main

import (
    "fmt"
    
    "github.com/danielkennedy1/sieve.git/ea"
    "github.com/danielkennedy1/sieve.git/genomes"
    "github.com/danielkennedy1/sieve.git/problems/bitstring"
)

func main() {
    const genomeSize = 50
    
    pop := ea.NewPopulation(
        100,
        0.05,
        2,
        func() genomes.BitString { return genomes.NewBitString(genomeSize) },
        bitstring.OneMaxFitness,
        genomes.SinglePointCrossover,
        genomes.Mutate,
        ea.Tournament(3),
    )
    
    pop.Evolve(100)
    
    _, fitness := pop.Best()
    fmt.Printf("Best fitness: %.0f/%d\n", fitness, genomeSize)
}
