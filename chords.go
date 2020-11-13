package chords

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"sort"
)

//go:generate goyacc -o chordparse.y.go -p chord chordparse.y

func init() {
	chordErrorVerbose = true

	// fix up the generated "token name" array so that error messages are nicer
	setTokenName(_SYM_NOTE, "note name ('A'-'G')")
	setTokenName(_SYM_TONE, "chord tone ('2'-'7')")
	setTokenName(_SYM_MAJ7, "'△', '∆', or 'maj'")
	setTokenName(_SYM_SUS, "'sus'")
	setTokenName(_SYM_ACCIDENTAL, "accidental ('n', '♮', 'b', '♭', 'bb', '𝄫', '#', '♯', 'x', or '𝄪')")
	setTokenName(_SYM_MIN, "'min'")
	setTokenName(_SYM_DIM, "'dim'")
	setTokenName(_SYM_HDIM, "'ø'")
	setTokenName(_SYM_FDIM, "'o'")
	setTokenName(_SYM_AUG, "'aug'")
}

func setTokenName(token int, text string) {
	// NB: this is based on logic in generated parse code that translates the
	// int returned from the lexer into an internal token number.
	var intern int
	if token < len(chordTok1) {
		intern = chordTok1[token]
	} else {
		if token >= chordPrivate {
			if token < chordPrivate+len(chordTok2) {
				intern = chordTok2[token-chordPrivate]
			}
		}
		if intern == 0 {
			for i := 0; i+1 < len(chordTok3); i += 2 {
				if chordTok3[i] == token {
					intern = chordTok3[i+1]
					break
				}
			}
		}
	}

	if intern >= 1 && intern-1 < len(chordToknames) {
		chordToknames[intern-1] = text
		return
	}

	panic(fmt.Sprintf("Unknown token value: %d", token))
}

// Chord represents a polyphonic sound in music. This structure can represent
// simple triads to 4- and 5-note chords, like found in jazz music. The basic
// chord is a root tone accompanied by its 3rd and 5th tone. For example, a C
// major chord consists of C. Its 3rd tone is E (C is considered the first, so
// D is the second, and so on). And its 5th tone is G.
type Chord struct {
	// Root is the root note of the chord. For example, the root note for
	// all three of C major, C minor, and C dominant flat 13 (C7♭13) is C.
	// The other tones in the chord are all relative to the root.
	Root Note
	// Triad indicates the basic "shape" of the chord, which is the distance
	// between the primary three tones of the chord. A special triad type,
	// SUS, indicates a chord where the 3rd tone is omitted. In such a case,
	// ExtraTones should contain the suspension note (the 2nd or 4th tone
	// from the root).
	Triad TriadType
	// The additional tones in the chord, other than the root, third, and
	// fifth.
	ExtraTones []ChordTone
	// Bass, if present, is the lowest pitch in the chord. It is considered
	// not present if this field has a zero value (more specifically, if Bass.N
	// is zero, which isn't valid). Bass tones usually indicate a chord
	// inversion. For example C/E indicates a C major chord with an E in the
	// bass. The usual notes of a C major chord are C, E, and G. So the notes in
	// C/E are re-ordered: E, G, C.
	Bass Note
	// canonical is true if Canonicalize has been called to ensure this
	// chord is a canonical form.
	canonical bool
}

// ParseChord parses the given string into a chord. The way the string is
// parsed should be intuitive for those familiar with reading chord names in
// music.
//
// Valid strings must first indicate their root tone as: A-G (must be capital)
// followed by an optional 'n', '♮' (natural), '#', '♯' (sharp), 'b', '♭'
// (flat), 'x', '𝄪' (double-sharp), 'bb', or '𝄫' (double-flat).
//
// The root tone may be followed by a triad indicator (major if omitted): '-',
// 'm', or 'min' for minor; 'aug' or '+' for augmented; 'dim' for diminished
// (fully diminished if 7th is present); 'ø' for half-diminished (implies the
// 7th); or 'o' for fully diminished (implies a flat 7th, aka minor 6th/13th).
//
// A 'sus' can be used in place of a triad indicator to mean that the 3rd is
// omitted. The 'sus' is followed by a '2' or '4', with an optional sharp (for 4)
// or flat (for 2) modifier in between, to indicate which note replaces the 3rd.
//
// For four-part (or more) chords, the next symbol is usually a '7' with an
// optional modifier indicating a major/sharp 7th: 'maj', '∆', '△', '#', or '♯'.
// This may be followed by additional tones, '2', '4', '5', '6', '9', '11',
// and/or '13', each of which may be preceded by an accidental. Presence of such
// a subsequent tone that is greater than 7 (e.g 9, 11, 13) implies presence of
// the 7th.
//
// A chord can end with a bass tone, indicated by a '/' followed by the bass tone
// (same syntax as the chord's root tone: a note name, A-G, followed by an
// optional accidental).
//
// Examples:
//  C    Cmaj   - Both forms are C major triads (C E G).
//  Bb7#9       - B-flat major chord with a dominant 7 and sharp 9
//                (Bb D F Ab C#)
//  A-7  Amin7  - Both forms are A minor 7 chords (A C E G).
//  G♯△9♯11     - G-sharp major chord with a major 7, 9, and sharp 11
//                (G# B# D# Fx A# Cx)
//
func ParseChord(s string) (*Chord, error) {
	lx := newLexer(s)
	chordParse(lx)
	return lx.res, lx.err
}

