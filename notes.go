package chords

import (
	"errors"
	"fmt"
)

func TransposePitch(root Pitch, intervals []Interval) []Pitch {
	ret := make([]Pitch, len(intervals))
	for i, interval := range intervals {
		ret[i] = root.Transpose(interval)
	}
	return ret
}

func TransposePitches(pitches []Pitch, interval Interval) []Pitch {
	ret := make([]Pitch, len(pitches))
	for i, pitch := range pitches {
		ret[i] = pitch.Transpose(interval)
	}
	return ret
}

func MeasureIntervals(root Pitch, pitches []Pitch) []Interval {
	// TODO
	return nil
}

func Negate(root Pitch, pitches []Pitch) []Pitch {
	// TODO
	return nil
}

type Note byte

const (
	NOTE_A Note = iota + 'A'
	NOTE_B
	NOTE_C
	NOTE_D
	NOTE_E
	NOTE_F
	NOTE_G
)

var noteCards []byte
func init() {
	noteCards = make([]byte, int(NOTE_G-'A')+1)
	noteCards[int(NOTE_A-'A')] = 0
	noteCards[int(NOTE_B-'A')] = 2
	noteCards[int(NOTE_C-'A')] = 3
	noteCards[int(NOTE_D-'A')] = 5
	noteCards[int(NOTE_E-'A')] = 7
	noteCards[int(NOTE_F-'A')] = 8
	noteCards[int(NOTE_G-'A')] = 10
}

func (n Note) Cardinal() byte {
	return noteCards[int(n-'A')]
}

func (n Note) String() string {
	return string(rune(n))
}

func (n Note) IsValid() bool {
	return n >= NOTE_A && n <= NOTE_G
}

type Pitch struct {
	N   Note
	Acc Accidental
}

func ParsePitch(s string) (Pitch, error) {
	if len(s) == 0 {
		return Pitch{}, errors.New("Cannot parse pitch from empty string")
	}
	n := Note(s[0])
	if !n.IsValid() {
		return Pitch{}, fmt.Errorf("Invalid note %q", n.String())
	}
	if len(s) == 1 {
		return Pitch{N: n}, nil
	}
	a, err := parseAccidental(s[1:])
	if err != nil {
		return Pitch{}, err
	}
	return Pitch{N: n, Acc: a}, nil
}

func MustParsePitch(s string) Pitch {
	p, err := ParsePitch(s)
	if err != nil {
		panic(err)
	}
	if !p.IsValid() {
		panic(fmt.Errorf("ParsePitch unexpectedly returned invalid pitch: %v", p))
	}
	return p
}


func (p Pitch) Cardinal() byte {
	c := int8(p.N.Cardinal()) + p.Acc.Offset()
	if c < 0 {
		c += 12
	} else if c >= 12 {
		c -= 12
	}
	return byte(c)
}

func (p Pitch) Transpose(interval Interval) Pitch {
	// computes modulo, always returning non-negative result
	mod := func(x byte, n byte) byte {
		return (x % n + n) % n;
	}
	np := majorScales[p][mod(interval.Val - 1, 7)]
	o := interval.Offset
	for o != 0 {
		if o >= -4 && o <= 4 {
			np = offsetsByPitch[np][o+4]
			break
		}
		if o < 0 {
			np = offsetsByPitch[np][0]
			o += 4
		} else {
			np = offsetsByPitch[np][4]
			o -= 4
		}
	}
	return np
}

func (p Pitch) adjust(offset int) Pitch {
	// TODO
	return Pitch{}
}

func (p Pitch) String() string {
	str := p.N.String()
	if p.Acc != NATURAL {
		return str + p.Acc.String()
	}
	return str
}

func (p Pitch) IsValid() bool {
	return p.N.IsValid() && p.Acc.IsValid()
}

type Interval struct {
	Val    byte
	Offset int8
}

