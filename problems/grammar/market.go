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
	FinalState *MarketState
	Config     *MarketConfig
	History    *MarketHistory
	Rng        *rand.Rand
	Generation int
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
}

type MarketConfig struct {
	Grammar         genomes.Grammar
	MaxGenes        int
	InitialPrice    float64
	InitialFunds    float64
	InitialHoldings int
	RoundsPerGen    int
	NoiseOrdersPerRound int
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

		participant := ms.FinalState.Participants[genotypeId]
		participantPortfolioValue := participant.Funds + float64(participant.Holdings)*ms.FinalState.Price

		passivePortfolioValue := ms.Config.InitialFunds + float64(ms.Config.InitialHoldings)*ms.FinalState.Price

		return participantPortfolioValue - passivePortfolioValue
	}

}

// FIXME: stateHistory takes a copy of all participants because it's a list of state objects, may be worth changing how participants
// are stored so they're not copied N*rounds*generations (not great)

func (ms *MarketSimulator) BeforeGeneration(genotypes []genomes.Genotype) {
	totalBuyVolume := 0
	totalSellVolume := 0

	state := MarketState{
		Price:                 ms.Config.InitialPrice,
		Volume:                0,
		FundamentalValue:      ms.Config.InitialPrice + (ms.Config.InitialPrice * (ms.Rng.Float64() - 0.5)),
		RelativeStrengthIndex: 50.0,
		SimpleMovingAverage:   ms.Config.InitialPrice,
		AverageTrueRange:      0.0,
		Participants:          make([]Participant, len(genotypes)),
	}

	for i, g := range genotypes {
		state.Participants[i] = Participant{
			Id:                 i,
			Strategy:           g.MapToGrammar(ms.Config.Grammar, ms.Config.MaxGenes).String(),
			Funds:              ms.Config.InitialFunds,
			Holdings:           ms.Config.InitialHoldings,
			ExecutedTradeCount: 0,
		}
	}

	stateHistory := make([]MarketState, ms.Config.RoundsPerGen)

	for round := 0; round < ms.Config.RoundsPerGen; round++ {

		if round % (ms.Config.RoundsPerGen/4) == 0 { // NOTE: Hardcoded regime changes per generation
			state.FundamentalValue = ms.Config.InitialPrice + (ms.Config.InitialPrice * (ms.Rng.Float64() - 0.5))
		}

		realOrders := make([]Order, len(state.Participants))

		for i, p := range state.Participants {
			order := ms.generateOrder(p, state, float64(round)/float64(ms.Config.RoundsPerGen))
			realOrders[i] = order
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

		state.Price = calculateNewPrice(state.Price, orders, state.FundamentalValue)

		for i, o := range realOrders {
			ms.executeOrder(&state.Participants[i], o, state)
		}

		stateHistory[round] = state

		priceHistory := make([]float64, round + 1)

		for i := range round {
			priceHistory[i] = stateHistory[i].Price
		}

		state.RelativeStrengthIndex = relativeStrengthIndex(priceHistory, 14) // NOTE: Hardcoded RSI period
		state.Volume = buyVolume + sellVolume
		state.AverageTrueRange = averageTrueRange(priceHistory, 20) // NOTE: Hardcoded ATR period
		state.SimpleMovingAverage = simpleMovingAverage(priceHistory, 14) // NOTE: Hardcoded SMA period

		ms.History.Timestamps = append(ms.History.Timestamps, ms.Generation*ms.Config.RoundsPerGen+round)
		ms.History.Prices = append(ms.History.Prices, priceHistory...)
		ms.History.Volumes = append(ms.History.Volumes, buyVolume+sellVolume)
	}

	ms.FinalState = &state

	ms.History.Generations = append(ms.History.Generations, GenerationSnapshot{
		Generation: ms.Generation,
		FinalPrice: state.Price,
		BuyOrders:  totalBuyVolume,
		SellOrders: totalSellVolume,
	})

	showChart(stateHistory)
}

func showChart(stateHistory []MarketState) {
    chart := tm.NewLineChart(100, 20)

    data := new(tm.DataTable)
    data.AddColumn("Round")
    data.AddColumn("Price")
    data.AddColumn("Fundamental Value")

	for i := range len(stateHistory) {
		data.AddRow(float64(i), stateHistory[i].Price, stateHistory[i].FundamentalValue)
    }
    
	tm.Println(chart.Draw(data))
	tm.Flush()
}

func (ms *MarketSimulator) AfterGeneration(fitnesses []float64) {
	totalFitness := 0.0
	bestFitness := -math.MaxFloat64
	worstFitness := math.MaxFloat64
	validCount := 0

	bestFitnessIdx := -1

	for i, f := range fitnesses {
		if !math.IsInf(f, 0) && !math.IsNaN(f) {
			totalFitness += f
			validCount++
			if f > bestFitness {
				bestFitness = f
				bestFitnessIdx = i
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

	fmt.Printf("\t\tMarket Price: $%.2f, Fundamental Value: $%.2f, Best fitness: %.2f, Avg fitness: %.2f\n", ms.FinalState.Price, ms.FinalState.FundamentalValue, bestFitness, avgFitness)

	fmt.Println("Highest fitness strategy: ", ms.FinalState.Participants[bestFitnessIdx].Strategy)

	slices.Sort(fitnesses)
    chart := tm.NewLineChart(100, 20)

    data := new(tm.DataTable)
    data.AddColumn("Rank")
    data.AddColumn("Fitness")

	for i := range len(fitnesses) {
		data.AddRow(float64(i), fitnesses[i])
    }
    
	tm.Println(chart.Draw(data))
	tm.Flush()


	ms.Generation++
}

func (ms *MarketSimulator) generateOrder(p Participant, s MarketState, progress float64) Order {
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

	var quantity int

	proportion, err := strconv.ParseFloat(elements[1], 64)

	if err != nil {
		return Order{GenotypeID: p.Id, Action: "HOLD", Quantity: 0}
	}

	switch action {
	case "BUY":
		if p.Funds >= s.Price {
			quantity = int(p.Funds/ s.Price * proportion)
		}
	case "SELL":
		quantity = p.Holdings * int(proportion*100) / 100
	default:
		quantity = 0
	}

	return Order{
		GenotypeID: p.Id,
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
	demandPush := (float64(netDemand) / float64(totalVolume)) * 0.05 // NOTE: Hardcoded demand push coefficient

	fundamentalPull := (fundamentalValue - price) * 0.1 // NOTE: Hardcoded fundamenal fundamental pull coefficient

	newPrice := price + demandPush + fundamentalPull

	if newPrice < 1.0 {
		newPrice = 1.0
	}

	return newPrice
}

func (ms *MarketSimulator) executeOrder(participant *Participant, order Order, state MarketState) {
	switch order.Action {
	case "BUY":
		maxAffordable := int(participant.Funds / state.Price)
		actualQuantity := min(order.Quantity, maxAffordable)

		if actualQuantity > 0 {
			cost := float64(actualQuantity) * state.Price
			participant.Funds -= cost
			participant.Holdings += actualQuantity
			participant.ExecutedTradeCount++
		}

	case "SELL":
		actualQuantity := min(order.Quantity, participant.Holdings)

		if actualQuantity > 0 {
			proceeds := float64(actualQuantity) * state.Price
			participant.Funds+= proceeds
			participant.Holdings -= actualQuantity
			participant.ExecutedTradeCount++
		}	
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
			quantity = ms.Rng.IntN(10) + 5
		} else {
			action = "BUY"
			quantity = ms.Rng.IntN(10) + 5
		}

		orders[i] = Order{
			GenotypeID: -1, // flag as noise trader
			Action:     action,
			Quantity:   quantity,
		}
	}
	return orders
}