// MustParseChord parses the given string and panics if it is not a valid
// chord representation.
func MustParseChord(s string) *Chord {
	ch, err := ParseChord(s)
	if err != nil {
		panic(err)
	}
	if ch == nil {
		panic(errors.New("ParseChord unexpectedly returned nil"))
	}
	return ch
}

// Validate checks that the current chord is valid. This checks that the
// chord definition is consistent. For example, if ExtraTones contains both
// a flat 7th and a sharp 7th, the chord is inconsistent and this method
// will return false. Similarly, a diminished chord that has a natural or
// sharp 5th or an augmented chord that has a natural or flat 5th is
// considered inconsistent and thus invalid. A chord whose triad type
// is SUS but has no valid suspension note in its ExtraTones is also invalid.
func (ch *Chord) Validate() error {
	if !ch.Root.IsValid() {
		return fmt.Errorf("chord root %v is invalid", ch.Root)
	}
	if ch.Bass.N != 0 && !ch.Bass.IsValid() {
		return fmt.Errorf("chord bass note %v is invalid", ch.Bass)
	}
	if !ch.Triad.IsValid() {
		return fmt.Errorf("chord triad type %v is invalid", ch.Triad)
	}

	t := map[int8]Accidental{}
	for _, e := range ch.ExtraTones {
		if !e.IsValid() {
			return fmt.Errorf("tone %v is invalid", e)
		}
		v := e.Val
		if v > 7 {
			v -= 7
		}
		if v < 2 || v == 3 || v > 7 {
			return fmt.Errorf("tone %d is not a valid chord extra", e.Val)
		}
		a, ok := t[v]
		if ok && a != e.Acc {
			return fmt.Errorf("tone %d has conflicting accidentals: %v and %v", e.Val, a, e.Acc)
		} else if !ok {
			t[v] = e.Acc
		}
	}

	if ch.Triad == FDim || ch.Triad == Dim3 {
		a, ok := t[7]
		if ok && a != Natural {
			return fmt.Errorf("diminished chord (other than half diminished) should not have modified 7th: %v", a)
		}
	}
	if ch.Triad == FDim || ch.Triad == HDim || ch.Triad == Dim3 {
		a, ok := t[5]
		if ok && a != Flat {
			return fmt.Errorf("diminished chord should not have non-flat 5th: %v", a)
		}
	} else if ch.Triad == Aug3 {
		a, ok := t[5]
		if ok && a != Sharp {
			return fmt.Errorf("augmented chord should not have non-sharp 5th: %v", a)
		}
	} else if ch.Triad == Sus {
		_, ok2 := t[2]
		if !ok2 {
			_, ok4 := t[4]
			if !ok4 {
				return errors.New("suspended chord must have 2nd or 4th as suspension note")
			}
		}
	}

	return nil
}

