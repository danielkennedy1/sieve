package main

import (
	"bufio"
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"

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
	gr := grammar.Parse(*s)
	gr.BuildRuleMap()

	f, err = os.Open("data/dominoes.txt")
	if err != nil {
		fmt.Println("Prices file not found")
		os.Exit(1)
	}

	var prices []float64

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		val, err := strconv.ParseFloat(scanner.Text(), 64)
		if err != nil {
			continue
		}
		prices = append(prices, val)
	}
	fmt.Println("Prices ", prices)

	transaction_fitness := grammar.NewTransactionFitness(gr)

	sample_maker := genomes.NewCreateGenotype(config.Population.GeneLength, r)

	for range 1 {
		sample := sample_maker()
		fmt.Println("----")
		fmt.Println(transaction_fitness(sample))
	}

}
