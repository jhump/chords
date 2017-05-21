package chords

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"sort"
)

//go:generate go tool yacc -o chordparse.y.go -p chord chordparse.y

type Chord struct {
	Root       Pitch
	Triad      TriadType
	ExtraTones []Tone
	Bass       Pitch
	canonical  bool
}

func ParseChord(s string) (*Chord, error) {
	lx := newLexer(s)
	chordParse(lx)
	return lx.res, lx.err
}

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

func (ch *Chord) Validate() error {
	if !ch.Root.IsValid() {
		return fmt.Errorf("Chord root %v is invalid", ch.Root)
	}
	if ch.Bass.N != 0 && !ch.Bass.IsValid() {
		return fmt.Errorf("Chord bass note %v is invalid", ch.Bass)
	}
	if !ch.Triad.IsValid() {
		return fmt.Errorf("Chord triad type %v is invalid", ch.Triad)
	}

	t := map[byte]Accidental{}
	for _, e := range ch.ExtraTones {
		if !e.IsValid() {
			return fmt.Errorf("Tone %v is invalid", e)
		}
		v := e.Val
		if v > 7 {
			v -= 7
		}
		if v < 2 || v == 3 || v > 7 {
			return fmt.Errorf("Tone %d is not a valid chord extra", e.Val)
		}
		a, ok := t[v]
		if ok && a != e.Acc {
			return fmt.Errorf("Tone %d has conflicting accidentals: %v and %v", e.Val, a, e.Acc)
		} else if !ok {
			t[v] = e.Acc
		}
	}

	if ch.Triad == FDIM || ch.Triad == DIM3 {
		a, ok := t[7]
		if ok && a != NATURAL {
			return fmt.Errorf("Diminished chord (other than half diminished) should not have modified 7th: %v", a)
		}
	}
	if ch.Triad == FDIM || ch.Triad == HDIM || ch.Triad == DIM3 {
		a, ok := t[5]
		if ok && a != FLAT {
			return fmt.Errorf("Diminished chord should not have non-flat 5th: %v", a)
		}
	} else if ch.Triad == AUG3 {
		a, ok := t[5]
		if ok && a != SHARP {
			return fmt.Errorf("Augmented chord should not have non-sharp 5th: %v", a)
		}
	} else if ch.Triad == SUS {
		_, ok2 := t[2]
		if !ok2 {
			_, ok4 := t[4]
			if !ok4 {
				return errors.New("Suspended chord must have 2nd or 4th as suspension note")
			}
		}
	}

	return nil
}

