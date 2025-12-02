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
	config, err := config.LoadConfig()
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

	populationSize := 1

	initialFunds := 30_000.0

	transactionFitness := grammar.NewTransactionFitness(gr, prices, initialFunds)

	sampleMaker := genomes.NewCreateGenotype(config.Population.GeneLength, r)

	for range populationSize {
		sample := sampleMaker()
		fmt.Println("----")

		fmt.Println(sample.MapToGrammar(gr, 10).String())
		
		fmt.Println(transactionFitness(sample))
	}

	population := ea.NewPopulation(
		config.Population.Size,
		config.Population.MutationRate,
		config.Population.CrossoverRate,
		config.Population.EliteCount,
		genomes.NewCreateGenotype(config.Population.GeneLength, r),
		grammar.NewTransactionFitness(gr, prices, initialFunds),
		genomes.NewCrossoverGenotype(r),
		genomes.NewMutateGenotype(r, config.Population.MutationRate),
		ea.Tournament(config.Population.TournamentSize),
		func(g genomes.Genotype) string {
			return string(g.Genes)
		},
		)

		start := time.Now()

		population.Evolve(config.Generations)

		elapsed := time.Since(start)

		best, fitness := population.Best()
		fmt.Printf("Best fitness: %.2f\n", fitness)
		fmt.Println("Best: ", best.MapToGrammar(gr, 100).String())
		fmt.Printf("Elapsed time: %s\n", elapsed)
}