// Canonicalize modifies the chord into a simpler and canonical representation.
// If there are multiple ways to describe the same chord, this converts the chord
// into the one "canonical" representation. For example a minor chord with a flat
// 5th in its ExtraTones is canonically a diminished chord (or half diminished if
// it has a 7th). Similarly, a major chord with a sharp 5th is canonically an
// augmented chord. Some other steps to canonicalize include adjusting and sorting
// the list of ExtraTones, like removing any duplicates and renaming 2nd to 9th or
// vice versa (based on the presence of the 7th and whether the tone is a
// suspension note). This also removes ExtraTones that are effectively redundant
// due to describing enharmonic equivalents. For example, a chord with #4 b5
// (sharp fourth and flat fifth) will not have a sharp fourth after it is
// canonicalized since the two tones are enharmonic equivalents.
func (ch *Chord) Canonicalize() {
	if ch.canonical {
		return
	}
	t := map[int8][]ChordTone{}
	hasSeventh := false
	hasNaturalSeventh := false
	impliedSeventh := 0
	if ch.Triad == FDim || ch.Triad == HDim {
		impliedSeventh++
	}
	for _, e := range ch.ExtraTones {
		// remove any double-sharp sevenths or double-flat seconds since they
		// are enharmonically equivalent to the root tone
		if (e.Val == 7 && e.Acc == DblSharp) ||
			(e.Val == 2 && e.Acc == DblFlat) {
			continue
		}
		if e.Val == 9 && e.Acc == DblFlat {
			// double-flat 9 is also the same as root tone, but implies 7th
			impliedSeventh++
			continue
		}

		if e.Val > 7 {
			impliedSeventh++
		} else if e.Val == 7 {
			hasSeventh = true
			if e.Acc == Natural {
				hasNaturalSeventh = true
			}
		}
		t[e.Val] = append(t[e.Val], e)
	}

	// remove any redundant 5th tones
	switch ch.Triad {
	case Maj3, Min3, Sus:
		t[5] = removeTone(t[5], ChordTone{Val: 5})
	case Aug3:
		t[5] = removeTone(t[5], ChordTone{Val: 5, Acc: Sharp})
	case Dim3, HDim, FDim:
		t[5] = removeTone(t[5], ChordTone{Val: 5, Acc: Flat})
	}

	// convert minor chord w/ b5 to some kind of diminished
	if ch.Triad == Min3 {
		convert := false
		for _, tn := range t[5] {
			if tn.Acc == Flat {
				convert = true
				break
			}
		}
		if convert {
			t[5] = removeTone(t[5], ChordTone{Val: 5, Acc: Flat})
			if impliedSeventh > 0 || hasSeventh {
				// "minor 7" -> "half diminished"
				ch.Triad = HDim
			} else {
				// "minor" (no 7th) -> "diminished"
				ch.Triad = Dim3
			}
		}
	}
	// similarly, convert major chord w/ #5 to augmented
	if ch.Triad == Maj3 {
		convert := false
		for _, tn := range t[5] {
			if tn.Acc == Sharp {
				convert = true
				break
			}
		}
		if convert {
			t[5] = removeTone(t[5], ChordTone{Val: 5, Acc: Sharp})
			ch.Triad = Aug3
		}
	}

	// canonicalize "dim7" -> "o"
	if ch.Triad == Dim3 && (hasNaturalSeventh || (impliedSeventh > 0 && !hasSeventh)) {
		ch.Triad = FDim
		impliedSeventh++
	}

	// half diminished w/ flat 7th is the same as fully diminished
	if ch.Triad == HDim && hasSeventh {
		onlyFlatSeventh := true
		for _, s := range t[7] {
			if s.Acc != Flat {
				onlyFlatSeventh = false
			}
		}
		if onlyFlatSeventh {
			for i := range t[7] {
				t[7][i].Acc = Natural
			}
			ch.Triad = FDim
		}
	}

	// if "7" is just implied, make it explicit
	if impliedSeventh > 0 && !hasSeventh {
		t[7] = append(t[7], ChordTone{Val: 7})
		hasSeventh = true
	}

	// now we want to eliminate several enharmonic equivalents
	// (e.g. if a chord has both A# and Bb, only keep one)

	// chords with a major third can remove any flatted fourths and chords
	// with a minor third can remove any sharp seconds (enharmonic equivalents)
	// likewise double-sharp second is equivalent to major 3rd and
	// double-flat fourth is equivalent to minor 3rd
	switch ch.Triad {
	case Maj3, Aug3:
		t[4] = removeTone(t[4], ChordTone{Val: 4, Acc: Flat})
		t[11] = removeTone(t[11], ChordTone{Val: 11, Acc: Flat})
		t[2] = removeTone(t[2], ChordTone{Val: 2, Acc: DblSharp})
		t[9] = removeTone(t[9], ChordTone{Val: 9, Acc: DblSharp})
	case Min3, Dim3, HDim, FDim:
		t[2] = removeTone(t[2], ChordTone{Val: 2, Acc: Sharp})
		t[9] = removeTone(t[9], ChordTone{Val: 9, Acc: Sharp})
		t[4] = removeTone(t[4], ChordTone{Val: 4, Acc: DblFlat})
		t[11] = removeTone(t[11], ChordTone{Val: 11, Acc: DblFlat})
	}

	// sus chords with a sharp second or flatted fourth can be converted
	// to minor or major (since their suspended note is enharmonically
	// equivalent to a third)
	if ch.Triad == Sus {
		for {
			// first check 4ths
			count := len(t[4]) + len(t[11])
			t[4] = removeTone(t[4], ChordTone{Val: 4, Acc: Flat})
			t[11] = removeTone(t[11], ChordTone{Val: 11, Acc: Flat})
			if count > len(t[4])+len(t[11]) {
				ch.Triad = Maj3
				break
			}
			t[4] = removeTone(t[4], ChordTone{Val: 4, Acc: DblFlat})
			t[11] = removeTone(t[11], ChordTone{Val: 11, Acc: DblFlat})
			if count > len(t[4])+len(t[11]) {
				ch.Triad = Min3
				break
			}

			// if none found, check 2nds
			count = len(t[2]) + len(t[9])
			t[2] = removeTone(t[2], ChordTone{Val: 2, Acc: Sharp})
			t[9] = removeTone(t[9], ChordTone{Val: 9, Acc: Sharp})
			if count > len(t[2])+len(t[9]) {
				ch.Triad = Min3
				break
			}
			t[2] = removeTone(t[2], ChordTone{Val: 2, Acc: DblSharp})
			t[9] = removeTone(t[9], ChordTone{Val: 9, Acc: DblSharp})
			if count > len(t[2])+len(t[9]) {
				ch.Triad = Maj3
				break
			}
			break
		}
	}

	// fully-diminished chords don't need to specify 6th
	// (since it's enharmonic equivalent of their flat 7th)
	if ch.Triad == FDim {
		t[6] = removeTone(t[6], ChordTone{Val: 6})
		t[13] = removeTone(t[13], ChordTone{Val: 13})
	}
	// augmented chords don't need to specify flat 6th
	// (since it's enharmonic equivalent of their sharp 5th)
	if ch.Triad == Aug3 || containsTone(t[5], ChordTone{Val: 5, Acc: Sharp}) {
		t[6] = removeTone(t[6], ChordTone{Val: 6, Acc: Flat})
		t[13] = removeTone(t[13], ChordTone{Val: 13, Acc: Flat})
	}
	// just as (non-sus) diminished chords don't need to specify sharp 4th
	if ch.Triad == Dim3 || ch.Triad == HDim || ch.Triad == FDim ||
		(ch.Triad != Sus && containsTone(t[5], ChordTone{Val: 5, Acc: Flat})) {
		t[4] = removeTone(t[4], ChordTone{Val: 4, Acc: Sharp})
		t[11] = removeTone(t[11], ChordTone{Val: 11, Acc: Sharp})
	}
	if ch.Triad == Sus && containsTone(t[5], ChordTone{Val: 5, Acc: Flat}) {
		// for sus chords w/ flat 5th, as long as there is another possible
		// suspension note (e.g. some other 2/9 or 4/11), then we can remove
		// a sharp 4th, too
		if len(t[2])+len(t[9]) > 0 {
			t[4] = removeTone(t[4], ChordTone{Val: 4, Acc: Sharp})
			t[11] = removeTone(t[11], ChordTone{Val: 11, Acc: Sharp})
		} else {
			count := len(t[4]) + len(t[11])
			t[4] = removeTone(t[4], ChordTone{Val: 4, Acc: Sharp})
			t[11] = removeTone(t[11], ChordTone{Val: 11, Acc: Sharp})
			if len(t[4])+len(t[11]) == 0 && count > 0 {
				// tsk. we removed the last 4th, so we have to put it back...
				t[4] = []ChordTone{{Val: 4, Acc: Sharp}}
			}
		}
	}
	// chords with perfect fifth don't need a (redundant) double-sharp fourth
	if (ch.Triad == Min3 || ch.Triad == Maj3) &&
		(len(t[5]) == 0 || containsTone(t[5], ChordTone{Val: 5})) {
		t[4] = removeTone(t[4], ChordTone{Val: 4, Acc: DblSharp})
		t[11] = removeTone(t[11], ChordTone{Val: 11, Acc: DblSharp})
	}

	// now we want to remove any redundant tones
	// 1. first consolidate like tones (combine 2s and 9s; 4s and 11s; etc)
	if hasSeventh {
		for k, v := range t {
			if k < 7 && k != 5 {
				for i := range v {
					v[i].Val = v[i].Val + 7
				}
				t[k+7] = append(t[k+7], v...)
				t[k] = nil
			} else if k == 12 || k == 14 {
				for i := range v {
					v[i].Val = v[i].Val - 7
				}
				t[k-7] = append(t[k-7], v...)
				t[k] = nil
			}
		}
	} else {
		for k, v := range t {
			if k > 7 {
				for i := range v {
					v[i].Val = v[i].Val - 7
				}
				t[k-7] = append(t[k-7], v...)
				t[k] = nil
			}
		}
	}
	// 2. remove tones that have identical modifiers
	for k, v := range t {
		tones := map[ChordTone]struct{}{}
		for _, tn := range v {
			tones[tn] = struct{}{}
		}
		v = nil
		for tn := range tones {
			v = append(v, tn)
		}
		t[k] = v
	}

	// if we have a seventh, then tones were consolidated above into
	// the high range (e.g. 9/11/13), but if the chord is a sus tone,
	// we need to move the suspended note down (9/11 -> 2/4)
	if hasSeventh && ch.Triad == Sus {
		elevens := t[11]
		if len(elevens) > 0 {
			toDemote := -1
			for i, tn := range elevens {
				// move a natural tone if present
				// otherwise move minimum accidental
				if tn.Acc == Natural {
					toDemote = i
					break
				} else if toDemote < 0 || tn.Acc < elevens[toDemote].Acc {
					toDemote = i
				}
			}
			elevens[toDemote].Val = 4
			t[4] = []ChordTone{elevens[toDemote]}
			t[11] = append(elevens[:toDemote], elevens[toDemote+1:]...)
		} else {
			nines := t[9]
			if len(nines) > 0 {
				toDemote := -1
				for i, tn := range nines {
					// move a natural tone if present
					// otherwise move minimum accidental
					if tn.Acc == Natural {
						toDemote = i
						break
					} else if toDemote < 0 || tn.Acc < nines[toDemote].Acc {
						toDemote = i
					}
				}
				nines[toDemote].Val = 2
				t[2] = []ChordTone{nines[toDemote]}
				t[9] = append(nines[:toDemote], nines[toDemote+1:]...)
			}
		}
	}

	// now we can construct the canonical slice of tones
	ch.ExtraTones = nil
	for _, e := range t {
		ch.ExtraTones = append(ch.ExtraTones, e...)
	}
	sort.Sort(tones(ch.ExtraTones))

	ch.canonical = true
}

