package main

import (
	"flag"
	"fmt"
	"os"
	"bufio"
	"math/rand/v2"
	"slices"
	"time"

	"runtime/pprof"

	"github.com/danielkennedy1/sieve/benchmark"
	"github.com/danielkennedy1/sieve/config"
	"github.com/danielkennedy1/sieve/ea"
	"github.com/danielkennedy1/sieve/genomes"
	"github.com/danielkennedy1/sieve/problems/grammar"
)

func main() {

	runGA := flag.Bool("ga", false, "Run genetic algorithm")
	makeChart := flag.Bool("chart", false, "Generate charts from existing data")
	runCompare := flag.Bool("compare", false, "Run comparison between strategies")
	dataFile := flag.String("data", "market_history.json", "Path to market history JSON file")
	outputDir := flag.String("output", "charts", "Directory for chart output")

	flag.Parse()

	switch {
	case *makeChart:
		if err := benchmark.GenerateCharts(*dataFile, *outputDir); err != nil {
			fmt.Printf("Error generating charts: %v\n", err)
			os.Exit(1)
		}
		return

	case *runGA:
		runMarketGE()
		return

	case *runCompare:
		benchmark.RunComparison()
		return
	}

	fmt.Println("No action specified. Use -ga to run genetic algorithm, -chart to generate charts, or -compare to run comparison.")
}

func runMarketGE() {
	config, err := config.LoadConfig("market")

	f, _ := os.Create("cpu.prof")
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	if err != nil {
		fmt.Printf("Fatal error loading configuration: %v\n", err)
		os.Exit(1)
	}

	f, err = os.Open(config.BNFFilePath)
	if err != nil {
		fmt.Printf("File not found: %s\n", config.BNFFilePath)
		os.Exit(1)
	}
	defer f.Close()

	r := rand.New(rand.NewPCG(0, 0))
	s := bufio.NewScanner(f)
	gr := grammar.Parse(*s)
	gr.BuildRuleMap()

	simulator := &grammar.MarketSimulator{
		Results: nil,
		Config: &grammar.MarketConfig{
			Grammar:                              gr,
			MaxGenes:                             config.MaxGenes,
			InitialPrice:                         config.Market.InitialPrice,
			InitialFunds:                         config.Market.InitialFunds,
			InitialHoldings:                      config.Market.InitialHoldings,
			RoundsPerSim:                         config.Market.RoundsPerGeneration,
			NoiseOrdersPerRound:                  config.Market.NoiseOrdersPerRound,
			SimsPerGeneration:                    config.Market.SimsPerGeneration,
			FundamentalValueChangesPerSimulation: config.Market.FundamentalValueChangesPerSimulation,
			DemandPushCoefficient:                config.Market.DemandPushCoefficient,
			FundamentalPullCoefficient:           config.Market.FundamentalPullCoefficient,
			RSIPeriod:                            config.Market.RSIPeriod,
			ATRPeriod:                            config.Market.ATRPeriod,
			SMAPeriod:                            config.Market.SMAPeriod,
		},
		History:    &grammar.MarketHistory{},
		Rng:        r,
		Generation: 0,
	}

	population := ea.NewPopulation(
		config.Population.Size,
		config.Population.MutationRate,
		config.Population.CrossoverRate,
		config.Population.EliteCount,
		genomes.NewCreateGenotype(config.Population.GeneLength, r),
		simulator.NewMarketFitness(),
		genomes.NewCrossoverGenotype(r),
		genomes.NewMutateGenotype(r, config.Population.MutationRate),
		ea.Tournament(config.Population.TournamentSize),
		func(g genomes.Genotype) string {
			return string(g.Genes)
		},
		config.Population.CacheBoolean,
	)

	population.BeforeEvaluate = simulator.BeforeGeneration
	population.AfterEvaluate = simulator.AfterGeneration

	start := time.Now()
	population.Evolve(config.Generations)
	elapsed := time.Since(start)

	best, fitness := population.Best()
	fmt.Printf("\n=== Results ===\n")
	fmt.Printf("Best fitness: $%.2f\n", fitness)
	fmt.Printf("Best strategy: %s\n", best.MapToGrammar(gr, 100).String())
	fmt.Printf("Elapsed time: %s\n", elapsed)

	err = simulator.History.ExportJSON("market_history.json")
	if err != nil {
		fmt.Printf("Error exporting history: %v\n", err)
	} else {
		fmt.Println("\nMarket history exported to market_history.json")
	}

	fmt.Printf("Total rounds: %d\n", len(simulator.History.Prices))
	fmt.Printf("Price range: $%.2f - $%.2f\n",
		slices.Min(simulator.History.Prices),
		slices.Max(simulator.History.Prices))

	totalVolume := 0

	for _, v := range simulator.History.Volumes {
		totalVolume += v
	}

	fmt.Printf("Total volume traded: %d\n", totalVolume)

	bestGen := grammar.FindBestGeneration(simulator.History.Generations)
	fmt.Printf("\nBest generation: %d (avg fitness: $%.2f)\n",
		bestGen.Generation, bestGen.AvgFitness)

	for _, gen := range simulator.History.Generations {
		if gen.Generation == bestGen.Generation {
			fmt.Printf("Final price: $%.2f\n", gen.FinalPrice)
			fmt.Printf("Total buy orders: %d\n", gen.BuyOrders)
			fmt.Printf("Total sell orders: %d\n", gen.SellOrders)
			break
		}
	}
}
