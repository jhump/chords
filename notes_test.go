package chords

import (
	"testing"
)

func TestNoteName(t *testing.T) {
	for i := 0; i < 256; i++ {
		nn := NoteName(i)
		if nn.IsValid() != (i >= 'A' && i <= 'G') {
			t.Errorf("NoteName.IsValid for %s returned wrong value", nn)
		}
	}
	expected := []int8{0, 2, 3, 5, 7, 8, 10}
	actual := []NoteName{A, B, C, D, E, F, G}
	for i, nn := range actual {
		if expected[i] != nn.Cardinal() {
			t.Errorf("NoteName.Cardinal for %s returned wrong value: %d", nn, nn.Cardinal())
		}
	}
}

func TestAccidental(t *testing.T) {
	for i := -128; i < 128; i++ {
		a := Accidental(i)
		if a.IsValid() != (i >= -2 && i <= 2) {
			t.Errorf("Accidental.IsValid for %s returned wrong value", a)
		}
	}
	expected := []int8{0, -1, 1, -2, 2}
	actual := []Accidental{Natural, Flat, Sharp, DblFlat, DblSharp}
	for i, acc := range actual {
		if expected[i] != acc.Offset() {
			t.Errorf("Accidental.Offset for %s returned wrong value: %d", acc, acc.Offset())
		}
	}
}

func TestInterval_IsValid(t *testing.T) {
	for i := -128; i < 128; i++ {
		valValid := i >= 1 && i <= 7
		for j := -128; j < 128; j++ {
			offsValid := j >= -2 && j <= 2
			intv := Interval{Val: int8(i), Offset: int8(j)}
			if intv.IsValid() != (valValid && offsValid) {
				t.Errorf("Interval.IsValid for %v returned wrong value", intv)
			}
		}
	}
}

func TestInterval_NumHalfSteps(t *testing.T) {
	expected := [][]int8{
		{10, 0, 2, 3, 5, 7, 9},
		{11, 1, 3, 4, 6, 8, 10},
		{0, 2, 4, 5, 7, 9, 11},
		{1, 3, 5, 6, 8, 10, 0},
		{2, 4, 6, 7, 9, 11, 1},
	}
	i := 0
	for a := DblFlat; a <= DblSharp; a++ {
		j := 0
		for n := 1; n <= 7; n++ {
			exp := expected[i][j]
			intv := Interval{Val: int8(n), Offset: int8(a)}
			if intv.NumHalfSteps() != exp {
				t.Errorf("Interval.NumHalfSteps for %v returned wrong value: %d", intv, intv.NumHalfSteps())
			}
			j++
		}
		i++
	}
}

func TestNote_IsValid(t *testing.T) {
	for i := 0; i < 256; i++ {
		nn := NoteName(i)
		for i := -128; i < 128; i++ {
			a := Accidental(i)
			n := Note{N: nn, Acc: a}
			if n.IsValid() != (nn.IsValid() && a.IsValid()) {
				t.Errorf("Note.IsValid for %s returned wrong value", n)
			}
		}
	}
}

func TestNote_Cardinal(t *testing.T) {
	expected := [][]int8{
		{10, 0, 1, 3, 5, 6, 8},
		{11, 1, 2, 4, 6, 7, 9},
		{0, 2, 3, 5, 7, 8, 10},
		{1, 3, 4, 6, 8, 9, 11},
		{2, 4, 5, 7, 9, 10, 0},
	}
	i := 0
	for a := DblFlat; a <= DblSharp; a++ {
		j := 0
		for n := A; n <= G; n++ {
			exp := expected[i][j]
			n := Note{N: n, Acc: a}
			if n.Cardinal() != exp {
				t.Errorf("Note.Cardinal for %s returned wrong value: %d", n, n.Cardinal())
			}
			j++
		}
		i++
	}
}

func TestNote_IntervalTo(t *testing.T) {
	// TODO
}

func TestNote_Transpose(t *testing.T) {
	// TODO
}

func TestParseNote(t *testing.T) {
	cases := []struct {
		acc rune
		exp Accidental
	}{
		{'n', Natural},
		{'#', Sharp},
		{'b', Flat},
	}
	for n := A; n <= G; n++ {
		for i, tc := range cases {
			_, _ = i, tc
			// TODO
		}
	}
}
