package grammar

import (
	"errors"
	"fmt"
	"math"

	"github.com/danielkennedy1/sieve/genomes"
	"github.com/expr-lang/expr"
)

type Sample struct {
	Variables []float64
	Output    float64
}

func GenerateSamples(exprStr string, numSamples int, variables []float64, gr genomes.Grammar) ([]Sample, error) {
	if numSamples <= 0 {
		return nil, errors.New("numSamples must be greater than zero")
	}
	if len(variables) == 0 {
		return nil, errors.New("variables slice cannot be empty")
	}

	program, err := expr.Compile(exprStr, expr.AllowUndefinedVariables())
	if err != nil {
		return nil, fmt.Errorf("failed to compile expression string: %w", err)
	}

	startRange := -10.0
	endRange := 10.0

	step := (endRange - startRange) / float64(numSamples-1)
	if numSamples == 1 {
		step = 0
	}

	varMap := genomes.BuildVarMapFromGrammar(gr)
	env := map[string]interface{}{}
	samples := make([]Sample, 0, numSamples)

	for i := 0; i < numSamples; i++ {
		currentInput := startRange + float64(i)*step

		variables[0] = currentInput

		for name, idx := range varMap {
			env[name] = variables[idx]
		}

		out, err := expr.Run(program, env)
		if err != nil {
			return nil, fmt.Errorf("runtime error during evaluation: %w", err)
		}

		result, ok := out.(float64)
		if !ok {
			return nil, errors.New("expression result was not a float64")
		}
		if math.IsNaN(result) || math.IsInf(result, 0) {
			return nil, fmt.Errorf("expression resulted in NaN or Infinity at input x0=%.2f", currentInput)
		}

		inputVarsClone := make([]float64, len(variables))
		copy(inputVarsClone, variables)

		samples = append(
			samples,
			Sample{
				Variables: inputVarsClone,
				Output:    result,
			},
		)
	}

	return samples, nil
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
