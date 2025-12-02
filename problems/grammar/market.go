package grammar

import "github.com/danielkennedy1/sieve/genomes"

func NewTransactionFitness() func(g genomes.Genotype) float64 {
	return func(g genomes.Genotype) float64 {
		return 1.0
	}
}
