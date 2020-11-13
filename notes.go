package chords

import (
	"errors"
	"fmt"
)

// TransposeNote returns a new set of notes that are the given root note
// transposed by the given intervals. The first returned note corresponds
// to the first given interval, and so on.
func TransposeNote(root Note, intervals ...Interval) []Note {
	ret := make([]Note, len(intervals))
	for i, interval := range intervals {
		ret[i] = root.Transpose(interval)
	}
	return ret
}

// TransposeNotes returns a new set of notes that are the given set of notes
// transposed by the given interval. The first returned node corresponds to
// the transposition of the first given note, and so on.
func TransposeNotes(notes []Note, interval Interval) []Note {
	ret := make([]Note, len(notes))
	for i, pitch := range notes {
		ret[i] = pitch.Transpose(interval)
	}
	return ret
}

// MeasureIntervals returns the intervals reflecting the distance of the given
// notes from the given root note. The first returned interval represents the
// distance between the root node and the first given note, and so on.
func MeasureIntervals(root Note, notes ...Note) []Interval {
	intvs := make([]Interval, len(notes))
	for i, n := range notes {
		intvs[i] = root.IntervalTo(n)
	}
	return intvs
}

// Negate returns a new set of notes that correspond to the "negation" of the
// given notes around the given root. A "negated" note is the reflection of that
// note around a given root. For example, if a given note is 3 half-steps higher
// than the root, its negated note is 3 half-steps below (or 9 half-steps higher)
// than the root. This can be used to shift notes and chords into
// "negative harmony".
func Negate(root Note, notes ...Note) []Note {
	neg := make([]Note, len(notes))
	for i, n := range notes {
		intv := root.IntervalTo(n)
		dist := 12 - intv.NumHalfSteps()
		if dist == 12 {
			neg[i] = n
			continue
		}
		negIntv := Interval{Val: 8 - intv.Val}
		offs := dist - negIntv.NumHalfSteps()
		for offs < -2 {
			negIntv.Val--
			if negIntv.Val < 1 {
				negIntv.Val += 7
			}
			offs = dist - negIntv.NumHalfSteps()
		}
		for offs > 2 {
			negIntv.Val++
			if negIntv.Val > 7 {
				negIntv.Val -= 7
			}
			offs = dist - negIntv.NumHalfSteps()
		}
		negIntv.Offset = offs
		neg[i] = root.Transpose(negIntv)
	}
	return neg
}

// NoteName is the single-letter name of a note. The A note is represented by
// the letter 'A', and so on up to 'G'. (The underlying type is byte instead
// of rune because all valid note names fall in the 7-bit ASCII range).
type NoteName byte

const (
	// Corresponds to the note named A.
	A NoteName = iota + 'A'
	// Corresponds to the note named B.
	B
	// Corresponds to the note named C.
	C
	// Corresponds to the note named D.
	D
	// Corresponds to the note named E.
	E
	// Corresponds to the note named F.
	F
	// Corresponds to the note named G.
	G
)

var noteCardinals []int8

func init() {
	noteCardinals = make([]int8, int(G-'A')+1)
	noteCardinals[int(A-'A')] = 0
	noteCardinals[int(B-'A')] = 2
	noteCardinals[int(C-'A')] = 3
	noteCardinals[int(D-'A')] = 5
	noteCardinals[int(E-'A')] = 7
	noteCardinals[int(F-'A')] = 8
	noteCardinals[int(G-'A')] = 10
}

// Cardinal returns the cardinality of this note name, as measured in half-step
// intervals from A. So A has a cardinality of zero. B, which is a whole-step
// higher has a cardinality of 2, C has a cardinality of 3, G has a cardinality
// of 10, etc.
func (n NoteName) Cardinal() int8 {
	return noteCardinals[int(n-'A')]
}

// String implements the Stringer interface.
func (n NoteName) String() string {
	return string(rune(n))
}

