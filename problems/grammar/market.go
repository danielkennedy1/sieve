package grammar

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"

	"github.com/danielkennedy1/sieve/genomes"
	"github.com/expr-lang/expr"
)

type MarketState struct {
	FinalPrice  float64
	CurrentVolume int
	CurrentATR    float64
	CurrentSMA    float64
	InitialPrice  float64
	PriceHistory  []float64
	VolumeHistory []int
	TradesPerActor []int
	FundamentalValue float64
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
		FinalPrice:  initialPrice,
		InitialPrice:  initialPrice,
		PriceHistory:  []float64{initialPrice},
		VolumeHistory: []int{},
		TradesPerActor: []int{},
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
	Market          *MarketState
	History         *MarketHistory
	Grammar         genomes.Grammar
	Rng 			*rand.Rand
	MaxGenes        int
	InitialFunds    float64
	InitialHoldings int
	RoundsPerGen    int
	generation      int
}

func NewMarketSimulator(grammar genomes.Grammar, rng *rand.Rand ,initialPrice, initialFunds float64, initialHoldings, roundsPerGen, maxGenes int) *MarketSimulator {
	return &MarketSimulator{
		Market:          NewMarketState(initialPrice),
		History:         NewMarketHistory(),
		Grammar:         grammar,
		Rng: 			 rng,
		MaxGenes:        maxGenes,
		InitialFunds:    initialFunds,
		InitialHoldings: int(initialHoldings),
		RoundsPerGen:    roundsPerGen,
		generation:      0,
	}
}

func (ms *MarketSimulator) NewMarketFitness() func(g genomes.Genotype) float64 {
	return func(g genomes.Genotype) float64 {
		if g.Attributes == nil {
			return 0
		}

		funds := 0.0

		if cashVal, ok := g.Attributes["cash"]; ok && cashVal != nil {
			if f, ok := cashVal.(float64); ok {
				funds = f
			}
		}

		holdings := 0.0

		if h, ok := g.Attributes["holdings"]; ok && h != nil {
			if h, ok := h.(float64); ok {
				holdings = h
			}
		}

		return funds + holdings * ms.Market.FinalPrice
	}

}

func (ms *MarketSimulator) BeforeGeneration(genotypes []genomes.Genotype) {
	totalBuyVolume := 0
	totalSellVolume := 0

	strategies := make([]string, len(genotypes))

	for i, g := range genotypes {
		strategies[i] = g.MapToGrammar(ms.Grammar, ms.MaxGenes).String()
	}

	p := ms.Market.InitialPrice
	rsi := 50.0

	f := ms.Market.InitialPrice + (ms.Market.InitialPrice * (ms.Rng.Float64()))
	ms.Market.FundamentalValue = f

	tradesPerActor := make([]int, len(genotypes))

	for _, g := range genotypes {
		g.Attributes["cash"] = ms.InitialFunds
		g.Attributes["holdings"] = ms.InitialHoldings
	}

	for round := 0; round < ms.RoundsPerGen; round++ {
		var orders []Order

		for i, g := range genotypes {
			order := ms.generateOrder(&g, i, strategies[i], p, rsi, float64(round) / float64(ms.RoundsPerGen))
			orders = append(orders, order)
		}

		buyVolume := 0
		sellVolume := 0
		for _, order := range orders {
			switch order.Action {
			case "BUY":
				buyVolume += order.Quantity
			case "SELL":
				sellVolume += order.Quantity
			}
		}

		totalBuyVolume += buyVolume
		totalSellVolume += sellVolume

		p = calculateNewPrice(p, orders, f)

		for i, order := range orders {
			ms.executeOrder(&genotypes[i], order, p, &tradesPerActor[i])
		}

		ms.Market.FinalPrice = p
		ms.Market.PriceHistory = append(ms.Market.PriceHistory, p)

		rsi = calculateRSI(ms.Market.PriceHistory, 14)
		ms.Market.CurrentVolume = buyVolume + sellVolume
		ms.Market.CurrentATR = calculateATR(ms.Market.PriceHistory, 20)
		ms.Market.CurrentSMA = calculateSMA(ms.Market.PriceHistory, 14)
		ms.Market.TradesPerActor = tradesPerActor

		ms.History.Timestamps = append(ms.History.Timestamps, ms.generation*ms.RoundsPerGen+round)
		ms.History.Prices = append(ms.History.Prices, p)
		ms.History.Volumes = append(ms.History.Volumes, buyVolume+sellVolume)
	}
	ms.Market.FinalPrice = p

	ms.History.Generations = append(ms.History.Generations, GenerationSnapshot{
		Generation: ms.generation,
		FinalPrice: p,
		BuyOrders:  totalBuyVolume,
		SellOrders: totalSellVolume,
	})

	fitnessFunction := ms.NewMarketFitness()

	fitnesses := make([]float64, len(genotypes))


	for i, g := range genotypes {
		fitnesses[i] = fitnessFunction(g)
	}

	bestFitness := math.Inf(-1)
	bestFitnessIdx := -1

	for i, f := range fitnesses {
		if !math.IsInf(f, 0) && !math.IsNaN(f) {
			if f > bestFitness {
				bestFitness = f
				bestFitnessIdx = i
			}
		}
	}

	fmt.Println("Best Individual: ", genotypes[bestFitnessIdx].MapToGrammar(ms.Grammar, ms.MaxGenes).String())

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

	fmt.Printf("\t\tMarket Price: $%.2f, Fundamental Value: $%.2f, Best fitness: %.2f, Avg fitness: %.2f\n", ms.Market.FinalPrice, ms.Market.FundamentalValue, bestFitness, avgFitness)

	ms.generation++
}

