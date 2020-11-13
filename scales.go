package chords

import (
	"fmt"
	"sort"
)

// ScaleType represents a sequence of notes relative to a scale root.
//
// Most scales are "heptatonic", which means they contain seven notes, one
// for each note name A through G.
type ScaleType []Interval

// IsValid returns true if all intervals in the scale are valid and if the
// scale contains a root tone, which is an interval {Val: 1, Offset: 0}.
func (t ScaleType) IsValid() bool {
	hasTonic := false
	for _, intv := range t {
		if !intv.IsValid() {
			return false
		}
		if intv.Val == 1 && intv.Offset == 0 {
			hasTonic = true
		}
	}
	return hasTonic
}

// Clean returns a new ScaleType that is sorted and has redundant (enharmonic
// equivalent) scale degrees removed.
func (t ScaleType) Clean() ScaleType {
	less := func(i, j int) bool {
		si := t[i].NumHalfSteps()
		sj := t[j].NumHalfSteps()
		if si == sj {
			return t[i].Val < t[j].Val
		}
		return si < sj
	}
	if !sort.SliceIsSorted(t, less) {
		clone := make(ScaleType, len(t))
		copy(clone, t)
		t = clone
		sort.Slice(t, less)
	}
	// now remove any redundant scale tones
	found := 0
	for i := range t {
		if i == 0 {
			continue
		}
		if t[i].NumHalfSteps() == t[i-1].NumHalfSteps() {
			found = i
			break
		}
	}
	if found == 0 {
		return t
	}
	newSlice := make(ScaleType, found, len(t))
	copy(newSlice, t)
	for i := found + 1; i < len(t); i++ {
		if t[i].NumHalfSteps() != t[i-1].NumHalfSteps() {
			newSlice = append(newSlice, t[i])
		}
	}
	return newSlice
}

// WithRoot creates a scale with the given root and this scale type.
func (t ScaleType) WithRoot(root Note) *Scale {
	return &Scale{
		Root: root,
		Type: t,
	}
}

// NthMode returns the nth mode of the t. It may return unexpected results for
// non-heptatonic scales since it is a simple pivot to the nth scale degree. So
// the second mode of a minor pentatonic or blues scale would actually start with
// the minor third, not a second, since that is the second note in the scale.
func (t ScaleType) NthMode(n int8) ScaleType {
	if n < 1 {
		panic(fmt.Sprintf("NthMode requires n >= 1, got %d", n))
	} else if int(n) > len(t) {
		panic(fmt.Sprintf("NthMode where n (%d) > than scale length (%d)", n, len(t)))
	}

	t = t.Clean()
	if n == 1 {
		return t
	}

	steps := make([]int8, len(t))
	for i := range t {
		steps[i] = t[i].NumHalfSteps()
	}
	modeStartDelta := t[n-1].Val - 1
	modeStartSteps := steps[n-1]

	tailSteps := append([]int8{}, steps[:n-1]...)
	steps = append(steps[n-1:], tailSteps...)

	intvs := make(ScaleType, len(t))
	copy(intvs, t)
	tailIntvs := append(ScaleType{}, intvs[:n-1]...)
	intvs = append(intvs[n-1:], tailIntvs...)

	for i := range intvs {
		intvSteps := posMod(intvs[i].NumHalfSteps()-modeStartSteps, 12)
		intvs[i].Val -= modeStartDelta
		if intvs[i].Val < 1 {
			intvs[i].Val += 7
		}
		intvs[i].Offset = 0
		delta := intvSteps - intvs[i].NumHalfSteps()
		for delta < -2 || delta > 2 {
			if delta < 0 {
				intvs[i].Val--
				if intvs[i].Val < 1 {
					intvs[i].Val += 7
				}
			} else {
				intvs[i].Val++
				if intvs[i].Val > 7 {
					intvs[i].Val -= 7
				}
			}
			delta = intvSteps - intvs[i].NumHalfSteps()
		}
		intvs[i].Offset = delta
	}

	return intvs
}

// HeptatonicScaleType is a factory function for creating heptatonic scale
// types from 7 integer offsets. Offsets of zero map to the major scale. So
// if the value in the 3rd element (index 2) is -1, the scale type will have
// a minor third.
func HeptatonicScaleType(offsets [7]int8) ScaleType {
	intvs := make([]Interval, 7)
	for idx, offs := range offsets {
		intvs[idx] = Interval{
			Val:    int8(idx + 1),
			Offset: offs,
		}
	}
	return intvs
}