// IsValid returns true only if this note name is valid. A valid note name is
// between 'A' and 'G'; all others are invalid.
func (n NoteName) IsValid() bool {
	return n >= A && n <= G
}

// Note represents a note in modern diatonic music. It is a note name combined
// with an accidental. For example, the note A is the note name 'A' with an
// accidental NATURAL. The note F# is the note name 'F' with an accidental SHARP.
type Note struct {
	N   NoteName
	Acc Accidental
}

// ParseNote parses a note from the given string. For example "Bb" will return the
// note with name 'B' and accidental FLAT. "Cx" will return a note with name "C"
// and accidental DBL_SHARP. It returns an error if the string cannot be parsed
// into a note.
func ParseNote(s string) (Note, error) {
	if len(s) == 0 {
		return Note{}, errors.New("cannot parse note from empty string")
	}
	n := NoteName(s[0])
	if !n.IsValid() {
		return Note{}, fmt.Errorf("invalid note name %q", n.String())
	}
	if len(s) == 1 {
		return Note{N: n}, nil
	}
	a, err := parseAccidental(s[1:])
	if err != nil {
		return Note{}, err
	}
	return Note{N: n, Acc: a}, nil
}

// MustParseNote parses the given string into a note and panics if the string is
// not valid. (See ParseNote.)
func MustParseNote(s string) Note {
	p, err := ParseNote(s)
	if err != nil {
		panic(err)
	}
	if !p.IsValid() {
		panic(fmt.Errorf("ParseNote unexpectedly returned invalid pitch: %v", p))
	}
	return p
}

// Cardinal returns the cardinality of this note, as measured in half-step
// intervals from A. An A (natural) has cardinality zero. Bb has a cardinality
// of 1. The returned cardinality will always be between 0 and 11, representing
// the 12 pitches in an octave of modern diatonic music. Notes that fall outside
// this range will wrap. For example, Ab has cardinality of 11, not -1.
func (n Note) Cardinal() int8 {
	return posMod(int8(n.N.Cardinal())+n.Acc.Offset(), 12)
}

// Transpose returns the note that results from transposing this note by the
// given interval.
func (n Note) Transpose(interval Interval) Note {
	np := majorScales[n][posMod(int8(interval.Val)-1, 7)]
	o := interval.Offset
	for o != 0 {
		if o >= -4 && o <= 4 {
			np = offsetsByNote[np][o+4]
			break
		}
		if o < 0 {
			np = offsetsByNote[np][0]
			o += 4
		} else {
			np = offsetsByNote[np][4]
			o -= 4
		}
	}
	return np
}

func (n Note) IntervalTo(other Note) Interval {
	var intv Interval
	intv.Val = posMod(int8(other.N-'a')-int8(n.N-'a'), 8)
	dHalfSteps := posMod(int8(other.Cardinal()-n.Cardinal()), 12)
	offs := dHalfSteps - intv.NumHalfSteps()
	for offs < -2 {
		intv.Val--
		if intv.Val < 1 {
			intv.Val += 7
		}
		offs = dHalfSteps - intv.NumHalfSteps()
	}
	for offs > 2 {
		intv.Val++
		if intv.Val > 7 {
			intv.Val -= 7
		}
		offs = dHalfSteps - intv.NumHalfSteps()
	}
	intv.Offset = offs
	return intv
}

// posMod computes modulo, but always returning non-negative result
func posMod(x int8, n int8) int8 {
	return (x%n + n) % n
}

// String implements the Stringer interface.
func (n Note) String() string {
	str := n.N.String()
	if n.Acc != Natural {
		return str + n.Acc.String()
	}
	return str
}

// IsValid returns true only if this note is valid. A valid note has a valid
// note name and valid accidental.
func (n Note) IsValid() bool {
	return n.N.IsValid() && n.Acc.IsValid()
}

