package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/danielkennedy1/sieve/benchmark"
)

var (
	runGA      = flag.Bool("ga", false, "Run genetic algorithm")
	makeChart  = flag.Bool("chart", false, "Generate charts from existing data")
	runCompare = flag.Bool("compare", true, "Run comparison between strategies")
	dataFile   = flag.String("data", "market_history.json", "Path to market history JSON file")
	outputDir  = flag.String("output", "charts", "Directory for chart output")
)

func main() {
	flag.Parse()

	if *makeChart {
		if err := benchmark.GenerateCharts(*dataFile, *outputDir); err != nil {
			fmt.Printf("Error generating charts: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *runGA {
		runGeneticAlgorithm()
		return
	}

	if *runCompare {
		benchmark.RunComparison()
		return
	}

}
