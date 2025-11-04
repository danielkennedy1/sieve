package bitstring

import "github.com/danielkennedy1/sieve.git/genomes"

func OneMaxFitness(bs genomes.BitString) float64 {
    count := 0
    for _, bit := range bs {
        if bit {
            count++
        }
    }
    return float64(count)
}