// Interval represents the distance between two notes. This is measured in scale
// steps, not half-steps. Examples of intervals are "second", "minor third",
// "sharp fourth", "minor sixth", etc.
//
// The Val will be a number between 1 and 7 inclusive. An interval of 1 with a
// zero Offset indicates the root tone (a distance of zero half-steps).
//
// For intervals with a "minor" or "flat" qualifier in their name, the Offset
// value would be -1. For intervals with a "sharp" qualifier in their name, the
// Offset value would be 1. And for other intervals (no qualifier or "major"
// qualifier), the Offset would be zero.
//
// The following table enumerates the typical values for an Interval. It is also
// possible to have other valid intervals, such as third with an offset of -2
// (which represents only two half steps), as might be the interval between A#
// and Cb.
//  +-----------------------+--------------------+-----------------+
//  | Name                  | Value              | Num. Half Steps |
//  +-----------------------+--------------------+-----------------+
//  | Tonic                 | Val: 1  Offset: 0  |              0  |
//  | Minor/Flat Second     | Val: 2  Offset: -1 |              1  |
//  | Major Second          | Val: 2  Offset: 0  |              2  |
//  | Minor Third           | Val: 3  Offset: -1 |              3  |
//  | Major Third           | Val: 3  Offset: 0  |              4  |
//  | Perfect Fourth        | Val: 4  Offset: 0  |              5  |
//  | Sharp Fourth          | Val: 4  Offset: 1  |              6  |
//  | Flat/Diminished Fifth | Val: 5  Offset: -1 |              6  |
//  | Perfect Fifth         | Val: 5  Offset: 0  |              7  |
//  | Sharp/Augmented Fifth | Val: 5  Offset: 1  |              8  |
//  | Minor/Flat Sixth      | Val: 6  Offset: -1 |              8  |
//  | Major Sixth           | Val: 6  Offset: 0  |              9  |
//  | Minor Seventh         | Val: 7  Offset: -1 |             10  |
//  | Major Seventh         | Val: 7  Offset: 0  |             11  |
//  +-----------------------+--------------------+-----------------+
// The number of half steps are cyclic every octave (which is 12 half steps).
// So the tonic interval is 0 half steps in distance in the same octave. But the
// tonic in the next octave is 12 half steps away (and 24 for the one after that
// and 36 and so on).
type Interval struct {
	Val    int8
	Offset int8
}