var stepsByInterval = []int8{0, 2, 4, 5, 7, 9, 11}
var offsetsByPitch_strings = map[string][]string {
	"Ax": {"Abb", "Ab","A", "A#", "Ax", "B#", "Bx", "Cx", "D#"},
	"A#": {"Gb", "Abb", "Ab","A", "A#", "Ax", "B#", "Bx", "Cx"},
	"A": {"Gbb", "Gb", "Abb", "Ab","A", "A#", "Ax", "B#", "Bx"},
	"Ab": {"Fb", "Gbb", "Gb", "Abb","Ab", "A", "A#", "Ax", "B#"},
	"Abb": {"Fbb", "Fb", "Gbb", "Gb", "Abb","Ab", "A", "A#", "Ax"},

	"Bx": {"Bbb", "Bb","B", "B#", "Bx", "Cx", "D#", "Dx", "E#"},
	"B#": {"Ab", "Bbb", "Bb","B", "B#", "Bx", "Cx", "D#", "Dx"},
	"B": {"Abb", "Ab", "Bbb", "Bb","B", "B#", "Bx", "Cx", "D#"},
	"Bb": {"Gb", "Abb", "Ab", "Bbb", "Bb","B", "B#", "Bx", "Cx"},
	"Bbb": {"Gbb", "Gb", "Abb", "Ab", "Bbb", "Bb","B", "B#", "Bx"},

	"Cx": {"Cbb", "Cb","C", "C#", "Cx", "D#", "Dx", "E#", "Ex"},
	"C#": {"Bbb", "Cbb", "Cb","C", "C#", "Cx", "D#", "Dx", "E#"},
	"C": {"Ab", "Bbb", "Cbb", "Cb","C", "C#", "Cx", "D#", "Dx"},
	"Cb": {"Abb", "Ab", "Bbb", "Cbb", "Cb","C", "C#", "Cx", "D#"},
	"Cbb": {"Gb", "Abb", "Ab", "Bbb", "Cbb", "Cb","C", "C#", "Cx"},

	"Dx": {"Dbb", "Db","D", "D#", "Dx", "E#", "Ex", "Fx", "G#"},
	"D#": {"Cb", "Dbb", "Db","D", "D#", "Dx", "E#", "Ex", "Fx"},
	"D": {"Cbb", "Cb", "Dbb", "Db","D", "D#", "Dx", "E#", "Ex"},
	"Db": {"Bbb", "Cbb", "Cb", "Dbb", "Db","D", "D#", "Dx", "E#"},
	"Dbb": {"Ab", "Bbb", "Cbb", "Cb", "Dbb", "Db","D", "D#", "Dx"},

	"Ex": {"Ebb", "Eb","E", "E#", "Ex", "Fx", "G#", "Gx", "A#"},
	"E#": {"Db", "Ebb", "Eb","E", "E#", "Ex", "Fx", "G#", "Gx"},
	"E": {"Dbb", "Db", "Ebb", "Eb","E", "E#", "Ex", "Fx", "G#"},
	"Eb": {"Cb", "Dbb", "Db", "Ebb", "Eb","E", "E#", "Ex", "Fx"},
	"Ebb": {"Cbb", "Cb", "Dbb", "Db", "Ebb", "Eb","E", "E#", "Ex"},

	"Fx": {"Fbb","Fb", "F", "F#", "Fx", "G#", "Gx", "A#", "Ax"},
	"F#": {"Ebb", "Fbb","Fb", "F", "F#", "Fx", "G#", "Gx", "A#"},
	"F": {"Db", "Ebb", "Fbb","Fb", "F", "F#", "Fx", "G#", "Gx"},
	"Fb": {"Dbb", "Db", "Ebb", "Fbb","Fb", "F", "F#", "Fx", "G#"},
	"Fbb": {"Cb", "Dbb", "Db", "Ebb", "Fbb","Fb", "F", "F#", "Fx"},

	"Gx": {"Gbb", "Gb", "G", "G#", "Gx", "A#", "Ax", "B#", "Bx"},
	"G#": {"Fb", "Gbb", "Gb", "G", "G#", "Gx", "A#", "Ax", "B#"},
	"G": {"Fbb", "Fb", "Gbb", "Gb", "G", "G#", "Gx", "A#", "Ax"},
	"Gb": {"Ebb", "Fbb", "Fb", "Gbb", "Gb", "G", "G#", "Gx", "A#"},
	"Gbb": {"Db", "Ebb", "Fbb", "Fb", "Gbb", "Gb", "G", "G#", "Gx"},
}
var offsetsByPitch map[Pitch][]Pitch
var majorScales_strings = map[string][]string {
	"A": {"A", "B", "C#", "D", "E", "F#", "G#"},
	"B": {"B", "C#", "D#", "E", "F#", "G#", "A#"},
	"C": {"C", "D", "E", "F", "G", "A", "B"},
	"D": {"D", "E", "F#", "G", "A", "B", "C#"},
	"E": {"E", "F#", "G#", "A", "B", "C#", "D#"},
	"F": {"F", "G", "A", "Bb", "C", "D", "E"},
	"G": {"G", "A", "B", "C", "D", "E", "F#"},
}
var majorScales map[Pitch][]Pitch

