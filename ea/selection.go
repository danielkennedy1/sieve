package ea

import (
	"fmt"
	"math/rand/v2"
)

func Tournament(k int) func([]float64, int) []int {
	return func(fitnesses []float64, n int) []int {
		selected := make([]int, n)
		popSize := len(fitnesses)

		for i := range n {
			tournamentCandidates := make([]int, 0, k)

			// Initial best candidate
			best := rand.IntN(popSize)
			bestFit := fitnesses[best]
			tournamentCandidates = append(tournamentCandidates, best)

			// Compare against k-1 other candidates
			for j := 1; j < k; j++ {
				candidate := rand.IntN(popSize)
				tournamentCandidates = append(tournamentCandidates, candidate)
				if fitnesses[candidate] > bestFit {
					best = candidate
					bestFit = fitnesses[candidate]
				}
			}

			selected[i] = best

			// Printing the results
			candidateStrings := make([]string, k)
			for idx, c := range tournamentCandidates {
				candidateStrings[idx] = fmt.Sprintf("Idx %d (Fit %.2f)", c, fitnesses[c])
			}

		}

		return selected
	}
}
func Roulette() func([]float64, int) []int {
	return func(fitnesses []float64, n int) []int {
		var f_worst float64 = -1e308
		var f_best float64 = 1e308

		for _, f := range fitnesses {
			if f > f_worst {
				f_worst = f
			}
			if f < f_best {
				f_best = f
			}
		}

		scaledFitnesses := make([]float64, len(fitnesses))
		var total float64

		for i, f := range fitnesses {
			f_prime := f_worst - f
			scaledFitnesses[i] = f_prime
			total += f_prime
		}

		if total == 0 {
			selected := make([]int, n)
			for i := range n {
				selected[i] = rand.IntN(len(fitnesses))
			}
			return selected
		}

		selected := make([]int, n)
		for i := range n {
			r := rand.Float64() * total
			acc := 0.0

			picked := len(fitnesses) - 1

			for idx, f_prime := range scaledFitnesses {
				acc += f_prime

				if acc >= r {
					picked = idx
					break
				}
			}

			selected[i] = picked

		}

		fmt.Printf("--- Roulette Selection End ---\n")
		return selected
	}
}
