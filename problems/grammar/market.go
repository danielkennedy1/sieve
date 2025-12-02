package grammar

import (
	"fmt"
	"strings"
	"strconv"
	"math"

	"github.com/expr-lang/expr"
	"github.com/danielkennedy1/sieve/genomes"
)

func NewTransactionFitness(gr genomes.Grammar, prices []float64, initialFunds float64) func(g genomes.Genotype) float64 {
	return func(g genomes.Genotype) float64 {

		var funds = initialFunds
		var holdings = 0

		for _, p := range prices {
			exprStr := g.MapToGrammar(gr, 100).String()
			strings.ReplaceAll(exprStr, "$PRICE", strconv.FormatFloat(p, 'f', -1, 64))
			program, err := expr.Compile(exprStr, expr.AllowUndefinedVariables())

			if err != nil {
				fmt.Println("Error compiling: ", exprStr)
				fmt.Println(err)
				return math.Inf(-1)
			}

			out, err := expr.Run(program, nil)

			if err != nil {
				fmt.Println("Error running: ", exprStr)
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

		return funds + prices[len(prices) - 1] * float64(holdings)
	}
}
