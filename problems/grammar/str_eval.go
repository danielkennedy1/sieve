package grammar

import (
	"fmt"
	"math"

	"github.com/danielkennedy1/sieve/genomes"
	"github.com/expr-lang/expr"
)

type Sample struct {
	Variables []float64
	Output    float64
}

func NewRMSE(samples []Sample, gr genomes.Grammar) func(s string) float64 {
	return func(exprStr string) float64 {
		varMap := genomes.BuildVarMapFromGrammar(gr)
		fmt.Println("VarMap:", varMap)

		program, err := expr.Compile(exprStr, expr.AllowUndefinedVariables())
		if err != nil {
			fmt.Println("Failed to compile")
			return math.Inf(-1)
		}

		total := 0.0
		env := map[string]interface{}{}

		for _, s := range samples {
			for name, idx := range varMap {
				env[name] = s.Variables[idx]
			}

			out, err := expr.Run(program, env)
			if err != nil {
				fmt.Println("Failed to run")
				return math.Inf(-1)
			}

			diff := out.(float64) - s.Output
			total += diff * diff
		}

		return -math.Sqrt(total / float64(len(samples)))
	}
}