func removeTone(tns []ChordTone, toRemove ChordTone) []ChordTone {
	var ret []ChordTone
	for _, tn := range tns {
		if tn != toRemove {
			ret = append(ret, tn)
		}
	}
	return ret
}

func containsTone(tns []ChordTone, search ChordTone) bool {
	for _, tn := range tns {
		if tn == search {
			return true
		}
	}
	return false
}

// String implements the Stringer interface to produce a string representation
// of the Chord. This should be invertible: string products can be parsed via
// ParseChord to re-create the Chord instance.
func (ch *Chord) String() string {
	var b bytes.Buffer
	b.WriteString(ch.Root.String())
	if ch.Triad != Maj3 {
		b.WriteString(ch.Triad.String())
	}
	var prev string
	for i, t := range ch.ExtraTones {
		str := t.String()
		if t.Val == 7 && (t.Acc == Natural || t.Acc == Sharp) &&
			(i == 0 || ch.Triad == Sus && i == 1) &&
			((i+1 < len(ch.ExtraTones) && ch.ExtraTones[i+1].Val > 7 && ch.ExtraTones[i+1].Acc == Natural) ||
				(i == len(ch.ExtraTones)-1 && (ch.Triad == FDim || ch.Triad == HDim))) {
			// omit the '7' since it is implied
			str = str[:len(str)-1]
		}
		if len(str) == 0 {
			continue
		}
		if len(prev) > 0 {
			c1 := prev[len(prev)-1]
			c2 := str[0]
			if c1 >= '0' && c1 <= '9' && c2 >= '0' && c2 <= '9' {
				// we don't want two numbers together, e.g. "9 11" instead of "911"
				b.WriteByte(' ')
			}
		}
		b.WriteString(str)
		prev = str
	}
	if ch.Bass.N > 0 {
		b.WriteByte('/')
		b.WriteString(ch.Bass.String())
	}
	return b.String()
}

