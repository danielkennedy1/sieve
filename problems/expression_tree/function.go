package expression_tree

import (
	"encoding/csv"
	"fmt"
	"math"
	"strconv"

	"github.com/danielkennedy1/sieve/genomes"
)

type Sample struct {
	Variables []float64 // [x0, x1, x2...]
	Output    float64   // f(x)
}

type Function struct {
	Samples   []Sample  // Samples for this function
	Constants []float64 // Constants to be used to fit this function
}

func LoadSamples(reader *csv.Reader) ([]Sample, error) {
	// Read header to determine column count
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("reading header: %w", err)
	}

	// Validate header format: y, x0, x1, ..., xn-1
	if len(header) < 2 || header[0] != "y" {
		return nil, fmt.Errorf("invalid header: expected 'y' as first column")
	}

	numVars := len(header) - 1

	var samples []Sample
	for {
		row, err := reader.Read()
		if err != nil {
			break // io.EOF or actual error
		}

		if len(row) != len(header) {
			return nil, fmt.Errorf("row length mismatch: expected %d columns", len(header))
		}

		// Parse y value
		y, err := strconv.ParseFloat(row[0], 64)
		if err != nil {
			return nil, fmt.Errorf("parsing y value: %w", err)
		}

		// Parse x values
		vars := make([]float64, numVars)
		for i := range numVars {
			vars[i], err = strconv.ParseFloat(row[i+1], 64)
			if err != nil {
				return nil, fmt.Errorf("parsing x%d value: %w", i, err)
			}
		}

		samples = append(samples, Sample{
			Variables: vars,
			Output:    y,
		})
	}

	if len(samples) == 0 {
		return nil, fmt.Errorf("no samples found in csv")
	}

	return samples, nil
}

//func RootMeanSquaredError(et genomes.Expression, variables *[]float64, samples *[]Sample) float64 {
//	total_squared_error := 0.0
//
//	for i := range *samples {
//		(*variables) = (*samples)[i].Variables
//		squared_error := math.Pow((et.GetValue() - (*samples)[i].Output), 2)
//		total_squared_error += squared_error
//	}
//
//	mean_squared_error := total_squared_error / float64(len(*samples))
//
//	return math.Sqrt(mean_squared_error)
//}

func NewRootMeanSquaredError(variables *[]float64, samples *[]Sample) func(e genomes.Expression) float64 {
	return func(e genomes.Expression) float64 {
		total_squared_error := 0.0

		for i := range *samples {
			(*variables) = (*samples)[i].Variables
			squared_error := math.Pow((e.GetValue() - (*samples)[i].Output), 2)
			total_squared_error += squared_error
		}

		mean_squared_error := total_squared_error / float64(len(*samples))

		return -math.Sqrt(mean_squared_error)
	}
}