var (
	// Diatonic scales/modes

	IonianMode     = HeptatonicScaleType([7]int8{0, 0, 0, 0, 0, 0, 0})
	MajorScale     = IonianMode
	DorianMode     = MajorScale.NthMode(2)
	PhrygianMode   = MajorScale.NthMode(3)
	LydianMode     = MajorScale.NthMode(4)
	MixolydianMode = MajorScale.NthMode(5)
	AeolianMode    = MajorScale.NthMode(6)
	MinorScale     = AeolianMode
	LocrianMode    = MajorScale.NthMode(7)

	// Other heptatonic scales

	HarmonicMinorScale       = HeptatonicScaleType([7]int8{0, 0, -1, 0, 0, -1, 0})
	MelodicMinorScale        = HeptatonicScaleType([7]int8{0, 0, -1, 0, 0, 0, 0})
	HungarianMinorScale      = HeptatonicScaleType([7]int8{0, 0, -1, 1, 0, -1, 0})
	DoubleHarmonicMinorScale = HungarianMinorScale

	// Non-heptatonic scales

	HalfWholeScale = ScaleType{
		{Val: 1, Offset: 0}, {Val: 2, Offset: -1}, {Val: 3, Offset: -1},
		{Val: 3, Offset: 0}, {Val: 4, Offset: 1}, {Val: 5, Offset: 0},
		{Val: 6, Offset: 0}, {Val: 7, Offset: -1},
	}
	WholeHalfScale = ScaleType{
		{Val: 1, Offset: 0}, {Val: 2, Offset: 0}, {Val: 3, Offset: -1},
		{Val: 4, Offset: 0}, {Val: 5, Offset: -1}, {Val: 6, Offset: -1},
		{Val: 6, Offset: 0}, {Val: 7, Offset: 0},
	}
	DiminishedScale = WholeHalfScale

	WholeToneScale = ScaleType{
		{Val: 1, Offset: 0}, {Val: 2, Offset: 0}, {Val: 3, Offset: 0},
		{Val: 4, Offset: 1}, {Val: 5, Offset: 1}, {Val: 7, Offset: -1},
	}

	PentatonicMajorScale = ScaleType{
		{Val: 1, Offset: 0}, {Val: 2, Offset: 0}, {Val: 3, Offset: 0},
		{Val: 5, Offset: 0}, {Val: 6, Offset: 0},
	}
	PentatonicMinorScale = ScaleType{
		{Val: 1, Offset: 0}, {Val: 3, Offset: -1}, {Val: 4, Offset: 0},
		{Val: 5, Offset: 0}, {Val: 7, Offset: -1},
	}
	BluesScale = ScaleType{
		{Val: 1, Offset: 0}, {Val: 3, Offset: -1}, {Val: 4, Offset: 0},
		{Val: 5, Offset: -1}, {Val: 5, Offset: 0}, {Val: 7, Offset: -1},
	}

	ChromaticScale = ScaleType{
		{Val: 1, Offset: 0}, {Val: 2, Offset: -1}, {Val: 2, Offset: 0},
		{Val: 3, Offset: -1}, {Val: 3, Offset: 0}, {Val: 4, Offset: 0},
		{Val: 4, Offset: 1}, {Val: 5, Offset: 0}, {Val: 6, Offset: -1},
		{Val: 6, Offset: 0}, {Val: 7, Offset: -1}, {Val: 7, Offset: 0},
	}
)

// Scale represents a scale, which is a set of notes. It is described by
// a root note and a scale type.
type Scale struct {
	Root Note
	Type ScaleType
}

// IsValid returns true if the scale's Root and Type are valid.
func (s *Scale) IsValid() bool {
	return s.Root.IsValid() && s.Type.IsValid()
}

// Clean ensures s.Type is clean. See ScaleType.Clean.
func (s *Scale) Clean() {
	s.Type = s.Type.Clean()
}

// Spell returns the notes in the scale.
func (s *Scale) Spell() []Note {
	notes := make([]Note, len(s.Type))
	for i, intv := range s.Type {
		notes[i] = s.Root.Transpose(intv)
	}
	return notes
}
