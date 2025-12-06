package grammar

import (
	"fmt"
	"testing"
)

func TestCalculateNewPrice(t *testing.T) {

	p := calculateNewPrice(100.0, []Order{
		{
			1,
			"BUY",
			20,
		},
	})

	fmt.Println(p)
}