func (ch *Chord) Canonicalize() {
	if ch.canonical {
		return
	}
	t := map[byte][]Tone{}
	hasSeventh := false
	hasNaturalSeventh := false
	impliedSeventh := 0
	if ch.Triad == FDIM || ch.Triad == HDIM {
		impliedSeventh++
	}
	for _, e := range ch.ExtraTones {
		// remove any double-sharp sevenths or double-flat seconds since they
		// are enharmonically equivalent to the root tone
		if (e.Val == 7 && e.Acc == DBL_SHARP) ||
				(e.Val == 2 && e.Acc == DBL_FLAT) {
			continue
		}
		if e.Val == 9 && e.Acc == DBL_FLAT {
			// double-flat 9 is also the same as root tone, but implies 7th
			impliedSeventh++
			continue
		}

		if e.Val > 7 {
			impliedSeventh++
		} else if e.Val == 7 {
			hasSeventh = true
			if e.Acc == NATURAL {
				hasNaturalSeventh = true
			}
		}
		t[e.Val] = append(t[e.Val], e)
	}

	// remove any redundant 5th tones
	switch (ch.Triad) {
	case MAJ3, MIN3, SUS:
		t[5] = removeTone(t[5], Tone{Val:5})
	case AUG3:
		t[5] = removeTone(t[5], Tone{Val:5, Acc: SHARP})
	case DIM3, HDIM, FDIM:
		t[5] = removeTone(t[5], Tone{Val:5, Acc: FLAT})
	}

	// convert "minor7 b5" to half diminished
	if ch.Triad == MIN3 && (impliedSeventh > 0 || hasSeventh) {
		convert := false
		for _, tn := range t[5] {
			if tn.Acc == FLAT {
				convert = true
				break
			}
		}
		if convert {
			t[5] = removeTone(t[5], Tone{Val:5, Acc: FLAT})
			ch.Triad = HDIM
		}
	}

	// canonicalize "dim7" -> "o"
	if ch.Triad == DIM3 && (hasNaturalSeventh || (impliedSeventh > 0 && !hasSeventh)) {
		ch.Triad = FDIM
		impliedSeventh++
	}

	// half diminished w/ flat 7th is the same as fully diminished
	if ch.Triad == HDIM && hasSeventh {
		onlyFlatSeventh := true
		for _, s := range t[7] {
			if s.Acc != FLAT {
				onlyFlatSeventh = false
			}
		}
		if onlyFlatSeventh {
			for i := range t[7] {
				t[7][i].Acc = NATURAL
			}
			ch.Triad = FDIM
		}
	}

	// if "7" is just implied, make it explicit
	if impliedSeventh > 0 && !hasSeventh {
		t[7] = append(t[7], Tone{Val:7})
		hasSeventh = true
	}

	// now we want to eliminate several enharmonic equivalents
	// (e.g. if a chord has both A# and Bb, only keep one)

	// chords with a major third can remove any flatted fourths and chords
	// with a minor third can remove any sharp seconds (enharmonic equivalents)
	// likewise double-sharp second is equivalent to major 3rd and
	// double-flat fourth is equivalent to minor 3rd
	switch (ch.Triad) {
	case MAJ3, AUG3:
		t[4] = removeTone(t[4], Tone{Val:4, Acc: FLAT})
		t[11] = removeTone(t[11], Tone{Val:11, Acc: FLAT})
		t[2] = removeTone(t[2], Tone{Val:2, Acc: DBL_SHARP})
		t[9] = removeTone(t[9], Tone{Val:9, Acc: DBL_SHARP})
	case MIN3, DIM3, HDIM, FDIM:
		t[2] = removeTone(t[2], Tone{Val:2, Acc: SHARP})
		t[9] = removeTone(t[9], Tone{Val:9, Acc: SHARP})
		t[4] = removeTone(t[4], Tone{Val:4, Acc: DBL_FLAT})
		t[11] = removeTone(t[11], Tone{Val:11, Acc: DBL_FLAT})
	}

	// sus chords with a sharp second or flatted fourth can be converted
	// to minor or major (since their suspended note is enharmonically
	// equivalent to a third)
	if ch.Triad == SUS {
		for {
			// first check 4ths
			count := len(t[4]) + len(t[11])
			t[4] = removeTone(t[4], Tone{Val:4, Acc: FLAT})
			t[11] = removeTone(t[11], Tone{Val:11, Acc: FLAT})
			if count > len(t[4]) + len(t[11]) {
				ch.Triad = MAJ3
				break
			}
			t[4] = removeTone(t[4], Tone{Val:4, Acc: DBL_FLAT})
			t[11] = removeTone(t[11], Tone{Val:11, Acc: DBL_FLAT})
			if count > len(t[4]) + len(t[11]) {
				ch.Triad = MIN3
				break
			}

			// if none found, check 2nds
			count = len(t[2]) + len(t[9])
			t[2] = removeTone(t[2], Tone{Val:2, Acc: SHARP})
			t[9] = removeTone(t[9], Tone{Val:9, Acc: SHARP})
			if count > len(t[2]) + len(t[9]) {
				ch.Triad = MIN3
				break
			}
			t[2] = removeTone(t[2], Tone{Val:2, Acc: DBL_SHARP})
			t[9] = removeTone(t[9], Tone{Val:9, Acc: DBL_SHARP})
			if count > len(t[2]) + len(t[9]) {
				ch.Triad = MAJ3
				break
			}
			break
		}
	}

	// fully-diminished chords don't need to specify 6th
	// (since it's enharmonic equivalent of their flat 7th)
	if ch.Triad == FDIM {
		t[6] = removeTone(t[6], Tone{Val:6})
		t[13] = removeTone(t[13], Tone{Val:13})
	}
	// augmented chords don't need to specify flat 6th
	// (since it's enharmonic equivalent of their sharp 5th)
	if ch.Triad == AUG3 || containsTone(t[5], Tone{Val:5, Acc:SHARP}) {
		t[6] = removeTone(t[6], Tone{Val:6, Acc:FLAT})
		t[13] = removeTone(t[13], Tone{Val:13, Acc:FLAT})
	}
	// just as (non-sus) diminished chords don't need to specify sharp 4th
	if ch.Triad == DIM3 || ch.Triad == HDIM || ch.Triad == FDIM ||
			(ch.Triad != SUS && containsTone(t[5], Tone{Val:5, Acc:FLAT})) {
		t[4] = removeTone(t[4], Tone{Val:4, Acc:SHARP})
		t[11] = removeTone(t[11], Tone{Val:11, Acc:SHARP})
	}
	if ch.Triad == SUS && containsTone(t[5], Tone{Val:5, Acc:FLAT}) {
		// for sus chords w/ flat 5th, as long as there is another possible
		// suspension note (e.g. some other 2/9 or 4/11), then we can remove
		// a sharp 4th, too
		if len(t[2]) + len(t[9]) > 0 {
			t[4] = removeTone(t[4], Tone{Val:4, Acc:SHARP})
			t[11] = removeTone(t[11], Tone{Val:11, Acc:SHARP})
		} else {
			count := len(t[4]) + len(t[11])
			t[4] = removeTone(t[4], Tone{Val:4, Acc:SHARP})
			t[11] = removeTone(t[11], Tone{Val:11, Acc:SHARP})
			if len(t[4]) + len(t[11]) == 0 && count > 0 {
				// tsk. we removed the last 4th, so we have to put it back...
				t[4] = []Tone{{Val:4, Acc:SHARP}}
			}
		}
	}
	// chords with perfect fifth don't need a (redundant) double-sharp fourth
	if (ch.Triad == MIN3 || ch.Triad == MAJ3) &&
			(len(t[5]) == 0 || containsTone(t[5], Tone{Val:5})) {
		t[4] = removeTone(t[4], Tone{Val:4, Acc:DBL_SHARP})
		t[11] = removeTone(t[11], Tone{Val:11, Acc:DBL_SHARP})
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
		tones := map[Tone]struct{}{}
		for _, tn := range v {
			tones[tn] = struct{}{}
		}
		v = nil
		for tn, _ := range tones {
			v = append(v, tn)
		}
		t[k] = v
	}

	// if we have a seventh, then tones were consolidated above into
	// the high range (e.g. 9/11/13), but if the chord is a sus tone,
	// we need to move the suspended note down (9/11 -> 2/4)
	if hasSeventh && ch.Triad == SUS {
		elevens := t[11]
		if len(elevens) > 0 {
			toDemote := -1
			for i, tn := range elevens {
				// move a natural tone if present
				// otherwise move minimum accidental
				if tn.Acc == NATURAL {
					toDemote = i
					break
				} else if toDemote < 0 || tn.Acc < elevens[toDemote].Acc {
					toDemote = i
				}
			}
			elevens[toDemote].Val = 4
			t[4] = []Tone{elevens[toDemote]}
			t[11] = append(elevens[:toDemote], elevens[toDemote+1:]...)
		} else {
			nines := t[9]
			if len(nines) > 0 {
				toDemote := -1
				for i, tn := range nines {
					// move a natural tone if present
					// otherwise move minimum accidental
					if tn.Acc == NATURAL {
						toDemote = i
						break
					} else if toDemote < 0 || tn.Acc < nines[toDemote].Acc {
						toDemote = i
					}
				}
				nines[toDemote].Val = 2
				t[2] = []Tone{nines[toDemote]}
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

func removeTone(tns []Tone, toRemove Tone) []Tone {
	var ret []Tone
	for _, tn := range tns {
		if tn != toRemove {
			ret = append(ret, tn)
		}
	}
	return ret
}

func containsTone(tns []Tone, search Tone) bool {
	for _, tn := range tns {
		if tn == search {
			return true
		}
	}
	return false
}

func (ch *Chord) String() string {
	var b bytes.Buffer
	b.WriteString(ch.Root.String())
	if ch.Triad != MAJ3 {
		b.WriteString(ch.Triad.String())
	}
	var prev string
	for i, t := range ch.ExtraTones {
		str := t.String()
		if t.Val == 7 && (t.Acc == NATURAL || t.Acc == SHARP) &&
				(i == 0 || ch.Triad == SUS && i == 1) &&
				((i + 1 < len(ch.ExtraTones) &&	ch.ExtraTones[i+1].Val > 7 && ch.ExtraTones[i+1].Acc == NATURAL) ||
						(i == len(ch.ExtraTones) - 1 && (ch.Triad == FDIM || ch.Triad == HDIM))) {
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

func (ch *Chord) Spell() []Pitch {
	tones := make([]Tone, 0, len(ch.ExtraTones) + 4)
	// root
	tones = append(tones, Tone{Val:1})
	// and third
	if ch.Triad != SUS {
		tones = append(tones, Tone{Val:3})
	}
	// then fifth
	hasFifth := false
	hasSeventh := false
	for _, tn := range ch.ExtraTones {
		if tn.Val == 5 {
			hasFifth = true
			if hasSeventh || (ch.Triad != FDIM && ch.Triad != HDIM) {
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
	if !hasSeventh && (ch.Triad == FDIM || ch.Triad == HDIM) {
		// fully and half diminished imply the 7th
		tones = append(tones, Tone{Val:7})
	}

	tones = append(tones, ch.ExtraTones...)
	sort.Sort(spellTonesFor(tones, ch.Triad == SUS))

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

	ret := TransposePitch(ch.Root, ints)
	if ch.Bass.N != 0 {
		p := make([]Pitch, 0, len(ret) + 1)
		p = append(p, ch.Bass)
		ret = append(p, ret...)
	}
	return ret
}

type Tone struct {
	Val byte // 1 through 14
	Acc Accidental
}

func (t Tone) String() string {
	var acc string
	if t.Val == 7 && t.Acc == SHARP {
		acc = "△"
	} else if t.Acc != NATURAL {
		acc = t.Acc.String()
	}
	return fmt.Sprintf("%s%d", acc, t.Val)
}

func (t Tone) IsValid() bool {
	return t.Val >= 1 && t.Val <= 14 && t.Acc.IsValid()
}

type TriadType int

const (
	MAJ3 TriadType = iota
	AUG3
	MIN3
	DIM3 // with "7" means fully-diminished
	HDIM // half diminished: implies 4th note, minor 7th
	FDIM // fully diminished: implies 4th note, major 6th
	SUS
)

func (t TriadType) String() string {
	switch t {
	case MAJ3:
		return "major"
	case AUG3:
		return "+"
	case MIN3:
		return "-"
	case DIM3:
		return "dim"
	case HDIM:
		return "ø"
	case FDIM:
		return "o"
	case SUS:
		return "sus"
	default:
		return fmt.Sprintf("?(%d)", t)
	}
}

func (t TriadType) IsValid() bool {
	return t >= MAJ3 && t <= SUS
}

func (t TriadType) fifthTone() Tone {
	switch t {
	case HDIM, DIM3, FDIM:
		return Tone{Val:5, Acc:FLAT}
	case AUG3:
		return Tone{Val:5, Acc:SHARP}
	default:
		return Tone{Val:5}
	}
}

var standardIntervals [][]int8
func init() {
	standardIntervals = make([][]int8, SUS+1)
	standardIntervals[MAJ3] = []int8{0, 0, 0, 0, 0, 0, -1}
	standardIntervals[SUS] = []int8{0, 0, 0, 0, 0, 0, -1}
	standardIntervals[MIN3] = []int8{0, 0, -1, 0, 0, 0, -1}
	standardIntervals[AUG3] = []int8{0, 0, 0, 0, 0, 0, -1}
	standardIntervals[DIM3] = []int8{0, 0, -1, 0, 0, 0, -2}
	standardIntervals[HDIM] = []int8{0, 0, -1, 0, 0, 0, -1}
	standardIntervals[FDIM] = []int8{0, 0, -1, 0, 0, 0, -2}
}

type triad struct {
	typ     TriadType
	susTone Tone
}

type tones []Tone

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

func toneOrder(b byte) byte {
	// modified 5s are last
	if b == 5 {
		return math.MaxUint8
	}
	return b
}

func (t tones) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

type spellTones struct {
	t          []Tone
	susTone    Tone
	hasSeventh bool
}

func spellTonesFor(tns []Tone, isSus bool) spellTones {
	var st Tone
	if isSus {
		var t2, t4 Tone
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

func (t spellTones) spellToneOrder(tn Tone) byte {
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