// Spell enumerates all of the notes in the chord. For example, a C major
// chord is spelled C, E, G. An E dominant 7 sharp 9 (aka E7#9, or the Hendrix
// chord) is spelled E, G#, B, D, Fx.
func (ch *Chord) Spell() []Note {
	tones := make([]ChordTone, 0, len(ch.ExtraTones)+4)
	// root
	tones = append(tones, ChordTone{Val: 1})
	// and third
	if ch.Triad != Sus {
		tones = append(tones, ChordTone{Val: 3})
	}
	// then fifth
	hasFifth := false
	hasSeventh := false
	for _, tn := range ch.ExtraTones {
		if tn.Val == 5 {
			hasFifth = true
			if hasSeventh || (ch.Triad != FDim && ch.Triad != HDim) {
				break
			}
		}
		if tn.Val == 7 {
			hasSeventh = true
			if hasFifth {
				break
			}
		}
	}
	if !hasFifth {
		tones = append(tones, ch.Triad.fifthTone())
	}
	// and maybe seventh
	if !hasSeventh && (ch.Triad == FDim || ch.Triad == HDim) {
		// fully and half diminished imply the 7th
		tones = append(tones, ChordTone{Val: 7})
	}

	tones = append(tones, ch.ExtraTones...)
	sort.Sort(spellTonesFor(tones, ch.Triad == Sus))

	// now we convert the tones into intervals
	std := standardIntervals[ch.Triad]
	ints := make([]Interval, len(tones))
	for i, tn := range tones {
		v := tn.Val
		if v > 7 {
			v -= 7
		}
		ints[i] = Interval{Val: v, Offset: std[v-1] + tn.Acc.Offset()}
	}

	ret := TransposeNote(ch.Root, ints...)
	if ch.Bass.N != 0 {
		p := make([]Note, 0, len(ret)+1)
		p = append(p, ch.Bass)
		ret = append(p, ret...)
	}
	return ret
}