var stepsByInterval = []int8{0, 2, 4, 5, 7, 9, 11}
var offsetsByNote_strings = map[string][]string{
	"Ax":  {"Abb", "Ab", "A", "A#", "Ax", "B#", "Bx", "Cx", "D#"},
	"A#":  {"Gb", "Abb", "Ab", "A", "A#", "Ax", "B#", "Bx", "Cx"},
	"A":   {"Gbb", "Gb", "Abb", "Ab", "A", "A#", "Ax", "B#", "Bx"},
	"Ab":  {"Fb", "Gbb", "Gb", "Abb", "Ab", "A", "A#", "Ax", "B#"},
	"Abb": {"Fbb", "Fb", "Gbb", "Gb", "Abb", "Ab", "A", "A#", "Ax"},

	"Bx":  {"Bbb", "Bb", "B", "B#", "Bx", "Cx", "D#", "Dx", "E#"},
	"B#":  {"Ab", "Bbb", "Bb", "B", "B#", "Bx", "Cx", "D#", "Dx"},
	"B":   {"Abb", "Ab", "Bbb", "Bb", "B", "B#", "Bx", "Cx", "D#"},
	"Bb":  {"Gb", "Abb", "Ab", "Bbb", "Bb", "B", "B#", "Bx", "Cx"},
	"Bbb": {"Gbb", "Gb", "Abb", "Ab", "Bbb", "Bb", "B", "B#", "Bx"},

	"Cx":  {"Cbb", "Cb", "C", "C#", "Cx", "D#", "Dx", "E#", "Ex"},
	"C#":  {"Bbb", "Cbb", "Cb", "C", "C#", "Cx", "D#", "Dx", "E#"},
	"C":   {"Ab", "Bbb", "Cbb", "Cb", "C", "C#", "Cx", "D#", "Dx"},
	"Cb":  {"Abb", "Ab", "Bbb", "Cbb", "Cb", "C", "C#", "Cx", "D#"},
	"Cbb": {"Gb", "Abb", "Ab", "Bbb", "Cbb", "Cb", "C", "C#", "Cx"},

	"Dx":  {"Dbb", "Db", "D", "D#", "Dx", "E#", "Ex", "Fx", "G#"},
	"D#":  {"Cb", "Dbb", "Db", "D", "D#", "Dx", "E#", "Ex", "Fx"},
	"D":   {"Cbb", "Cb", "Dbb", "Db", "D", "D#", "Dx", "E#", "Ex"},
	"Db":  {"Bbb", "Cbb", "Cb", "Dbb", "Db", "D", "D#", "Dx", "E#"},
	"Dbb": {"Ab", "Bbb", "Cbb", "Cb", "Dbb", "Db", "D", "D#", "Dx"},

	"Ex":  {"Ebb", "Eb", "E", "E#", "Ex", "Fx", "G#", "Gx", "A#"},
	"E#":  {"Db", "Ebb", "Eb", "E", "E#", "Ex", "Fx", "G#", "Gx"},
	"E":   {"Dbb", "Db", "Ebb", "Eb", "E", "E#", "Ex", "Fx", "G#"},
	"Eb":  {"Cb", "Dbb", "Db", "Ebb", "Eb", "E", "E#", "Ex", "Fx"},
	"Ebb": {"Cbb", "Cb", "Dbb", "Db", "Ebb", "Eb", "E", "E#", "Ex"},

	"Fx":  {"Fbb", "Fb", "F", "F#", "Fx", "G#", "Gx", "A#", "Ax"},
	"F#":  {"Ebb", "Fbb", "Fb", "F", "F#", "Fx", "G#", "Gx", "A#"},
	"F":   {"Db", "Ebb", "Fbb", "Fb", "F", "F#", "Fx", "G#", "Gx"},
	"Fb":  {"Dbb", "Db", "Ebb", "Fbb", "Fb", "F", "F#", "Fx", "G#"},
	"Fbb": {"Cb", "Dbb", "Db", "Ebb", "Fbb", "Fb", "F", "F#", "Fx"},

	"Gx":  {"Gbb", "Gb", "G", "G#", "Gx", "A#", "Ax", "B#", "Bx"},
	"G#":  {"Fb", "Gbb", "Gb", "G", "G#", "Gx", "A#", "Ax", "B#"},
	"G":   {"Fbb", "Fb", "Gbb", "Gb", "G", "G#", "Gx", "A#", "Ax"},
	"Gb":  {"Ebb", "Fbb", "Fb", "Gbb", "Gb", "G", "G#", "Gx", "A#"},
	"Gbb": {"Db", "Ebb", "Fbb", "Fb", "Gbb", "Gb", "G", "G#", "Gx"},
}
var offsetsByNote map[Note][]Note
var majorScales_strings = map[string][]string{
	"A": {"A", "B", "C#", "D", "E", "F#", "G#"},
	"B": {"B", "C#", "D#", "E", "F#", "G#", "A#"},
	"C": {"C", "D", "E", "F", "G", "A", "B"},
	"D": {"D", "E", "F#", "G", "A", "B", "C#"},
	"E": {"E", "F#", "G#", "A", "B", "C#", "D#"},
	"F": {"F", "G", "A", "Bb", "C", "D", "E"},
	"G": {"G", "A", "B", "C", "D", "E", "F#"},
}
var majorScales map[Note][]Note

