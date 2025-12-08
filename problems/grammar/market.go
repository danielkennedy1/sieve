package grammar

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"slices"
	"strconv"
	"strings"

	tm "github.com/buger/goterm"
	"github.com/expr-lang/expr"

	"github.com/danielkennedy1/sieve/genomes"
)

type MarketSimulator struct {
	Results    []StrategyResult
	Config     *MarketConfig
	History    *MarketHistory
	Rng        *rand.Rand
	Generation int
}

type StrategyResult struct {
	Id           int
	Strategy     string
	ActiveReturn float64
}

type MarketState struct {
	Price                 float64
	Volume                int
	FundamentalValue      float64
	RelativeStrengthIndex float64
	SimpleMovingAverage   float64
	AverageTrueRange      float64
	Participants          []Participant
}

type Participant struct {
	Id                 int
	Strategy           string
	Funds              float64
	Holdings           int
	ExecutedTradeCount int
	Solvent            bool
}

type MarketConfig struct {
	Grammar                              genomes.Grammar
	MaxGenes                             int
	InitialPrice                         float64
	InitialFunds                         float64
	InitialHoldings                      int
	RoundsPerSim                         int
	NoiseOrdersPerRound                  int
	SimsPerGeneration                    int
	FundamentalValueChangesPerSimulation int
	DemandPushCoefficient                float64
	FundamentalPullCoefficient           float64
	RSIPeriod                            int
	ATRPeriod                            int
	SMAPeriod                            int
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

func (ms *MarketSimulator) NewMarketFitness() func(g genomes.Genotype) float64 {
	return func(g genomes.Genotype) float64 {
		if g.Attributes == nil {
			return 0
		}

		genotypeId := 0

		if idAny, ok := g.Attributes["id"]; ok && idAny != nil {
			if id, ok := idAny.(int); ok {
				genotypeId = id
			}
		}

		return ms.Results[genotypeId].ActiveReturn
	}

}

// FIXME: stateHistory takes a copy of all participants because it's a list of state objects, may be worth changing how participants
// are stored so they're not copied N*rounds*generations (not great)

func (ms *MarketSimulator) BeforeGeneration(genotypes *[]genomes.Genotype) {

	totalBuyVolume := 0
	totalSellVolume := 0

	marketStates := []MarketState{}

	initialState := MarketState{
		Price:                 ms.Config.InitialPrice,
		Volume:                0,
		FundamentalValue:      ms.Config.InitialPrice + (ms.Config.InitialPrice * (ms.Rng.Float64() - 0.5)),
		RelativeStrengthIndex: 50.0,
		SimpleMovingAverage:   ms.Config.InitialPrice,
		AverageTrueRange:      0.0,
		Participants:          make([]Participant, len(*genotypes)),
	}

	for i, g := range *genotypes {
		if (*genotypes)[i].Attributes == nil {
			(*genotypes)[i].Attributes = make(map[string]any)
		}
		(*genotypes)[i].Attributes["id"] = i

		initialState.Participants[i] = Participant{
			Id:                 i,
			Strategy:           g.MapToGrammar(ms.Config.Grammar, ms.Config.MaxGenes).String(),
			Funds:              ms.Config.InitialFunds,
			Holdings:           ms.Config.InitialHoldings,
			ExecutedTradeCount: 0,
		}
	}

	for range ms.Config.SimsPerGeneration {
		state := initialState
		state.Participants = make([]Participant, len(*genotypes))
		copy(state.Participants, initialState.Participants)
		marketStates = append(marketStates, state)
	}

	for i := range marketStates {
		for round := 0; round < ms.Config.RoundsPerSim; round++ {

			if round%(ms.Config.RoundsPerSim/ms.Config.FundamentalValueChangesPerSimulation) == 0 {
				marketStates[i].FundamentalValue = ms.Config.InitialPrice + (ms.Config.InitialPrice * (ms.Rng.Float64() - 0.5))
			}

			for j := range marketStates[i].Participants {
			    marketStates[i].Participants[j].Solvent = (marketStates[i].Participants[j].Funds + float64(marketStates[i].Participants[j].Holdings) * marketStates[i].Price) > 0
			}

			realOrders := make([]Order, len(marketStates[i].Participants))

			for j, p := range marketStates[i].Participants {
				order := ms.generateOrder(p, marketStates[i], float64(round)/float64(ms.Config.RoundsPerSim))
				realOrders[j] = order
			}

			buyVolume := 0
			sellVolume := 0
			for _, order := range realOrders {
				switch order.Action {
				case "BUY":
					buyVolume += order.Quantity
				case "SELL":
					sellVolume += order.Quantity
				}
			}

			totalBuyVolume += buyVolume
			totalSellVolume += sellVolume

			noiseOrders := ms.generateNoiseOrders(ms.Config.NoiseOrdersPerRound)

			orders := append(realOrders, noiseOrders...)


			for j, o := range realOrders {
				ms.executeOrder(&marketStates[i].Participants[j], o, marketStates[i])
			}

			marketStates[i].Price = calculateNewPrice(marketStates[i].Price, orders, marketStates[i].FundamentalValue, ms.Config.DemandPushCoefficient, ms.Config.FundamentalPullCoefficient)

			priceHistory := make([]float64, round+1)

			marketStates[i].RelativeStrengthIndex = relativeStrengthIndex(priceHistory, ms.Config.RSIPeriod)
			marketStates[i].Volume = buyVolume + sellVolume
			marketStates[i].AverageTrueRange = averageTrueRange(priceHistory, ms.Config.ATRPeriod)
			marketStates[i].SimpleMovingAverage = simpleMovingAverage(priceHistory, ms.Config.SMAPeriod)

			ms.History.Timestamps = append(ms.History.Timestamps, ms.Generation*ms.Config.RoundsPerSim+round)
			ms.History.Prices = append(ms.History.Prices, priceHistory...)
			ms.History.Volumes = append(ms.History.Volumes, buyVolume+sellVolume)
		}
	}

	results := []StrategyResult{}

	for genotypeId := range *genotypes {
		results = append(results, StrategyResult{
			Id:           genotypeId,
			Strategy:     marketStates[0].Participants[genotypeId].Strategy,
			ActiveReturn: 0,
		})
		for marketIdx := range ms.Config.SimsPerGeneration {
			// If the strategy blows up the account in any market, bin it
			if (!marketStates[marketIdx].Participants[genotypeId].Solvent) {
				results[genotypeId].ActiveReturn = math.Inf(-1)
				break
			}

			portfolioValue := marketStates[marketIdx].Participants[genotypeId].Funds + float64(marketStates[marketIdx].Participants[genotypeId].Holdings)*marketStates[marketIdx].Price
			passivePortfolioValue := ms.Config.InitialFunds + float64(ms.Config.InitialHoldings)*marketStates[marketIdx].Price
			results[genotypeId].ActiveReturn += portfolioValue - passivePortfolioValue
		}
	}

	ms.Results = results

	//ms.History.Generations = append(ms.History.Generations, GenerationSnapshot{
	//	Generation: ms.Generation,
	//	FinalPrice: state.Price,
	//	BuyOrders:  totalBuyVolume,
	//	SellOrders: totalSellVolume,
	//})

	//ms.showChart(stateHistory)
}

func (ms MarketSimulator) showChart(stateHistory []MarketState) {
	chart := tm.NewLineChart(100, 20)

	data := new(tm.DataTable)
	data.AddColumn("Round")
	data.AddColumn("Price")
	data.AddColumn("Fundamental Value")
	data.AddColumn("Initial Price")

	for i := range len(stateHistory) {
		data.AddRow(float64(i), stateHistory[i].Price, stateHistory[i].FundamentalValue, ms.Config.InitialPrice)
	}

	tm.Println(chart.Draw(data))
	tm.Flush()
}

func (ms *MarketSimulator) AfterGeneration(fitnesses []float64) {

	totalFitness := 0.0
	bestFitness := -math.MaxFloat64
	worstFitness := math.MaxFloat64
	survivorCount := 0

	bestFitnessIdx := -1

	for i, f := range fitnesses {
		if !math.IsInf(f, 0) && !math.IsNaN(f) {
			totalFitness += f
			survivorCount++
			if f > bestFitness {
				bestFitness = f
				bestFitnessIdx = i
			}
			if f < worstFitness {
				worstFitness = f
			}
		}
	}

	//avgFitness := 0.0
	//if validCount > 0 {
	//	avgFitness = totalFitness / float64(validCount)
	//}

	//idx := len(ms.History.Generations) - 1
	//ms.History.Generations[idx].AvgFitness = avgFitness
	//ms.History.Generations[idx].BestFitness = bestFitness
	//ms.History.Generations[idx].WorstFitness = worstFitness

	//fmt.Printf("\t\tMarket Price: $%.2f, Fundamental Value: $%.2f, Best fitness: %.2f, Avg fitness: %.2f\n", ms.FinalState.Price, ms.FinalState.FundamentalValue, bestFitness, avgFitness)

	fmt.Println("Survivor count: ", survivorCount)
	fmt.Println("Highest fitness strategy: ", ms.Results[bestFitnessIdx].Strategy)
	//fmt.Println("Trade count: ", ms.FinalState.Participants[bestFitnessIdx].ExecutedTradeCount)
	fmt.Println("Fitness: ", bestFitness)

	slices.Sort(fitnesses)

	// Histogram
	numBins := 20
	binWidth := (bestFitness - worstFitness) / float64(numBins)
	bins := make([]int, numBins)

	for _, f := range fitnesses {
		if math.IsInf(f, 0) || math.IsNaN(f) {
			continue
		}
		binIdx := int((f - worstFitness) / binWidth)
		if binIdx >= numBins {
			binIdx = numBins - 1
		}
		bins[binIdx]++
	}

	maxCount := 0
	for _, count := range bins {
		if count > maxCount {
			maxCount = count
		}
	}

	histogram := tm.NewLineChart(100, 20)
	histogramData := new(tm.DataTable)
	histogramData.AddColumn("Fitness")
	histogramData.AddColumn("Frequency")

	f := (worstFitness + binWidth) / 2
	for i := range bins {
		histogramData.AddRow(f, float64(bins[i]))
		f += binWidth
	}

	tm.Println(histogram.Draw(histogramData))
	tm.Flush()

	ms.Generation++
}

func (ms *MarketSimulator) generateOrder(p Participant, s MarketState, progress float64) Order {
	if (!p.Solvent) {
		return Order{GenotypeID: p.Id, Action: "HOLD", Quantity: 0}
	}
	program, err := expr.Compile(p.Strategy)

	if err != nil {
		fmt.Println("Error compiling expression for Genotype", p.Id, ":", err)
		return Order{GenotypeID: p.Id, Action: "HOLD", Quantity: 0}
	}

	out, err := expr.Run(program, map[string]any{
		"$PRICE":       s.Price,
		"$RSI":         s.RelativeStrengthIndex,
		"$PROGRESS":    progress,
		"$CASH":        p.Funds,
		"$HOLDINGS":    p.Holdings,
		"$VOLUME":      s.Volume,
		"$ATR":         s.AverageTrueRange,
		"$SMA":         s.SimpleMovingAverage,
		"$FUNDAMENTAL": s.FundamentalValue,
	})

	if err != nil {
		return Order{GenotypeID: p.Id, Action: "HOLD", Quantity: 0}
	}

	str, ok := out.(string)

	if !ok {
		return Order{GenotypeID: p.Id, Action: "HOLD", Quantity: 0}
	}

	elements := strings.Split(strings.Trim(str, " "), " ")

	action := elements[0]

	if action == "HOLD" {
		return Order{
			GenotypeID: p.Id,
			Action:     "HOLD",
			Quantity:   0,
		}
	}

	quantity, err := strconv.ParseInt(elements[1], 10, 64)

	if err != nil {
		fmt.Println("Error parsing int from ", str)
		return Order{GenotypeID: p.Id, Action: "HOLD", Quantity: 0}
	}

	return Order{
		GenotypeID: p.Id,
		Action:     action,
		Quantity:   int(quantity),
	}
}

func calculateNewPrice(price float64, orders []Order, fundamentalValue, demandPushCoefficient, fundamentalPullcoefficient float64) float64 {
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
	demandPush := (float64(netDemand) / float64(totalVolume)) * demandPushCoefficient

	fundamentalPull := (fundamentalValue - price) * fundamentalPullcoefficient

	newPrice := price + demandPush + fundamentalPull

	if newPrice < 1.0 {
		newPrice = 1.0
	}

	return newPrice
}

func (ms *MarketSimulator) executeOrder(participant *Participant, order Order, state MarketState) {
	switch order.Action {
	case "BUY":
		cost := float64(order.Quantity) * state.Price
		participant.Funds -= cost
		participant.Holdings += order.Quantity
		participant.ExecutedTradeCount++

	case "SELL":
		proceeds := float64(order.Quantity) * state.Price
		participant.Funds += proceeds
		participant.Holdings -= order.Quantity
		participant.ExecutedTradeCount++
	}
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

func (ms MarketSimulator) generateNoiseOrders(count int) []Order {
	orders := make([]Order, count)

	direction := ms.Rng.Float64()

	for i := 0; i < count; i++ {
		action := "HOLD"
		quantity := 0

		r := ms.Rng.Float64()
		if r < direction {
			action = "SELL"
			quantity = ms.Rng.IntN(100) + 5
		} else {
			action = "BUY"
			quantity = ms.Rng.IntN(100) + 5
		}

		orders[i] = Order{
			GenotypeID: -1, // flag as noise trader
			Action:     action,
			Quantity:   quantity,
		}
	}
	return orders
}
