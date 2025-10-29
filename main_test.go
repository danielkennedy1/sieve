package main

import (
	"testing"
	"slices"
)

func TestSinglePointCrossover(t *testing.T) {
	tests := []struct {
		a, b []bool
		point int
		want_a, want_b []bool
	}{
		{[]bool{true, true}, []bool{false, false}, 0, []bool{true, true}, []bool{false, false}},
		{[]bool{true, true}, []bool{false, false}, 1, []bool{true, false}, []bool{false, true}},
		{[]bool{true, true}, []bool{false, false}, 2, []bool{true, true}, []bool{false, false}},
	}

	for _, in := range tests {
		got_a, got_b := SinglePointCrossover(in.a, in.b, in.point)

		if !slices.Equal(got_a, in.want_a) || !slices.Equal(got_b, in.want_b) {
			t.Errorf("got %v, %v. want %v, %v", got_a, got_b, in.want_a, in.want_b)
		}

	}

}

func TestMutate(t *testing.T) {
	want := []bool{false, true}
	got := []bool{false, false}
	Mutate(got, 1)

	if !slices.Equal(want, got) {
		t.Errorf("got %v, want %v", got, want)
	}
}
