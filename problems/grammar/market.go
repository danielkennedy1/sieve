package grammar

import (
	"fmt"

	"github.com/danielkennedy1/sieve/genomes"
)

func NewTransactionFitness(gr genomes.Grammar) func(g genomes.Genotype) float64 {
	return func(g genomes.Genotype) float64 {
		fmt.Println(g.MapToGrammar(gr, 100))
		return 1.0
	}
}
