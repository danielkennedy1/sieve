package charts

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/danielkennedy1/sieve/problems/grammar"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

// GenerateCharts creates all chart visualizations from market history data
func GenerateCharts(dataFile, outputDir string) error {
	history, err := loadMarketHistory(dataFile)
	if err != nil {
		return fmt.Errorf("failed to load market history: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	fmt.Println("Generating charts...")

	if err := createPriceChart(history, outputDir); err != nil {
		fmt.Printf("Error creating price chart: %v\n", err)
	} else {
		fmt.Println("✓ Price chart created")
	}

	if err := createVolumeChart(history, outputDir); err != nil {
		fmt.Printf("Error creating volume chart: %v\n", err)
	} else {
		fmt.Println("✓ Volume chart created")
	}

	if err := createFitnessChart(history, outputDir); err != nil {
		fmt.Printf("Error creating fitness chart: %v\n", err)
	} else {
		fmt.Println("✓ Fitness chart created")
	}

	if err := createOrderFlowChart(history, outputDir); err != nil {
		fmt.Printf("Error creating order flow chart: %v\n", err)
	} else {
		fmt.Println("✓ Order flow chart created")
	}

	if err := createCombinedDashboard(history, outputDir); err != nil {
		fmt.Printf("Error creating dashboard: %v\n", err)
	} else {
		fmt.Println("✓ Dashboard created")
	}

	fmt.Printf("\nCharts generated successfully in %s/\n", outputDir)
	fmt.Printf("Open %s/dashboard.html in your browser to view all charts\n", outputDir)

	return nil
}

func loadMarketHistory(filename string) (*grammar.MarketHistory, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var history grammar.MarketHistory
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, err
	}

	return &history, nil
}

func createPriceChart(history *grammar.MarketHistory, outputDir string) error {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme:  types.ThemeWesteros,
			Width:  "1400px",
			Height: "600px",
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    "Market Price Over Time",
			Subtitle: "Price evolution across generations",
		}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true)}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true)}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "slider",
			Start: 0,
			End:   100,
		}),
	)

	xAxis := make([]string, len(history.Timestamps))
	for i, t := range history.Timestamps {
		xAxis[i] = fmt.Sprintf("%d", t)
	}

	priceData := make([]opts.LineData, len(history.Prices))
	for i, p := range history.Prices {
		priceData[i] = opts.LineData{Value: p}
	}

	line.SetXAxis(xAxis).
		AddSeries("Price", priceData).
		SetSeriesOptions(
			charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)}),
			charts.WithLabelOpts(opts.Label{Show: opts.Bool(false)}),
		)

	f, err := os.Create(fmt.Sprintf("%s/price_chart.html", outputDir))
	if err != nil {
		return err
	}
	defer f.Close()

	return line.Render(f)
}

func createVolumeChart(history *grammar.MarketHistory, outputDir string) error {
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme:  types.ThemeWesteros,
			Width:  "1400px",
			Height: "600px",
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    "Trading Volume Over Time",
			Subtitle: "Buy and sell volumes",
		}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true)}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true)}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "slider",
			Start: 0,
			End:   100,
		}),
	)

	xAxis := make([]string, len(history.Timestamps))
	for i, t := range history.Timestamps {
		xAxis[i] = fmt.Sprintf("%d", t)
	}

	volumeData := make([]opts.BarData, len(history.Volumes))
	for i, v := range history.Volumes {
		volumeData[i] = opts.BarData{Value: v}
	}

	bar.SetXAxis(xAxis).
		AddSeries("Volume", volumeData)

	f, err := os.Create(fmt.Sprintf("%s/volume_chart.html", outputDir))
	if err != nil {
		return err
	}
	defer f.Close()

	return bar.Render(f)
}

func createFitnessChart(history *grammar.MarketHistory, outputDir string) error {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme:  types.ThemeWesteros,
			Width:  "1400px",
			Height: "600px",
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    "Fitness Evolution",
			Subtitle: "Best, average, and worst fitness per generation",
		}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true)}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true)}),
	)

	xAxis := make([]string, len(history.Generations))
	bestData := make([]opts.LineData, len(history.Generations))
	avgData := make([]opts.LineData, len(history.Generations))
	worstData := make([]opts.LineData, len(history.Generations))

	for i, gen := range history.Generations {
		xAxis[i] = fmt.Sprintf("%d", gen.Generation)
		bestData[i] = opts.LineData{Value: gen.BestFitness}
		avgData[i] = opts.LineData{Value: gen.AvgFitness}
		worstData[i] = opts.LineData{Value: gen.WorstFitness}
	}

	line.SetXAxis(xAxis).
		AddSeries("Best", bestData).
		AddSeries("Average", avgData).
		AddSeries("Worst", worstData).
		SetSeriesOptions(
			charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)}),
		)

	f, err := os.Create(fmt.Sprintf("%s/fitness_chart.html", outputDir))
	if err != nil {
		return err
	}
	defer f.Close()

	return line.Render(f)
}

