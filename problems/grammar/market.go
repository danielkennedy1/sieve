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
	Market          MarketState
	Grammar         genomes.Grammar
	MaxGenes        int
	InitialFunds    float64
	InitialHoldings int
	RoundsPerGen    int
	generation      int
	Results         []StrategyResult
	Config          *MarketConfig
	History         *MarketHistory
	Rng             *rand.Rand
	Generation      int
	MarketStates    []MarketState
}

func NewMarketSimulator(grammar genomes.Grammar, initialPrice, initialFunds float64, initialHoldings, roundsPerGen, maxGenes int, rng *rand.Rand) *MarketSimulator {
	return &MarketSimulator{
		Market:          *NewMarketState(initialPrice),
		History:         NewMarketHistory(),
		Grammar:         grammar,
		MaxGenes:        maxGenes,
		InitialFunds:    initialFunds,
		InitialHoldings: int(initialHoldings),
		RoundsPerGen:    roundsPerGen,
		generation:      0,
		Rng:             rng,
	}
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

type StrategyResult struct {
	Id           int
	Strategy     string
	ActiveReturn float64
	SharpeRatio  float64
}

type MarketState struct {
	CurrentPrice          float64
	InitialPrice          float64
	PriceHistory          []float64
	VolumeHistory         []int
	Price                 float64
	Volume                int
	FundamentalValue      float64
	RelativeStrengthIndex float64
	SimpleMovingAverage   float64
	AverageTrueRange      float64
	Participants          []Participant
}

type Participant struct {
	Id                    int
	Strategy              string
	Funds                 float64
	Holdings              int
	ExecutedTradeCount    int
	Solvent               bool
	PortfolioValueHistory []float64
}

type MarketConfig struct {
	Grammar                              genomes.Grammar
	MaxGenes                             int
	InitialPrice                         float64
	InitialFunds                         float64
	RiskFreeRate                         float64
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
		PriceHistory:          []float64{ms.Config.InitialPrice},
		VolumeHistory:         []int{0},
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
		state := MarketState{
			Price:                 initialState.Price,
			Volume:                initialState.Volume,
			FundamentalValue:      initialState.FundamentalValue,
			RelativeStrengthIndex: initialState.RelativeStrengthIndex,
			SimpleMovingAverage:   initialState.SimpleMovingAverage,
			AverageTrueRange:      initialState.AverageTrueRange,
			PriceHistory:          make([]float64, len(initialState.PriceHistory)),
			VolumeHistory:         make([]int, len(initialState.VolumeHistory)),
			Participants:          make([]Participant, len(initialState.Participants)),
		}
		copy(state.PriceHistory, initialState.PriceHistory)
		copy(state.VolumeHistory, initialState.VolumeHistory)
		copy(state.Participants, initialState.Participants)

		marketStates = append(marketStates, state)
	}

	for i := range marketStates {
		for round := 0; round < ms.Config.RoundsPerSim; round++ {

			if round%(ms.Config.RoundsPerSim/ms.Config.FundamentalValueChangesPerSimulation) == 0 {
				marketStates[i].FundamentalValue = ms.Config.InitialPrice + (ms.Config.InitialPrice * (ms.Rng.Float64() - 0.5))
			}

			for j := range marketStates[i].Participants {
				marketStates[i].Participants[j].Solvent = (marketStates[i].Participants[j].Funds + float64(marketStates[i].Participants[j].Holdings)*marketStates[i].Price) > 0
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

			// Calculate the new price based on ALL orders (real participants + extra noise)
			newPrice := calculateNewPrice(
				marketStates[i].Price,
				orders,
				marketStates[i].FundamentalValue,
				ms.Config.DemandPushCoefficient,
				ms.Config.FundamentalPullCoefficient,
			)

			marketStates[i].Price = newPrice

			for j, o := range realOrders {
				ms.executeOrder(&marketStates[i].Participants[j], o, marketStates[i])
			}

			for j := range marketStates[i].Participants {
				ms.trackPortfolioValue(&marketStates[i].Participants[j], marketStates[i].Price)
			}

			marketStates[i].PriceHistory = append(marketStates[i].PriceHistory, marketStates[i].Price)
			currentHistory := marketStates[i].PriceHistory

			marketStates[i].RelativeStrengthIndex = relativeStrengthIndex(currentHistory, ms.Config.RSIPeriod)
			marketStates[i].Volume = buyVolume + sellVolume
			marketStates[i].AverageTrueRange = averageTrueRange(currentHistory, ms.Config.ATRPeriod)
			marketStates[i].SimpleMovingAverage = simpleMovingAverage(currentHistory, ms.Config.SMAPeriod)

			if i == 0 {
				ms.History.Timestamps = append(ms.History.Timestamps, ms.Generation*ms.Config.RoundsPerSim+round)
				ms.History.Prices = append(ms.History.Prices, marketStates[i].Price)
				ms.History.Volumes = append(ms.History.Volumes, buyVolume+sellVolume)
			}
		}
	}

	results := []StrategyResult{}

	for genotypeId := range *genotypes {

		totalSharpe := 0.0
		results = append(results, StrategyResult{
			Id:           genotypeId,
			Strategy:     marketStates[0].Participants[genotypeId].Strategy,
			ActiveReturn: 0,
			SharpeRatio:  0,
		})
		for marketIdx := range ms.Config.SimsPerGeneration {
			// If the strategy blows up the account in any market, bin it
			if !marketStates[marketIdx].Participants[genotypeId].Solvent {
				results[genotypeId].ActiveReturn = math.Inf(-1)
				break
			}
			sharpe := calculateSharpeRatio(marketStates[marketIdx].Participants[genotypeId].PortfolioValueHistory, ms.Config.RiskFreeRate)
			portfolioValue := marketStates[marketIdx].Participants[genotypeId].Funds + float64(marketStates[marketIdx].Participants[genotypeId].Holdings)*marketStates[marketIdx].Price
			passivePortfolioValue := ms.Config.InitialFunds + float64(ms.Config.InitialHoldings)*marketStates[marketIdx].Price
			// sharpeMultiplier := math.Max(0, math.Min(sharpe, 3.0))
			results[genotypeId].ActiveReturn += portfolioValue - passivePortfolioValue
			totalSharpe += sharpe
		}
		results[genotypeId].SharpeRatio = totalSharpe / float64(ms.Config.SimsPerGeneration)
	}
	ms.MarketStates = marketStates
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

	// --- Add Check Here ---
	if bestFitness == worstFitness {
		tm.Println("Skipping histogram: all individuals have the same fitness.")
		tm.Flush()
		ms.Generation++
		return
	}
	// ----------------------

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

	// print price graph
	priceChart := tm.NewLineChart(100, 20)
	priceData := new(tm.DataTable)
	priceData.AddColumn("Round")
	priceData.AddColumn("Price")

	for i := range ms.History.Prices {
		priceData.AddRow(float64(i), ms.History.Prices[i])
	}
	tm.Println(priceChart.Draw(priceData))
	tm.Flush()

	ms.Generation++
}

func (ms *MarketSimulator) generateOrder(p Participant, s MarketState, progress float64) Order {
	if !p.Solvent {
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
		"$RANDOM":      ms.Rng.Float64(),
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
		switch o.Action {
		case "BUY":
			buyVolume += o.Quantity
		case "SELL":
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

func calculateSharpeRatio(returns []float64, riskFreeRate float64) float64 {
	if len(returns) < 2 {
		return 0.0
	}

	excessReturns := make([]float64, len(returns))
	for i, r := range returns {
		excessReturns[i] = r - riskFreeRate
	}

	sum := 0.0
	for _, r := range excessReturns {
		sum += r
	}
	meanExcessReturn := sum / float64(len(excessReturns))

	varianceSum := 0.0
	for _, r := range excessReturns {
		diff := r - meanExcessReturn
		varianceSum += diff * diff
	}
	stdDev := math.Sqrt(varianceSum / float64(len(excessReturns)))

	if stdDev == 0 {
		return 0.0
	}

	return meanExcessReturn / stdDev
}

func calculateReturns(portfolioValues []float64) []float64 {
	if len(portfolioValues) < 2 {
		return []float64{}
	}

	returns := make([]float64, len(portfolioValues)-1)
	for i := 1; i < len(portfolioValues); i++ {
		if portfolioValues[i-1] != 0 {
			returns[i-1] = (portfolioValues[i] - portfolioValues[i-1]) / portfolioValues[i-1]
		}
	}
	return returns
}

func (ms *MarketSimulator) trackPortfolioValue(participant *Participant, currentPrice float64) {
	portfolioValue := participant.Funds + float64(participant.Holdings)*currentPrice
	participant.PortfolioValueHistory = append(participant.PortfolioValueHistory, portfolioValue)
}
