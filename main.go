package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/danielkennedy1/sieve/benchmark"
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
		runGeneticAlgorithm()
		return

	case *runCompare:
		benchmark.RunComparison()
		return
	}

	fmt.Println("No action specified. Use -ga to run genetic algorithm, -chart to generate charts, or -compare to run comparison.")
}
