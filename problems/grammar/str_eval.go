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
	varMap := genomes.BuildVarMapFromGrammar(gr)

	namesByIdx := make([]string, len(varMap))
	for name, idx := range varMap {
		namesByIdx[idx] = name
	}

	n := float64(len(samples))

	return func(g genomes.Genotype) float64 {
		// Map genotype to expression
		exprStr := g.MapToGrammar(gr, 1000).String()

		program, err := expr.Compile(exprStr, expr.AllowUndefinedVariables())
		if err != nil {
			// fmt.Println("Failed to compile")
			return math.Inf(-1)
		}

		// Per-call env: goroutine-local
		env := make(map[string]interface{}, len(namesByIdx))

		total := 0.0

		for _, s := range samples {
			for i, name := range namesByIdx {
				env[name] = s.Variables[i]
			}

			out, err := expr.Run(program, env)
			if err != nil {
				return math.Inf(-1)
			}

			v, ok := out.(float64)
			if !ok {
				return math.Inf(-1)
			}

			diff := v - s.Output
			total += diff * diff
		}
		return -math.Sqrt(total / n)
	}
}
