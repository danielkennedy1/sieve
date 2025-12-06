package grammar

import (
	"math"
)

func relativeStrengthIndex(prices []float64, period int) float64 {
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

func simpleMovingAverage(prices []float64, period int) float64 {
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

func averageTrueRange(prices []float64, period int) float64 {
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