func init() {
	offsetsByPitch = map[Pitch][]Pitch{}
	for k, vs := range offsetsByPitch_strings {
		p := MustParsePitch(k)
		ps := make([]Pitch, len(vs))
		for i, v := range vs {
			ps[i] = MustParsePitch(v)
		}
		offsetsByPitch[p] = ps
	}
	majorScales = map[Pitch][]Pitch{}
	for k, vs := range majorScales_strings {
		p := MustParsePitch(k)
		ps := make([]Pitch, len(vs))
		for i, v := range vs {
			ps[i] = MustParsePitch(v)
		}
		for acc := NATURAL; acc < DBL_SHARP; acc++ {
			if acc.Offset() == 0 {
				majorScales[p] = ps
				continue
			}
			accp := Pitch{N: p.N, Acc: acc}
			accps := make([]Pitch, len(ps))
			for i, pp := range ps {
				accps[i] = offsetsByPitch[pp][acc.Offset() + 4]
			}
			majorScales[accp] = accps
		}
	}
}

func (i Interval) NumHalfSteps() int8 {
	return stepsByInterval[i.Val -1] + i.Offset
}

func (i Interval) IsValid() bool {
	return i.Val >= 1 && i.Val <= 7 && i.Offset >= -4 && i.Offset <= 4
}

type Accidental int

const (
	NATURAL Accidental = iota
	FLAT
	SHARP
	DBL_FLAT
	DBL_SHARP
)

func (a Accidental) String() string {
	switch (a) {
	case NATURAL:
		return "♮"
	case SHARP:
		return "♯"
	case FLAT:
		return "♭"
	case DBL_SHARP:
		return "𝄪"
	case DBL_FLAT:
		return "𝄫"
	default:
		return fmt.Sprintf("?(%d)", a)
	}
}

func (a Accidental) Offset() int8 {
	switch (a) {
	case SHARP:
		return 1
	case FLAT:
		return -1
	case DBL_SHARP:
		return 2
	case DBL_FLAT:
		return -2
	default:
		return 0
	}
}

func (a Accidental) IsValid() bool {
	return a >= NATURAL && a <= DBL_SHARP
}

func parseAccidental(s string) (Accidental, error) {
	switch s {
	case "n", "♮":
		return NATURAL, nil
	case "#", "♯":
		return SHARP, nil
	case "b", "♭":
		return FLAT, nil
	case "x", "𝄪":
		return DBL_SHARP, nil
	case "bb", "𝄫":
		return DBL_FLAT, nil
	default:
		return 0, fmt.Errorf("Invalid accidental: %q", s)
	}
}
