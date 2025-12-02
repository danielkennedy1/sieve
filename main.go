package main

import (
	"bufio"
	"fmt"
	"math/rand/v2"
	"os"

	"github.com/danielkennedy1/sieve/config"
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

	initialVariables := make([]float64, config.NumVars)

	for i := 0; i < config.NumVars; i++ {
		initialVariables[i] = 0.0
	}

	if err != nil {
		fmt.Printf("Error generating samples: %v\n", err)
		return
	}

	transaction_fitness := grammar.NewTransactionFitness()

	for range 100 {
		sample_maker := genomes.NewCreateGenotype(config.Population.GeneLength, r)
		sample := sample_maker()
		fmt.Println("----")
		fmt.Println(sample.MapToGrammar(g, 100).String())
		fmt.Println(transaction_fitness(sample))
	}

}
