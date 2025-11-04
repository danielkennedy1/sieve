package main

import (
    "fmt"
    
    "github.com/danielkennedy1/sieve/ea"
    "github.com/danielkennedy1/sieve/genomes"
    "github.com/danielkennedy1/sieve/problems/bitstring"
)


func main() {
    const genomeSize = 200
    
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
    fmt.Printf("Best fitness: %.2f\n", fitness)
}
