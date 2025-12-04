package main

import (
	"bufio"
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
	"time"

	"github.com/danielkennedy1/sieve/config"
	"github.com/danielkennedy1/sieve/ea"
	"github.com/danielkennedy1/sieve/genomes"
	"github.com/danielkennedy1/sieve/problems/grammar"
)

func main() {
	config, err := config.LoadConfig("nyse")
	if err != nil {
		fmt.Printf("Fatal error loading configuration: %v\n", err)
		os.Exit(1)
	}

	f, err := os.Open(config.BNFFilePath)
	if err != nil {
		fmt.Printf("File not found: %s\n", config.BNFFilePath)
		os.Exit(1)
	}
	defer f.Close()

	r := rand.New(rand.NewPCG(0, 0))
	s := bufio.NewScanner(f)
	gr := grammar.Parse(*s)
	gr.BuildRuleMap()

	f, err = os.Open("data/dominoes.txt")
	if err != nil {
		fmt.Println("Prices file not found")
		os.Exit(1)
	}
	defer f.Close()

	var prices []float64
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		val, err := strconv.ParseFloat(scanner.Text(), 64)
		if err != nil {
			continue
		}
		prices = append(prices, val)
	}

	initialFunds := 1500.0
	initialPrice := 100.0
	initialHoldings := 100
	roundsPerGen := 100

	simulator := grammar.NewMarketSimulator(
		gr,
		initialPrice,
		initialFunds,
		initialHoldings,
		roundsPerGen,
	)

	attributes := make(map[string]any)
	attributes["cash"] = initialFunds
	attributes["holdings"] = 100

	population := ea.NewPopulation(
		config.Population.Size,
		config.Population.MutationRate,
		config.Population.CrossoverRate,
		config.Population.EliteCount,
		genomes.NewCreateGenotype(config.Population.GeneLength, r, attributes),
		simulator.NewMarketFitness(),
		genomes.NewCrossoverGenotype(r),
		genomes.NewMutateGenotype(r, config.Population.MutationRate),
		ea.Tournament(config.Population.TournamentSize),
		func(g genomes.Genotype) string {
			return string(g.Genes)
		},
	)

	population.BeforeEvaluate = simulator.BeforeGeneration
	population.AfterEvaluate = simulator.AfterGeneration
	population.AfterSelection = simulator.ResetOffspring

	start := time.Now()
	population.Evolve(config.Generations)
	elapsed := time.Since(start)

	best, fitness := population.Best()
	fmt.Printf("\n=== Results ===\n")
	fmt.Printf("Best fitness: $%.2f\n", fitness)
	fmt.Printf("Best strategy: %s\n", best.MapToGrammar(gr, 100).String())
	fmt.Printf("Final market price: $%.2f\n", simulator.Market.CurrentPrice)
	fmt.Printf("Price change: %.2f%%\n",
		(simulator.Market.CurrentPrice-simulator.Market.InitialPrice)/simulator.Market.InitialPrice*100)
	fmt.Printf("Elapsed time: %s\n", elapsed)

	err = simulator.History.ExportJSON("market_history.json")
	if err != nil {
		fmt.Printf("Error exporting history: %v\n", err)
	} else {
		fmt.Println("\nMarket history exported to market_history.json")
	}

	fmt.Printf("Total rounds: %d\n", len(simulator.History.Prices))
	fmt.Printf("Price range: $%.2f - $%.2f\n",
		grammar.MinPrice(simulator.History.Prices),
		grammar.MaxPrice(simulator.History.Prices))
	fmt.Printf("Total volume traded: %d\n", grammar.SumVolume(simulator.History.Volumes))

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
