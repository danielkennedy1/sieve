package grammar

import (
	"encoding/json"
	"fmt"
	"math"
	"os"

	"github.com/danielkennedy1/sieve/genomes"
	"github.com/expr-lang/expr"
)

type MarketState struct {
	CurrentPrice  float64
	InitialPrice  float64
	PriceHistory  []float64
	VolumeHistory []int
}

type Order struct {
	GenotypeID int
	Action     string
	Quantity   int
}

type MarketHistory struct {
	Timestamps  []int
	Prices      []float64
	Volumes     []int
	Generations []GenerationSnapshot
}

type GenerationSnapshot struct {
	Generation   int
	FinalPrice   float64
	BuyOrders    int
	SellOrders   int
	AvgFitness   float64
	BestFitness  float64
	WorstFitness float64
}

func NewMarketState(initialPrice float64) *MarketState {
	return &MarketState{
		CurrentPrice:  initialPrice,
		InitialPrice:  initialPrice,
		PriceHistory:  []float64{initialPrice},
		VolumeHistory: []int{},
	}
}

func NewMarketHistory() *MarketHistory {
	return &MarketHistory{
		Timestamps:  []int{},
		Prices:      []float64{},
		Volumes:     []int{},
		Generations: []GenerationSnapshot{},
	}
}

type MarketSimulator struct {
	Market       *MarketState
	History      *MarketHistory
	Grammar      genomes.Grammar
	InitialFunds float64
	RoundsPerGen int
	generation   int
}

func NewMarketSimulator(grammar genomes.Grammar, initialPrice, initialFunds float64, roundsPerGen int) *MarketSimulator {
	return &MarketSimulator{
		Market:       NewMarketState(initialPrice),
		History:      NewMarketHistory(),
		Grammar:      grammar,
		InitialFunds: initialFunds,
		RoundsPerGen: roundsPerGen,
		generation:   0,
	}
}

func (ms *MarketSimulator) NewMarketFitness() func(g genomes.Genotype) float64 {
	return func(g genomes.Genotype) float64 {
		if g.Attributes == nil {
			g.Attributes = make(map[string]interface{})
			g.Attributes["cash"] = ms.InitialFunds
			g.Attributes["holdings"] = 0
		}

		funds := g.Attributes["cash"].(float64)
		holdings := g.Attributes["holdings"].(int)

		portfolioValue := funds + float64(holdings)*ms.Market.CurrentPrice

		return portfolioValue
	}
}

func (ms *MarketSimulator) BeforeGeneration(genotypes []genomes.Genotype) {
	totalBuyOrders := 0
	totalSellOrders := 0

	for round := 0; round < ms.RoundsPerGen; round++ {
		roundNumber := ms.generation*ms.RoundsPerGen + round
		var orders []Order

		for i, g := range genotypes {
			order := ms.generateOrder(&g, i)
			orders = append(orders, order)
		}

		buyOrders := 0
		sellOrders := 0
		for _, order := range orders {
			switch order.Action {
			case "BUY":
				buyOrders += order.Quantity
			case "SELL":
				sellOrders += order.Quantity
			}
		}

		totalBuyOrders += buyOrders
		totalSellOrders += sellOrders

		newPrice := ms.calculateNewPrice(buyOrders, sellOrders)

		for i, order := range orders {
			ms.executeOrder(&genotypes[i], order, newPrice)
		}

		ms.Market.CurrentPrice = newPrice
		ms.Market.PriceHistory = append(ms.Market.PriceHistory, newPrice)

		ms.History.Timestamps = append(ms.History.Timestamps, roundNumber)
		ms.History.Prices = append(ms.History.Prices, newPrice)
		ms.History.Volumes = append(ms.History.Volumes, buyOrders+sellOrders)
	}

	ms.History.Generations = append(ms.History.Generations, GenerationSnapshot{
		Generation: ms.generation,
		FinalPrice: ms.Market.CurrentPrice,
		BuyOrders:  totalBuyOrders,
		SellOrders: totalSellOrders,
	})
}

func (ms *MarketSimulator) AfterGeneration(fitnesses []float64) {
	totalFitness := 0.0
	bestFitness := -math.MaxFloat64
	worstFitness := math.MaxFloat64
	validCount := 0

	for _, f := range fitnesses {
		if !math.IsInf(f, 0) && !math.IsNaN(f) {
			totalFitness += f
			validCount++
			if f > bestFitness {
				bestFitness = f
			}
			if f < worstFitness {
				worstFitness = f
			}
		}
	}

	avgFitness := 0.0
	if validCount > 0 {
		avgFitness = totalFitness / float64(validCount)
	}

	idx := len(ms.History.Generations) - 1
	ms.History.Generations[idx].AvgFitness = avgFitness
	ms.History.Generations[idx].BestFitness = bestFitness
	ms.History.Generations[idx].WorstFitness = worstFitness

	fmt.Printf("\t\tMarket Price: $%.2f, Best: $%.2f, Avg: $%.2f\n",
		ms.Market.CurrentPrice, bestFitness, avgFitness)

	ms.generation++
}

