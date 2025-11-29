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
	const parsimonyPenalty = 0.001

	return func(g genomes.Genotype) float64 {
		varMap := genomes.BuildVarMapFromGrammar(gr)
		exprStr := g.MapToGrammar(gr, 100).String()

		lengthPenalty := float64(len(exprStr)) * parsimonyPenalty

		program, err := expr.Compile(exprStr, expr.AllowUndefinedVariables())
		if err != nil {
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
				return math.Inf(-1)
			}

			diff := out.(float64) - s.Output
			total += diff * diff
		}

		rmse := math.Sqrt(total / float64(len(samples)))

		return -rmse - lengthPenalty
	}
}
