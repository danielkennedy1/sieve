package genomes

import (
	"math/rand"
)

func createRandomTerminal(variables, constants *[]float64, numVars int, r *rand.Rand) Expression {
	if numVars == 0 || r.Float64() < 0.5 {
		return Primitive{Value: (*constants)[rand.Intn(len(*constants))]}
	} else {
		return Variable{
			Variables: variables,
			Index:     r.Intn(numVars),
		}
	}
}

func createRandomExpression(currentDepth, maxDepth int, variables, constants *[]float64, numVars int, r *rand.Rand) Expression {
	if currentDepth == maxDepth {
		return createRandomTerminal(variables, constants, numVars, r)
	}

	// TODO: Make this negative exponential with max depth
	probNonTerminal := 1.0 - (float64(currentDepth) / float64(maxDepth))

	if r.Float64() < probNonTerminal {
		return NonTerminal{
			Operator: Operator(r.Intn(int(numOperators))),
			Left:     createRandomExpression(currentDepth+1, maxDepth, variables, constants, numVars, r),
			Right:    createRandomExpression(currentDepth+1, maxDepth, variables, constants, numVars, r),
		}
	} else {
		return createRandomTerminal(variables, constants, numVars, r)
	}
}

func RandomFormula(maxDepth int, variables, constants *[]float64, numVars int, r *rand.Rand) Expression {
	if maxDepth <= 0 {
		return createRandomTerminal(variables, constants, numVars, r)
	}
	return createRandomExpression(0, maxDepth, variables, constants, numVars, r)
}