func (c *Chord) ChordType() *ChordType {
	var bassInterval Interval
	if c.Bass.N != 0 {
		bassInterval = c.Root.IntervalTo(c.Bass)
	}
	return &ChordType{
		Triad:      c.Triad,
		ExtraTones: c.ExtraTones,
		Bass:       bassInterval,
		canonical:  c.canonical,
	}
}

// ChordTone represents an element of a chord. A Tone is very similar to an
// Interval except that it has context of the chord (for example a value of 3
// for a ChordTone represents a minor third for minor and diminished chords, but
// a major third for major and augmented chords). Also chord tones allow values
// greater than seven, which imply a seventh tone. For example, a chord with a 9
// Tone implies both a seventh and a second tone (i.e. 5-part harmony).
type ChordTone struct {
	// Value indicates what note it is relative to the root of the
	// chord. In diatonic music, there are seven notes, so the value
	// will be 1 through 7. Values above 7 (8 through 14) are synonymous
	// with the lower values (e.g. Val minus 7), and are often used in
	// jazz chords where the 7th is present in a chord. For example, the
	// 9 tone means the 2nd tone relative to the chord root, but in a
	// chord where a 7 tone is also present. Values greater than 14, as
	// well as the value zero, are not valid.
	Val int8
	// Acc modifies the value. For example, if a chord has 5th that is
	// one half-step higher than the normal perfect 5th, the 5 tone may
	// have Acc set to Sharp.
	Acc Accidental
}

// String implements the Stringer interface.
func (t ChordTone) String() string {
	var acc string
	if t.Val == 7 && t.Acc == Sharp {
		acc = "△"
	} else if t.Acc != Natural {
		acc = t.Acc.String()
	}
	return fmt.Sprintf("%s%d", acc, t.Val)
}

// IsValid returns true if this tone contains only valid values. If Val
// is outside the allowed range (1 to 14) or if Acc is not valid, this
// will return false.
func (t ChordTone) IsValid() bool {
	return t.Val >= 1 && t.Val <= 14 && t.Acc.IsValid()
}

// TriadType indicates the basic "shape" of a chord. The shape of
// a triad describes the distance (3 or 4 half-steps) between its
// three main tones: root, 3rd, and 5th. There are several special
// triad types that actually have implications for other notes
// other than the core root, 3rd, and 5th tones.
type TriadType int

