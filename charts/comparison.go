package charts

import (
	"bufio"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"time"

	"github.com/danielkennedy1/sieve/config"
	"github.com/danielkennedy1/sieve/genomes"
	"github.com/danielkennedy1/sieve/problems/grammar"
)

type AgentStats struct {
	Name        string
	Count       int
	TotalCash   float64
	TotalStock  int
	TotalWealth float64
	TotalFit    float64
}

func RunComparison() {
	cfg, err := config.LoadConfig("market")
	if err != nil {
		fmt.Printf("Fatal error loading configuration: %v\n", err)
		os.Exit(1)
	}

	f, err := os.Open(cfg.BNFFilePath)
	if err != nil {
		fmt.Printf("File not found: %s\n", cfg.BNFFilePath)
		os.Exit(1)
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	gr := grammar.Parse(*s)
	gr.BuildRuleMap()

	// 3. Initialize Simulator
	simulator := grammar.NewMarketSimulator(
		gr,
		cfg.Market.InitialPrice,
		cfg.Market.InitialFunds,
		cfg.Market.InitialHoldings,
		cfg.Market.RoundsPerGeneration,
		cfg.MaxGenes,
	)

	fmt.Println("\n=== Starting Comparison Benchmark ===")
	fmt.Printf("Market Rounds: %d\n", cfg.Market.RoundsPerGeneration)
	fmt.Printf("Initial Funds: $%.2f\n", cfg.Market.InitialFunds)

	// 4. Define Agent Types
	bestGAStrategy := `($PRICE <= 100) ? "BUY 1.0" : ( ($PRICE >= 100) ? "SELL 1.0" : "HOLD" )`

	agentTypes := []struct {
		Name     string
		Strategy string
	}{
		{Name: "Best GA", Strategy: bestGAStrategy},
		{Name: "Buy & Hold", Strategy: `(true) ? "BUY 0.5" : "SELL 0.0"`},
		{Name: "Simple", Strategy: `($PRICE <= 110) ? "BUY 0.5" : "SELL 0.5"`},
		{Name: "Random", Strategy: `($RANDOM >= 0.5) ? "BUY 1" : "SELL 1"`},
	}

	// Create a shared random source for the simulation
	r := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 100))

	// 5. Instantiate Population (e.g., 25 of each type = 100 agents)
	clonesPerType := 25
	var genotypes []genomes.Genotype

	// We map Genotype ID (index) back to Agent Name for reporting later
	idToName := make(map[int]string)

	fmt.Printf("Populating market with %d agents (%d types x %d clones)...\n",
		len(agentTypes)*clonesPerType, len(agentTypes), clonesPerType)

	idx := 0
	for _, agent := range agentTypes {
		for i := 0; i < clonesPerType; i++ {
			g := genomes.NewCreateGenotype(1, r, make(map[string]any))()

			g.Attributes["HardcodedStrategy"] = agent.Strategy
			g.Attributes["AgentType"] = agent.Name

			g.Attributes["cash"] = cfg.Market.InitialFunds
			g.Attributes["holdings"] = cfg.Market.InitialHoldings

			genotypes = append(genotypes, g)
			idToName[idx] = agent.Name
			idx++
		}
	}

	start := time.Now()
	simulator.BeforeGeneration(genotypes)
	elapsed := time.Since(start)

	stats := make(map[string]*AgentStats)
	for _, a := range agentTypes {
		stats[a.Name] = &AgentStats{Name: a.Name}
	}

	fitnessFunc := simulator.NewMarketFitness()

	for _, g := range genotypes {
		typeName := g.Attributes["AgentType"].(string)

		cash := 0.0
		if c, ok := g.Attributes["cash"].(float64); ok {
			cash = c
		}

		holdings := 0
		if h, ok := g.Attributes["holdings"].(int); ok {
			holdings = h
		}

		wealth := cash + float64(holdings)*simulator.Market.CurrentPrice
		fitness := fitnessFunc(g)

		s := stats[typeName]
		s.Count++
		s.TotalCash += cash
		s.TotalStock += holdings
		s.TotalWealth += wealth

		if !math.IsNaN(fitness) && !math.IsInf(fitness, 0) {
			s.TotalFit += fitness
		}
	}

	fmt.Printf("\n=== Results (Sim Time: %s) ===\n", elapsed)
	fmt.Printf("Final Market Price: $%.2f (Change: %.2f%%)\n",
		simulator.Market.CurrentPrice,
		((simulator.Market.CurrentPrice-cfg.Market.InitialPrice)/cfg.Market.InitialPrice)*100,
	)
	fmt.Println("--------------------------------------------------------------------------------------")
	fmt.Printf("%-15s | %-6s | %-14s | %-12s | %-14s | %-8s\n",
		"Agent Type", "Count", "Avg Wealth", "Avg Cash", "Avg Holdings", "Avg Fit")
	fmt.Println("--------------------------------------------------------------------------------------")

	for _, a := range agentTypes {
		s := stats[a.Name]
		if s.Count == 0 {
			continue
		}

		avgWealth := s.TotalWealth / float64(s.Count)
		avgCash := s.TotalCash / float64(s.Count)
		avgHoldings := float64(s.TotalStock) / float64(s.Count)
		avgFit := s.TotalFit / float64(s.Count)

		fmt.Printf("%-15s | %-6d | $%-13.2f | $%-11.2f | %-14.1f | %.4f\n",
			a.Name, s.Count, avgWealth, avgCash, avgHoldings, avgFit)
	}
	fmt.Println("--------------------------------------------------------------------------------------")

	simulator.History.ExportJSON("comparison_history.json")
	fmt.Println("\nHistory exported to comparison_history.json")
}