func createOrderFlowChart(history *grammar.MarketHistory, outputDir string) error {
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme:  types.ThemeWesteros,
			Width:  "1400px",
			Height: "600px",
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    "Order Flow by Generation",
			Subtitle: "Buy vs Sell orders",
		}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true)}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true)}),
	)

	xAxis := make([]string, len(history.Generations))
	buyData := make([]opts.BarData, len(history.Generations))
	sellData := make([]opts.BarData, len(history.Generations))

	for i, gen := range history.Generations {
		xAxis[i] = fmt.Sprintf("Gen %d", gen.Generation)
		buyData[i] = opts.BarData{Value: gen.BuyOrders}
		sellData[i] = opts.BarData{Value: gen.SellOrders}
	}

	bar.SetXAxis(xAxis).
		AddSeries("Buy Orders", buyData).
		AddSeries("Sell Orders", sellData)

	f, err := os.Create(fmt.Sprintf("%s/order_flow_chart.html", outputDir))
	if err != nil {
		return err
	}
	defer f.Close()

	return bar.Render(f)
}

func createCombinedDashboard(history *grammar.MarketHistory, outputDir string) error {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Market Simulation Dashboard</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        h1 {
            text-align: center;
            color: #333;
        }
        .stats {
            background: white;
            padding: 20px;
            margin: 20px 0;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .stats h2 {
            margin-top: 0;
            color: #2c3e50;
        }
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            margin-top: 15px;
        }
        .stat-item {
            padding: 10px;
            background: #f8f9fa;
            border-radius: 4px;
        }
        .stat-label {
            font-size: 0.9em;
            color: #666;
            margin-bottom: 5px;
        }
        .stat-value {
            font-size: 1.3em;
            font-weight: bold;
            color: #2c3e50;
        }
        .chart-container {
            margin: 20px 0;
        }
        iframe {
            width: 100%;
            border: none;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
    </style>
</head>
<body>
    <h1>🔬 Market Simulation Dashboard</h1>
    
    <div class="stats">
        <h2>Summary Statistics</h2>
        <div class="stats-grid">
            <div class="stat-item">
                <div class="stat-label">Total Rounds</div>
                <div class="stat-value">%d</div>
            </div>
            <div class="stat-item">
                <div class="stat-label">Total Generations</div>
                <div class="stat-value">%d</div>
            </div>
            <div class="stat-item">
                <div class="stat-label">Min Price</div>
                <div class="stat-value">$%.2f</div>
            </div>
            <div class="stat-item">
                <div class="stat-label">Max Price</div>
                <div class="stat-value">$%.2f</div>
            </div>
            <div class="stat-item">
                <div class="stat-label">Total Volume</div>
                <div class="stat-value">%d</div>
            </div>
            <div class="stat-item">
                <div class="stat-label">Best Generation</div>
                <div class="stat-value">%d</div>
            </div>
            <div class="stat-item">
                <div class="stat-label">Best Avg Fitness</div>
                <div class="stat-value">%.2fx</div>
            </div>
        </div>
    </div>
    
    <div class="chart-container">
        <iframe src="price_chart.html" height="650"></iframe>
    </div>
    
    <div class="chart-container">
        <iframe src="volume_chart.html" height="650"></iframe>
    </div>
    
    <div class="chart-container">
        <iframe src="fitness_chart.html" height="650"></iframe>
    </div>
    
    <div class="chart-container">
        <iframe src="order_flow_chart.html" height="650"></iframe>
    </div>
</body>
</html>`

	bestGen := grammar.FindBestGeneration(history.Generations)

	content := fmt.Sprintf(html,
		len(history.Timestamps),
		len(history.Generations),
		grammar.MinPrice(history.Prices),
		grammar.MaxPrice(history.Prices),
		grammar.SumVolume(history.Volumes),
		bestGen.Generation,
		bestGen.AvgFitness,
	)

	return os.WriteFile(fmt.Sprintf("%s/dashboard.html", outputDir), []byte(content), 0644)
}