const (
	// Maj3 represents a major triad. This chord shape has 4
	// half-steps between its root and 3rd and 3 half-steps between
	// its 3rd and 5th.
	Maj3 TriadType = iota
	// Aug3 represents an augmented triad. This chord shape has 4
	// half-steps both between its root and 3rd and between its 3rd
	// and 5th.
	Aug3
	// Min3 represents a minor triad. This chord shape has 3
	// half-steps between its root and 3rd and 4 half-steps between
	// its 3rd and 5th (reverse of a major chord).
	Min3
	// Dim3 represents a diminished triad. This chord shape has 3
	// half-steps both between its root and 3rd and between its 3rd
	// and 5th. This triad shape is indicated by the word "dim" in
	// the chord name. In addition to this shape (minor 3rd and flat
	// 5th), it also means if that the 7th tone is present, it is a
	// flat 7th (forming a "fully diminished" four tone chord).
	Dim3
	// HDim is a special triad indicator because it implies a fourth
	// tone in the chord: the 7th. It means "half diminished"
	// and is indicated by a circle with slash through it: or "ø".
	// This indicator is equivalent to "min7 b5" (e.g. minor 7th chord
	// with a flat 5th). So Bø and Bmin7b5 are the same.
	HDim
	// FDim is a special triad indicator because it implies a fourth
	// tone in the chord: the flat 7th. It means "fully diminished"
	// and is indicated by a circle, or "o". This indicator is
	// equivalent to "dim7". So C#o and C#dim7 are the same.
	FDim
	// Sus is another special triad indicator that means the 3rd
	// tone is replaced with either a 2nd or 4th tone.
	Sus
)

// String implements the Stringer interface.
func (t TriadType) String() string {
	switch t {
	case Maj3:
		return "maj"
	case Aug3:
		return "+"
	case Min3:
		return "-"
	case Dim3:
		return "dim"
	case HDim:
		return "ø"
	case FDim:
		return "o"
	case Sus:
		return "sus"
	default:
		return fmt.Sprintf("?(%d)", t)
	}
}

// IsValid returns true if this TriadType is valid. If it has a numeric
// value that does not correspond to one of the constants in this
// package then it is not valid.
func (t TriadType) IsValid() bool {
	return t >= Maj3 && t <= Sus
}

func (t TriadType) fifthTone() ChordTone {
	switch t {
	case HDim, Dim3, FDim:
		return ChordTone{Val: 5, Acc: Flat}
	case Aug3:
		return ChordTone{Val: 5, Acc: Sharp}
	default:
		return ChordTone{Val: 5}
	}
}

var standardIntervals [][]int8

func init() {
	// chord tones assume major second, perfect fourth, perfect fifth, major sixth,
	// and minor seventh; this table makes further adjustment since the chord has
	// a minor third for some triad types but major third for others and this also
	// includes the extra offset for the seventh tone of fully diminished chords
	// (all offsets are compared to major scale/ionian mode)
	standardIntervals = make([][]int8, Sus+1)
	standardIntervals[Maj3] = []int8{0, 0, 0, 0, 0, 0, -1}
	standardIntervals[Sus] = []int8{0, 0, 0, 0, 0, 0, -1}
	standardIntervals[Min3] = []int8{0, 0, -1, 0, 0, 0, -1}
	standardIntervals[Aug3] = []int8{0, 0, 0, 0, 0, 0, -1}
	standardIntervals[Dim3] = []int8{0, 0, -1, 0, 0, 0, -2}
	standardIntervals[HDim] = []int8{0, 0, -1, 0, 0, 0, -1}
	standardIntervals[FDim] = []int8{0, 0, -1, 0, 0, 0, -2}
}

type triad struct {
	typ     TriadType
	susTone ChordTone
}

// sorting tones for canonicalizing chords' extra tones
type tones []ChordTone

func (t tones) Len() int {
	return len(t)
}

func (t tones) Less(i, j int) bool {
	bi := toneOrder(t[i].Val)
	bj := toneOrder(t[j].Val)
	if bi < bj {
		return true
	} else if bi == bj {
		return t[i].Acc.Offset() < t[j].Acc.Offset()
	}
	return false
}

func toneOrder(b int8) int8 {
	// modified 5s are last
	if b == 5 {
		return math.MaxInt8
	}
	return b
}

func (t tones) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

// sorting tones for spelling chords
type spellTones struct {
	t          []ChordTone
	susTone    ChordTone
	hasSeventh bool
}

func spellTonesFor(tns []ChordTone, isSus bool) spellTones {
	var st ChordTone
	if isSus {
		var t2, t4 ChordTone
		for _, t := range tns {
			if t.Val == 2 {
				t2 = t
			} else if t.Val == 4 {
				t4 = t
			}
		}
		if t4.Val != 0 {
			st = t4
		} else if t2.Val != 0 {
			st = t2
		} else {
			for _, t := range tns {
				if t.Val == 9 {
					t2 = t
				} else if t.Val == 11 {
					t4 = t
				}
			}
			if t4.Val != 0 {
				st = t4
			} else if t2.Val != 0 {
				st = t2
			}
		}
	}
	hasSeventh := false
	for _, t := range tns {
		if t.Val == 7 {
			hasSeventh = true
			break
		}
	}
	return spellTones{t: tns, susTone: st, hasSeventh: hasSeventh}
}

