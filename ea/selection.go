package ea

import "math/rand"

func Tournament(k int) func([]float64, int) []int {
    return func(fitnesses []float64, n int) []int {
        selected := make([]int, n)
		for i := range n {
            best := rand.Intn(len(fitnesses))
            bestFit := fitnesses[best]
            
            for j := 1; j < k; j++ {
                candidate := rand.Intn(len(fitnesses))
                if fitnesses[candidate] > bestFit {
                    best = candidate
                    bestFit = fitnesses[candidate]
                }
            }
            selected[i] = best
        }
        return selected
    }
}
