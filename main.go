package main

import (
	"bufio"
	"fmt"
	"math/rand/v2"
	"os"
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
	g := grammar.Parse(*s)
	g.BuildRuleMap()

	targetExpressionString := config.TargetExpressionString
	numSamplesToGenerate := config.NumSamplesToGenerate
	initialVariables := make([]float64, config.NumVars)

	for i := 0; i < config.NumVars; i++ {
		initialVariables[i] = 0.0
	}

	samples, err := grammar.GenerateSamples(targetExpressionString, numSamplesToGenerate, initialVariables, g)

	if err != nil {
		fmt.Printf("Error generating samples: %v\n", err)
		return
	}

	// Use config variables for Population setup, accessing the nested fields
	population := ea.NewPopulation(
		config.Population.Size,
		config.Population.MutationRate,
		config.Population.CrossoverRate,
		config.Population.EliteCount,
		genomes.NewCreateGenotype(config.Population.GeneLength, r),
		grammar.NewRMSE(samples, g, config.ParsiomonyPenalty, config.MaxGenes),
		genomes.NewCrossoverGenotype(r),
		genomes.NewMutateGenotype(r, config.Population.MutationRate),
		ea.Tournament(config.Population.TournamentSize),
	)

	start := time.Now()

	// Use config variable for evolution generations
	population.Evolve(config.Generations)

	elapsed := time.Since(start)

	best, fitness := population.Best()
	fmt.Printf("Best fitness: %.2f\n", fitness)
	fmt.Println("Best: ", best.MapToGrammar(g, 100).String())
	fmt.Printf("Elapsed time: %s\n", elapsed)

}