func (t spellTones) Len() int {
	return len(t.t)
}

func (t spellTones) Less(i, j int) bool {
	bi := t.spellToneOrder(t.t[i])
	bj := t.spellToneOrder(t.t[j])
	if bi < bj {
		return true
	} else if bi == bj {
		return t.t[i].Acc.Offset() < t.t[j].Acc.Offset()
	}
	return false
}

func (t spellTones) spellToneOrder(tn ChordTone) int8 {
	if tn.Val == 1 || tn.Val == 3 || tn.Val == 5 || tn.Val == 7 {
		return tn.Val
	}
	if tn.Val < 5 && tn == t.susTone {
		return tn.Val
	}
	if tn.Val == 6 && !t.hasSeventh {
		return tn.Val
	}
	return tn.Val + 7
}

func (t spellTones) Swap(i, j int) {
	t.t[i], t.t[j] = t.t[j], t.t[i]
}

// ChordType represents an abstract chord -- a general type of chord, but
// without a particular root note. For example, a "major seventh" chord is
// a ChordType, whereas a "C major seven" is a Chord (whose chord type is
// a "major seventh" and root is C).
type ChordType struct {
	// Triad indicates the basic "shape" of the chord, which is the distance
	// between the primary three tones of the chord. A special triad type,
	// SUS, indicates a chord where the 3rd tone is omitted. In such a case,
	// ExtraTones should contain the suspension note (the 2nd or 4th tone
	// from the root).
	Triad TriadType
	// The additional tones in the chord, other than the root, third, and
	// fifth.
	ExtraTones []ChordTone
	// Bass represents the lowest pitch in the chord. It is an interval
	// interval relative to the chord's root. A zero value is considered the
	// same as the 1st interval (e.g. tonic).
	Bass Interval
	// canonical is true if Canonicalize has been called to ensure this
	// chord is a canonical form.
	canonical bool
}

func (c *ChordType) Chord(root Note) *Chord {
	var zero Interval
	var bassNote Note
	if c.Bass != zero {
		bassNote = root.Transpose(c.Bass)
	}
	return &Chord{
		Root:       root,
		Triad:      c.Triad,
		ExtraTones: c.ExtraTones,
		Bass:       bassNote,
		canonical:  c.canonical,
	}
}

// TODO: ChordType.Canonicalize()

// ScaleChord represents a chord that can be transposed to any scale.
// Instead of having chord tones represented as notes (like C# for example),
// they are represented as an interval relative to a scale root.
//
// ScaleChords have a string form that uses roman numeral notation for
// chords. It uses lower-case roman numerals for chords that are minor or
// diminished, and upper-case roman numerals for chords that are major or
// augmented. For inversions, the bass note is also represented as a roman
// numeral, indicating the bass note's interval from the scale root.
//
// For example, a ScaleChord with a root of {3,0} (i.e. a major third) and
// a type that is a major triad with a dominant 7 would be printed to string
// as "III 7 9". If the ScaleChord were a minor triad with no extra tones and
// and a root of {4,0} (e.g. a perfect fourth), it would be "iv".
//
// Whether a root interval of a major third is printed as "iii" vs "# iii"
// (or similarly, a minor third printed as "iii" vs "♭iii") depends on
// whether the ScaleChord is in the context of a minor key or a major key.
type ScaleChord struct {
	// The root of the chord, relative ot the root of some scale.
	Root Interval
	// If InMinorKey is true, then when the ScaleChord is printed via
	// String(), the roman numeral intervals are unadorned (no accidentals)
	// if they match the intervals of a minor scale. For example, if true,
	// then "iii" or "III" has a root note that is a minor third above the
	// scale root. If false (NOT a minor key), then "iii" would have a root that
	// is a major third above; and a chord whose root was a minor third above
	// would be printed as "♭iii".
	InMinorKey bool
	// The actual type of the chord.
	Type ChordType
}

func (s *ScaleChord) InKey(keyName Note) *Chord {
	chordRoot := keyName.Transpose(s.Root)
	return s.Type.Chord(chordRoot)
}

func (s *ScaleChord) String() string {
	// TODO
	// iv
	return ""
}

// TODO: ParseScaleChord?

func NewScaleChord(s ScaleType, root int8, extraTones ...int8) *ScaleChord {
	// TODO
	return nil
}

func InferChord(notes ...Note) *Chord {
	// TODO: wouldn't this be cool
	return nil
}