func (ms *MarketSimulator) ResetOffspring(offspring []genomes.Genotype) {
	for i := range offspring {
		offspring[i].Attributes = make(map[string]any)
		offspring[i].Attributes["cash"] = ms.InitialFunds
		offspring[i].Attributes["holdings"] = ms.InitialHoldings
	}
}

func (ms *MarketSimulator) generateOrder(g *genomes.Genotype, id int, strategy string, price, rsi, progress float64) Order {
	if g.Attributes == nil {
		g.Attributes = make(map[string]interface{})
		g.Attributes["cash"] = ms.InitialFunds
		g.Attributes["holdings"] = ms.InitialHoldings
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

	// fmt.Println("Evaluating strategy for Genotype", id, ":")
	// fmt.Println(exprStr)
	program, err := expr.Compile(strategy)

	if err != nil {
		fmt.Println("Error compiling expression for Genotype", id, ":", err)
		return Order{GenotypeID: id, Action: "HOLD", Quantity: 0}
	}

	out, err := expr.Run(program, map[string]any{
		"$PRICE":    price,
		"$RSI":      rsi,
		"$PROGRESS": progress,
		"$CASH":     funds,
		"$HOLDINGS": holdings,
		"$VOLUME":   ms.Market.CurrentVolume,
		"$ATR":      ms.Market.CurrentATR,
		"$SMA":      ms.Market.CurrentSMA,
		"$FUNDAMENTAL": ms.Market.FundamentalValue,
	})

	// fmt.Println(out)

	if err != nil {
		return Order{GenotypeID: id, Action: "HOLD", Quantity: 0}
	}

	str, ok := out.(string)

	if !ok {
		return Order{GenotypeID: id, Action: "HOLD", Quantity: 0}
	}

	elements := strings.Split(strings.Trim(str, " "), " ")

	action := elements[0]

	if action == "HOLD" {
		return Order{
			GenotypeID: id,
			Action:     "HOLD",
			Quantity:   0,
		}
	}

	var quantity int

	proportion, err := strconv.ParseFloat(elements[1], 64)

	if err != nil {
		return Order{GenotypeID: id, Action: "HOLD", Quantity: 0}
	}

	switch action {
	case "BUY":
		if funds >= price {
			quantity = int(funds / price * proportion)
		}
	case "SELL":
		quantity = holdings * int(proportion*100) / 100
	default:
		quantity = 0
	}

	return Order{
		GenotypeID: id,
		Action:     action,
		Quantity:   quantity,
	}
}

func calculateNewPrice(price float64, orders []Order, fundamentalValue float64) float64 {
	buyVolume := 0
	sellVolume := 0

	for _, o := range orders {
		if o.Action == "BUY" {
			buyVolume += o.Quantity
		} else if o.Action == "SELL" {
			sellVolume += o.Quantity
		}
	}

	totalVolume := buyVolume + sellVolume

	if totalVolume == 0 {
		return price
	}

	netDemand := buyVolume - sellVolume
	demandPush := (float64(netDemand) / float64(totalVolume)) * 0.05

	fundamentalPull := (fundamentalValue - price) * 0.1

	newPrice := price + demandPush + fundamentalPull

	if newPrice < 1.0 {
		newPrice = 1.0
	}

	return newPrice
}

func (ms *MarketSimulator) executeOrder(g *genomes.Genotype, order Order, executionPrice float64, tradeCount *int) {
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
			*tradeCount++
		}

	case "SELL":
		actualQuantity := min(order.Quantity, holdings)

		if actualQuantity > 0 {
			// fmt.Println("Executing SELL order for Genotype", order.GenotypeID, ": Selling", actualQuantity, "at price", executionPrice)
			proceeds := float64(actualQuantity) * executionPrice
			funds += proceeds
			holdings -= actualQuantity
			*tradeCount++
		} else {
			// fmt.Println("No holdings to sell for Genotype", order.GenotypeID)
		}
	}

	g.Attributes["cash"] = funds
	g.Attributes["holdings"] = holdings
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

func calculateRSI(prices []float64, period int) float64 {
	if len(prices) <= period {
		return 50.0
	}

	initialGains := 0.0
	initialLosses := 0.0

	for i := 1; i <= period; i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			initialGains += change
		} else {
			initialLosses += -change
		}
	}

	avgGain := initialGains / float64(period)
	avgLoss := initialLosses / float64(period)

	for i := period + 1; i < len(prices); i++ {

		change := prices[i] - prices[i-1]
		currentGain := 0.0
		currentLoss := 0.0

		if change > 0 {
			currentGain = change
		} else {
			currentLoss = -change
		}

		avgGain = (avgGain*float64(period-1) + currentGain) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + currentLoss) / float64(period)
	}

	rs := 0.0
	if avgLoss == 0 {
		rs = math.MaxFloat64
	} else {
		rs = avgGain / avgLoss
	}

	rsi := 100.0 - (100.0 / (1.0 + rs))
	return rsi
}

func calculateSMA(prices []float64, period int) float64 {
	if len(prices) < period {
		if len(prices) > 0 {
			return prices[len(prices)-1]
		}
		return 0.0
	}

	sum := 0.0
	startIdx := len(prices) - period
	for i := startIdx; i < len(prices); i++ {
		sum += prices[i]
	}

	return sum / float64(period)
}

func calculateATR(prices []float64, period int) float64 {
	if len(prices) <= period {
		return 0.0
	}

	sumTR := 0.0
	for i := 1; i <= period; i++ {
		tr := math.Abs(prices[i] - prices[i-1])
		sumTR += tr
	}
	atr := sumTR / float64(period)

	for i := period + 1; i < len(prices); i++ {
		currentTR := math.Abs(prices[i] - prices[i-1])
		atr = ((atr * float64(period-1)) + currentTR) / float64(period)
	}

	return atr
}
