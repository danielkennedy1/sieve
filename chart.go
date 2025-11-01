package main

import (
	"fmt"
	"os"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

func EvaluateImprovementPopulation(oldPop, newPop [][]bool) float64 {
	oldBest, _ := GetBestFitness(oldPop)
	newBest, _ := GetBestFitness(newPop)

	if oldBest == 0 {
		return 0.0
	}
	improvement := (newBest - oldBest) / oldBest
	return improvement
}

func EvaluateFitnessPopulation(population [][]bool) []float64 {
	fitnessList := make([]float64, len(population))

	// Compute fitness for each individual
	for i := range population {
		fitnessList[i] = GetFitness(population[i])
	}

	bestFitness := fitnessList[0]
	worstFitness := fitnessList[0]
	sum := 0.0

	for _, f := range fitnessList {
		if f > bestFitness {
			bestFitness = f
		}
		if f < worstFitness {
			worstFitness = f
		}
		sum += f
	}

	averageFitness := sum / float64(len(fitnessList))

	return []float64{bestFitness, averageFitness, worstFitness}
}

func EvaluateDiversityPopulation(population [][]bool) float64 {
	geneLength := len(population[0])
	geneCounts := make([]int, geneLength)

	for _, individual := range population {
		for j, gene := range individual {
			if gene {
				geneCounts[j]++
			}
		}
	}

	diversitySum := 0.0
	popSize := float64(len(population))

	for _, count := range geneCounts {
		frequency := float64(count) / popSize
		diversitySum += frequency * (1 - frequency)
	}

	averageDiversity := diversitySum / float64(geneLength)
	return averageDiversity
}

func CreateChart(
	fitnessMetricsList [][]float64, diversityMetricsList, improvementMetricsList []float64,
) {
	// X-axis = generation indices 0..N-1
	n := len(fitnessMetricsList)
	x := make([]int, n)
	for i := 0; i < n; i++ {
		x[i] = i
	}

	best := make([]opts.LineData, 0, n)
	avg := make([]opts.LineData, 0, n)
	worst := make([]opts.LineData, 0, n)
	for _, row := range fitnessMetricsList {
		var b, a, w float64
		if len(row) > 0 {
			b = row[0]
		}
		if len(row) > 1 {
			a = row[1]
		}
		if len(row) > 2 {
			w = row[2]
		}
		best = append(best, opts.LineData{Value: b})
		avg = append(avg, opts.LineData{Value: a})
		worst = append(worst, opts.LineData{Value: w})
	}

	// Diversity
	div := make([]opts.LineData, 0, len(diversityMetricsList))
	for i := range diversityMetricsList {
		v := diversityMetricsList[i]
		div = append(div, opts.LineData{Value: v})
	}

	// Improvement
	imp := make([]opts.LineData, 0, len(improvementMetricsList))
	for i := range improvementMetricsList {
		v := improvementMetricsList[i]
		imp = append(imp, opts.LineData{Value: v})
	}

	// Fitness chart
	lf := charts.NewLine()
	lf.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			// makes it responsive
			PageTitle: "GA Metrics",
			Theme:     types.ThemeInfographic,
		}),
		charts.WithTitleOpts(opts.Title{
			Title: "Fitness over Generations",
			Left:  "center",
		}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(false), Trigger: "axis"}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "Generation",
			Type: "category",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "Fitness",
			Type: "value",
		}),
	)
	lf.SetXAxis(x).
		AddSeries("Best", best, charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)})).
		AddSeries("Average", avg, charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)})).
		AddSeries("Worst", worst, charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)}))

	// Diversity chart
	ld := charts.NewLine()
	ld.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Population Diversity",
			Left:  "center",
		}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Trigger: "axis"}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Generation", Type: "category"}),
		charts.WithYAxisOpts(opts.YAxis{Name: "Diversity", Type: "value"}),
	)
	ld.SetXAxis(x).
		AddSeries("", div, charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)}))

	// Improvement chart
	li := charts.NewLine()
	li.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Relative Improvement",
			Left:  "center",
		}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Trigger: "axis"}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Generation", Type: "category"}),
		charts.WithYAxisOpts(opts.YAxis{Name: "Improvement", Type: "value"}),
	)
	li.SetXAxis(x).
		AddSeries("", imp, charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)}))

	// Put all three on one page
	page := components.NewPage()
	page.PageTitle = "GA Metrics Dashboard"
	page.AddCharts(lf, ld, li)

	f, err := os.Create("ga_metrics.html")
	if err != nil {
		fmt.Println("failed to create chart file:", err)
		return
	}
	defer f.Close()

	if err := page.Render(f); err != nil {
		fmt.Println("failed to render charts:", err)
	}
}
