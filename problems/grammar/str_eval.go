package grammar

import (
	"math"

	"github.com/danielkennedy1/sieve/genomes"
	"github.com/expr-lang/expr"
)

type Sample struct {
	Variables []float64
	Output    float64
}

func NewRMSE(samples []Sample, gr genomes.Grammar) func(g genomes.Genotype) float64 {
	// Define the penalty coefficient (lambda)
	// This value controls how harshly the EA is punished for complexity.
	// Start with a small value like 0.001 and adjust if needed.
	const parsimonyPenalty = 0.001

	return func(g genomes.Genotype) float64 {
		varMap := genomes.BuildVarMapFromGrammar(gr)
		exprStr := g.MapToGrammar(gr, 100).String()

		// 1. Calculate Length Penalty (Complexity Component)
		// We use the number of characters in the resulting expression string as a proxy for complexity.
		lengthPenalty := float64(len(exprStr)) * parsimonyPenalty

		program, err := expr.Compile(exprStr, expr.AllowUndefinedVariables())
		if err != nil {
			// fmt.Println("Failed to compile")
			return math.Inf(-1)
		}

		total := 0.0
		env := map[string]interface{}{}

		for _, s := range samples {
			for name, idx := range varMap {
				env[name] = s.Variables[idx]
			}

			out, err := expr.Run(program, env)
			if math.IsNaN(out.(float64)) {
				return math.Inf(-1)
			}
			if err != nil {
				// fmt.Println("Failed to run")
				return math.Inf(-1)
			}

			diff := out.(float64) - s.Output
			total += diff * diff
		}

		// 2. Calculate RMSE (Accuracy Component)
		rmse := math.Sqrt(total / float64(len(samples)))

		// 3. Combine Fitness:
		// Fitness = -RMSE - LengthPenalty
		// Since we are maximizing fitness, a lower RMSE is better (more positive),
		// and a shorter length is better (less penalty subtracted).
		return -rmse - lengthPenalty
	}
}
