# Sieve
LK315 MSc Immersive Software Engineering

CS6425 Software Meets World - Evolutionary Algorithims

22340017 Daniel Kennedy

22343288 Dominick Stephens

Instructor: Dr. James Patten

## Overview

- Built in Golang
- Evolutionary algorithm framework 
- GA, GP and GE genomes
- Arbitrary problem space / fitness capabilities
- Primary focus: Simulated Market for GE of trading strategies

## Quick Start
Run the genetic algorithm to evolve trading strategies
```bash
go run main.go -ga
```

Generate charts from existing market data
```bash
go run main.go -chart -data market_history.json -output charts
```

Compare evolved strategy against baseline strategies
```bash
go run main.go -compare
```

## Architecture
```bash
ea/					    # Core evolutionary algorithm
├── population.go		# Population management, parallel evaluation, evolution loop
└── selection.go		# Tournament and roulette selection

genomes/				# Genome representations
├── grammar.go			# Grammar-based genotypes (main approach)
├── expression_tree.go  # Expression tree genotypes (legacy, still functional)
└── bitstring.go		# Simple bitstring genotypes

problems/				# Problem domains and fitness functions
├── grammar/
│   ├── market.go		# Market simulation with trading agents
│   ├── str_eval.go		# Symbolic regression via grammar
│   ├── parser.go		# BNF parser
│   └── indicators.go   # Technical indicators (RSI, SMA, ATR)
├── expression_tree/    # Symbolic regression (tree-based)
└── bitstring/			# Simple problems (OneMax)

config/					# Configuration management
└── *.toml				# TOML configs for different experiments

benchmark/				# Visualization and analysis
├── chart.go			# Generate HTML charts with go-echarts
└── comparison.go       # Compare strategies against baselines

data/					# Grammar definitions
├── market.bnf			# Full trading strategy grammar
└── sensible_market.bnf # Constrained grammar (better for evolution)
```


## Grammatical Evolution 

### Evolutionary behaviour
- Grammar: expr-lang compliant CFG, externally defined in BNF
- Genotype: Array of uint8 values (codons)
- Mapping: Codons select productions from grammar rules
- Phenotype: Valid program AST (e.g., trading strategy)
- Evaluation: Run strategy in market simulation
- Selection: Tournament selection based on fitness
- Variation: Single-point crossover + per-gene mutation

### Graceful wrapping
When genes run out, mapping wraps but picks least-recursive productions.

### Market Simulation
- Run N simulations
- Each simulation has M rounds
- Agents generate orders using their evolved strategies
- Price updates based on supply/demand + fundamental pull
- Track portfolio value, calculate returns, Sharpe ratio
- Fitness = average active return across simulations

Agents have access to market state: $PRICE, $RSI, $FUNDAMENTAL, $HOLDINGS, etc.
Noise traders provide liquidity and prevent everyone just holding.

### Configuration
Edit config/market.toml
```toml
generations = 50
max_genes = 200
bnf_file_path = "data/market.bnf"

[market]
initial_funds = 1500.0
initial_price = 100.0
initial_holdings = 15
rounds_per_generation = 100
noise_orders_per_round = 800
sims_per_generation = 30
fundamental_value_changes_per_simulation = 5
demand_push_coefficient = 0.2
fundamental_pull_coefficient = 0.01
rsi_period = 14

[population]
size = 500
mutation_rate = 0.05
crossover_rate = 0.6
gene_length = 100
tournament_size = 7
elite_count = 50
```


### Key parameters:
- noise_orders_per_round: More noise = more realistic but slower convergence
- sims_per_generation: More sims = more robust fitness but slower
- demand_push_coefficient: How much supply/demand affects price
- fundamental_pull_coefficient: How strongly price reverts to fundamental value

### Grammar Definition
Example from data/sensible_market.bnf:
```bnf
<strategy> ::= <condition> ? ( " SELL <natural> " ) : ( <condition> ? ( " BUY <natural> " ) : ( " HOLD " ) )
<condition> ::= ( <condition> && <condition> ) | ( <price_expr> <comp> <price_expr> )
<price_expr> ::= $PRICE | $FUNDAMENTAL | ( $PRICE <op> $PRICE )
<comp> ::= >
<natural> ::= 1..10
```

### BNF Extension
The 1..10 syntax expands to individual productions for each integer in range.
Variables starting with $ get replaced with actual values during evaluation (via expr-lang/expr).

### Parallel Evaluation
Population evaluation parallelizes across N workers (defaults to min(popSize, 8)). Each worker:

- Takes genotypes from job channel
- Evaluates fitness
- Optionally caches results (keyed by phenotype string)

Market simulation parallelizes across simulations - each sim runs independently, results get averaged.

### Benchmarking & Comparison
The comparison mode runs your best evolved strategy against baselines:
```bash
go run main.go -compare
```

Baselines include:

- Buy & Hold
- Simple threshold strategy
- Random trading
- Your best GA strategy (set in config)

### Charts
After running with -ga, generate visualizations:

```bash
go run main.go -chart
```

Creates charts/dashboard.html with:

- Price evolution over time
- Trading volume
- Fitness progression (best/avg/worst)
- Order flow (buy vs sell pressure)

Charts use go-echarts, render as interactive HTML.


## Extending
### New Problem Domain

- Define grammar in data/your_problem.bnf
- Create fitness function: func(genomes.Genotype) float64
- Wire it up in main.go:

```go
fitness := yourpkg.NewYourFitness(grammar, params)
pop := ea.NewPopulation(
    size, mutRate, crossRate, eliteCount,
    genomes.NewCreateGenotype(geneLength, rng),
    fitness,
    genomes.NewCrossoverGenotype(rng),
    genomes.NewMutateGenotype(rng, mutRate),
    ea.Tournament(tournamentSize),
    toKeyFunc,
    useCache,
)
```

### Custom Selection
Implement `func(fitnesses []float64, n int) []int` - takes fitness values, returns indices of selected parents.
See ea.Tournament() and ea.Roulette() for examples.

## Expression Trees
The framework started with expression trees before pivoting to grammars.
Expression tree code's still there and functional (`genomes/expression_tree.go`, `problems/expression_tree/`).

## Performance Notes

Caching helps when fitness is expensive and populations converge (set `cache_boolean = true`)
More noise traders = more realistic but also slower
Increase workers if you've got cores to spare (edit `numWorkers` in population.go)

## Testing
```bash
go test ./...
```

## Dependencies

- expr-lang/expr - evaluating strategy expressions
- spf13/viper - config management
- go-echarts - charting
- buger/goterm - terminal visualization during evolution