func init() {
	offsetsByNote = map[Note][]Note{}
	for k, vs := range offsetsByNote_strings {
		n := MustParseNote(k)
		ns := make([]Note, len(vs))
		for i, v := range vs {
			ns[i] = MustParseNote(v)
			// sanity check that the table has valid data
			if int8(ns[i].Cardinal()) != posMod(int8(n.Cardinal())+int8(i)-4, 12) {
				panic(fmt.Errorf("offset table is incorrect! %v (%d) offset by %d (%d) != %v (%d)",
					n, n.Cardinal(), i-4, posMod(int8(n.Cardinal())+int8(i)-4, 12), ns[i], ns[i].Cardinal()))
			}
		}
		offsetsByNote[n] = ns
	}
	majorScales = map[Note][]Note{}
	for k, vs := range majorScales_strings {
		n := MustParseNote(k)
		ns := make([]Note, len(vs))
		for i, v := range vs {
			ns[i] = MustParseNote(v)
		}
		for acc := Natural; acc < DblSharp; acc++ {
			if acc.Offset() == 0 {
				majorScales[n] = ns
				continue
			}
			accn := Note{N: n.N, Acc: acc}
			accns := make([]Note, len(ns))
			for i, pp := range ns {
				accns[i] = offsetsByNote[pp][acc.Offset()+4]
			}
			majorScales[accn] = accns
		}
	}
}

// NumHalfSteps returns the distance, in half-steps, that this interval
// represents. The value returned by a valid octave is the distance within a
// single octave, so it is always positive and between 0 and 11.
func (i Interval) NumHalfSteps() int8 {
	return posMod(stepsByInterval[i.Val-1]+i.Offset, 12)
}

// IsValid returns true if the interval is valid. The interval is valid if its
// Val is between 1 and 7 (inclusive) and its Offset is between -2 and 2 (which
// correspond to the extents of a double-flat or double-sharp qualifier).
func (i Interval) IsValid() bool {
	return i.Val >= 1 && i.Val <= 7 && i.Offset >= -2 && i.Offset <= 2
}

// Accidental describes a note modifier. An unmodified note is a "natural" note,
// which means no accidental. The others are standard symbols used in music
// notation to indicate pitches that fall outside a key signature and to
// describe notes on the black keys of a piano/keyboard (natural notes are the
// white keys).
type Accidental int8

const (
	// Natural is the "no accidental" accidental.
	Natural Accidental = 0
	// Flat corresponds to a flat symbol, ♭, in music notation. It means the note
	// is modified by moving it one half-step lower.
	Flat Accidental = -1
	// Sharp corresponds to a sharp symbol, ♯, in music notation. It means the
	// note is modified by moving it one half-step higher.
	Sharp Accidental = 1
	// DblFlat corresponds to a double-flat symbol, 𝄫, in music notation. It
	// means the note is modified by moving it two half-steps lower.
	DblFlat Accidental = -2
	// DblSharp corresponds to a double-sharp symbol, 𝄪, in music notation. It
	// means the note is modified by moving it two half-steps higher.
	DblSharp Accidental = 2
)

// String implements the Stringer interface.
func (a Accidental) String() string {
	switch a {
	case Natural:
		return "♮"
	case Sharp:
		return "♯"
	case Flat:
		return "♭"
	case DblSharp:
		return "𝄪"
	case DblFlat:
		return "𝄫"
	default:
		return fmt.Sprintf("?(%d)", a)
	}
}

// Offset returns the number of half-steps by which this accidental moves a
// note.
func (a Accidental) Offset() int8 {
	return int8(a)
}

// IsValid returns true if the accidental value is valid. It is valid if it
// represents a known accidental: DBL_FLAT, FLAT, NATURAL, SHARP, or DBL_SHARP.
func (a Accidental) IsValid() bool {
	return a >= DblFlat && a <= DblSharp
}

func parseAccidental(s string) (Accidental, error) {
	switch s {
	case "n", "♮":
		return Natural, nil
	case "#", "♯":
		return Sharp, nil
	case "b", "♭":
		return Flat, nil
	case "x", "𝄪":
		return DblSharp, nil
	case "bb", "𝄫":
		return DblFlat, nil
	default:
		return 0, fmt.Errorf("invalid accidental: %q", s)
	}
}
