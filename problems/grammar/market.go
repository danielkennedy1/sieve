package grammar

import (
	"math"

	"github.com/danielkennedy1/sieve/genomes"
	"github.com/expr-lang/expr"
)

func NewTransactionFitness(gr genomes.Grammar, prices []float64, initialFunds float64) func(g genomes.Genotype) float64 {
	return func(g genomes.Genotype) float64 {
		var funds = initialFunds
		var holdings = 0

		exprStr := g.MapToGrammar(gr, 7).String()
		program, err := expr.Compile(exprStr, expr.Env(map[string]interface{}{
			"$PRICE": 0.0,
			"$HOLDINGS": holdings,
		}))
		if err != nil {
			return math.Inf(-1)
		}

		// Run with different prices
		for _, p := range prices {
			out, err := expr.Run(program, map[string]interface{}{
				"$PRICE": p,
				"$HOLDINGS": holdings,
			})

			if err != nil {
				return math.Inf(-1)
			}
			// TODO: will have to manage proportions
			switch out {
			case "BUY":
				maxToBuy := int(funds / p)
				value := float64(maxToBuy) * p
				funds -= value
				holdings += maxToBuy
			case "SELL":
				value := float64(holdings) * p
				funds += value
				holdings = 0
			case "HOLD":
			default:
			}
		}

		return funds + prices[len(prices)-1]*float64(holdings)
	}
}