func (ms *MarketSimulator) ResetOffspring(offspring []genomes.Genotype) {
	for i := range offspring {
		offspring[i].Attributes = make(map[string]interface{})
		offspring[i].Attributes["cash"] = ms.InitialFunds
		offspring[i].Attributes["holdings"] = 0
	}
}

func (ms *MarketSimulator) generateOrder(g *genomes.Genotype, id int) Order {
	if g.Attributes == nil {
		g.Attributes = make(map[string]interface{})
		g.Attributes["cash"] = ms.InitialFunds
		g.Attributes["holdings"] = 0
	}

	funds := ms.InitialFunds
	holdings := 0

	if cashVal, ok := g.Attributes["cash"]; ok && cashVal != nil {
		if f, ok := cashVal.(float64); ok {
			funds = f
		}
	}

	if holdingsVal, ok := g.Attributes["holdings"]; ok && holdingsVal != nil {
		if h, ok := holdingsVal.(int); ok {
			holdings = h
		}
	}

	exprStr := g.MapToGrammar(ms.Grammar, 7).String()
	program, err := expr.Compile(exprStr, expr.Env(map[string]interface{}{
		"$PRICE": 0.0,
	}))

	if err != nil {
		return Order{GenotypeID: id, Action: "HOLD", Quantity: 0}
	}

	out, err := expr.Run(program, map[string]interface{}{
		"$PRICE": ms.Market.CurrentPrice,
	})

	if err != nil {
		return Order{GenotypeID: id, Action: "HOLD", Quantity: 0}
	}

	action := "HOLD"
	if str, ok := out.(string); ok {
		action = str
	}
	var quantity int
	switch action {
	case "BUY":
		if funds >= ms.Market.CurrentPrice {
			quantity = int(funds / ms.Market.CurrentPrice)
		}
	case "SELL":
		quantity = holdings
	default:
		quantity = 0
	}

	return Order{
		GenotypeID: id,
		Action:     action,
		Quantity:   quantity,
	}
}

func (ms *MarketSimulator) calculateNewPrice(buyOrders, sellOrders int) float64 {
	totalOrders := buyOrders + sellOrders

	if totalOrders == 0 {
		return ms.Market.CurrentPrice
	}

	netDemand := buyOrders - sellOrders
	impactFactor := 0.05
	priceChange := (float64(netDemand) / float64(totalOrders)) * impactFactor

	fundamentalValue := ms.Market.InitialPrice
	meanReversionStrength := 0.1
	meanReversion := (fundamentalValue - ms.Market.CurrentPrice) / ms.Market.CurrentPrice * meanReversionStrength

	newPrice := ms.Market.CurrentPrice * (1.0 + priceChange + meanReversion)

	if newPrice < 1.0 {
		newPrice = 1.0
	}

	return newPrice
}

func (ms *MarketSimulator) executeOrder(g *genomes.Genotype, order Order, executionPrice float64) {
	funds := g.Attributes["cash"].(float64)
	holdings := g.Attributes["holdings"].(int)

	switch order.Action {
	case "BUY":
		maxAffordable := int(funds / executionPrice)
		actualQuantity := min(order.Quantity, maxAffordable)

		if actualQuantity > 0 {
			cost := float64(actualQuantity) * executionPrice
			funds -= cost
			holdings += actualQuantity
		}

	case "SELL":
		actualQuantity := min(order.Quantity, holdings)

		if actualQuantity > 0 {
			proceeds := float64(actualQuantity) * executionPrice
			funds += proceeds
			holdings -= actualQuantity
		}
	}

	g.Attributes["cash"] = funds
	g.Attributes["holdings"] = holdings
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (mh *MarketHistory) ExportJSON(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(mh)
}

func MinPrice(prices []float64) float64 {
	if len(prices) == 0 {
		return 0
	}
	min := prices[0]
	for _, p := range prices {
		if p < min {
			min = p
		}
	}
	return min
}

func MaxPrice(prices []float64) float64 {
	if len(prices) == 0 {
		return 0
	}
	max := prices[0]
	for _, p := range prices {
		if p > max {
			max = p
		}
	}
	return max
}

func SumVolume(volumes []int) int {
	sum := 0
	for _, v := range volumes {
		sum += v
	}
	return sum
}

func FindBestGeneration(gens []GenerationSnapshot) GenerationSnapshot {
	if len(gens) == 0 {
		return GenerationSnapshot{}
	}
	best := gens[0]
	for _, g := range gens {
		if g.AvgFitness > best.AvgFitness {
			best = g
		}
	}
	return best
}
