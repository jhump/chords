%{

package chords

import (
	"errors"
)

%}

// fields inside this union end up as the fields in a structure known
// as ${PREFIX}SymType, of which a reference is passed to the lexer.
%union{
	ch    Chord
	n     Note
	t     ChordTone
	acc   Accidental
	triad triad
	tones []ChordTone
	b     int8
}

// any non-terminal which returns a value needs a type, which is
// really a field name in the above union struct
%type <ch>    chord fullChord
%type <n>     note
%type <t>     tone
%type <acc>   toneMod
%type <triad> triad sus
%type <tones> extras
%type <b>     toneVal

// same for terminals
%token <b>     _SYM_NOTE _SYM_TONE _SYM_MAJ7 _SYM_SUS
%token <acc>   _SYM_ACCIDENTAL
%token <triad> _SYM_MIN _SYM_DIM _SYM_HDIM _SYM_FDIM _SYM_AUG

%%

fullChord	: chord
		{
			$$ = $1
			chordlex.(*chordLex).res = &$$
		}
	| chord '/' note
		{
			$$ = $1
			$$.Bass = $3
			chordlex.(*chordLex).res = &$$
		}

// numerous formulations to prevent any ambiguity between an accidental
// adjusting the root note vs. the 7th
chord	: note extras
		{
			$$ = Chord{ Root: $1, ExtraTones: $2 }
		}
	| note '7' extras
		{
			$$ = Chord{ Root: $1, ExtraTones: append($3, ChordTone{Val: 7}) }
		}
	| note _SYM_MAJ7 '7' extras
		{
			$$ = Chord{ Root: $1, ExtraTones: append($4, ChordTone{Val: 7, Acc: Sharp}) }
		}
	| note triad extras
		{
			if $2.susTone.Val != 0 {
				$$ = Chord{ Root: $1, Triad: $2.typ, ExtraTones: append($3, $2.susTone) }
			} else {
				$$ = Chord{ Root: $1, Triad: $2.typ, ExtraTones: $3 }
			}
		}
	| note _SYM_MAJ7 extras
		{
			hasHighTone := false
			for _, tn := range $3 {
				if tn.Val > 7 {
					hasHighTone = true
					break
				}
			}
			if hasHighTone {
				$$ = Chord{ Root: $1, ExtraTones: append($3, ChordTone{Val: 7, Acc: Sharp}) }
			} else {
				$$ = Chord{ Root: $1, ExtraTones: $3 }
			}
		}
	| note triad '7' extras
		{
			if $2.susTone.Val != 0 {
				$$ = Chord{ Root: $1, Triad: $2.typ, ExtraTones: append($4, ChordTone{Val: 7}, $2.susTone) }
			} else {
				$$ = Chord{ Root: $1, Triad: $2.typ, ExtraTones: append($4, ChordTone{Val: 7}) }
			}
		}
	| note triad _SYM_MAJ7 '7' extras
		{
			if $2.susTone.Val != 0 {
				$$ = Chord{ Root: $1, Triad: $2.typ, ExtraTones: append($5, ChordTone{Val: 7, Acc: Sharp}, $2.susTone) }
			} else {
				$$ = Chord{ Root: $1, Triad: $2.typ, ExtraTones: append($5, ChordTone{Val: 7, Acc: Sharp}) }
			}
		}
	| note triad _SYM_MAJ7 _SYM_TONE extras
		{
			if $2.susTone.Val != 0 {
				$$ = Chord{ Root: $1, Triad: $2.typ, ExtraTones: append($5, ChordTone{Val: 7, Acc: Sharp}, ChordTone{Val: $4}, $2.susTone) }
			} else {
				$$ = Chord{ Root: $1, Triad: $2.typ, ExtraTones: append($5, ChordTone{Val: 7, Acc: Sharp}, ChordTone{Val: $4}) }
			}
		}
	| note triad _SYM_ACCIDENTAL '7' extras
		{
			if $2.susTone.Val != 0 {
				$$ = Chord{ Root: $1, Triad: $2.typ, ExtraTones: append($5, ChordTone{Val: 7, Acc: $3}, $2.susTone) }
			} else {
				$$ = Chord{ Root: $1, Triad: $2.typ, ExtraTones: append($5, ChordTone{Val: 7, Acc: $3}) }
			}
		}

note	: _SYM_NOTE
		{
			$$ = Note{ N: NoteName($1) }
		}
	| _SYM_NOTE _SYM_ACCIDENTAL
		{
			$$ = Note{ N: NoteName($1), Acc: $2 }
		}

triad	: _SYM_MIN
		{
			$$ = triad{ typ: Min3 }
		}
	| '-'
		{
			$$ = triad{ typ: Min3 }
		}
	| _SYM_DIM
		{
			$$ = triad{ typ: Dim3 }
		}
	| _SYM_HDIM
		{
			$$ = triad{ typ: HDim }
		}
	| _SYM_FDIM
		{
			$$ = triad{ typ: FDim }
		}
	| _SYM_AUG
		{
			$$ = triad{ typ: Aug3 }
		}
	| '+'
		{
			$$ = triad{ typ: Aug3 }
		}
	| sus
		{
			$$ = $1
		}

sus	: _SYM_SUS '2'
		{
			$$ = triad{ typ: Sus, susTone: ChordTone{ Val: 2 } }
		}
	| _SYM_SUS _SYM_ACCIDENTAL '2'
		{
			$$ = triad{ typ: Sus, susTone: ChordTone{ Val: 2, Acc: $2 } }
		}
	| _SYM_SUS '4'
		{
			$$ = triad{ typ: Sus, susTone: ChordTone{ Val: 4 } }
		}
	| _SYM_SUS _SYM_ACCIDENTAL '4'
		{
			$$ = triad{ typ: Sus, susTone: ChordTone{ Val: 4, Acc: $2 } }
		}

extras	: /* empty */
		{
			$$ = nil
		}
	| tone extras
		{
			$$ = append([]ChordTone{$1}, $2...)
		}

tone	: toneVal
		{
			$$ = ChordTone{ Val: $1 }
		}
	| toneMod toneVal
		{
			$$ = ChordTone{ Val: $2, Acc: $1 }
		}

toneVal	: _SYM_TONE
		{
			$$ = $1
		}
	| '2'
		{
			$$ = 2
		}
	| '4'
		{
			$$ = 4
		}
	| '5'
		{
			$$ = 5
		}
	| '6'
		{
			$$ = 6
		}

toneMod	: '-'
		{
			$$ = Flat
		}
	| '+'
		{
			$$ = Sharp
		}
	| _SYM_ACCIDENTAL
		{
			$$ = $1
		}

%%      /*  start  of  programs  */

type chordLex struct {
	input []rune
	pos   int
	err   error
	res   *Chord
}

func newLexer(s string) *chordLex {
	var r []rune
	for _, ch := range s {
		r = append(r, ch)
	}
	return &chordLex{ input: r }
}

const lexEOF = rune(-1)

func (l *chordLex) next() rune {
	var c rune = ' '
	for c == ' ' {
		if l.pos == len(l.input) {
			return lexEOF
		}
		c = l.input[l.pos]
		l.pos += 1
	}
	return c
}

func (l *chordLex) peek(dist int) rune {
	if l.pos + dist >= len(l.input) {
		return lexEOF
	}
	return l.input[l.pos + dist]
}

func (l *chordLex) skip(dist int) {
	l.pos += dist
}

func (l *chordLex) Lex(lval *chordSymType) int {
	c := l.next()
	if c == lexEOF {
		return 0
	}

	if c >= 'A' && c <= 'G' {
		lval.b = int8(c)
		return _SYM_NOTE
	} else {
		switch (c) {
		case 's':
			if l.peek(0) == 'u' && l.peek(1) == 's' {
				l.skip(2)
				return _SYM_SUS
			}
		case '#', '♯':
			lval.acc = Sharp
			return _SYM_ACCIDENTAL
		case 'x', '𝄪':
			lval.acc = DblSharp
			return _SYM_ACCIDENTAL
		case '♭':
			lval.acc = Flat
			return _SYM_ACCIDENTAL
		case '𝄫':
			lval.acc = DblFlat
			return _SYM_ACCIDENTAL
		case 'b':
			if l.peek(0) == 'b' {
				l.skip(1)
				lval.acc = DblFlat
				return _SYM_ACCIDENTAL
			}
			lval.acc = Flat
			return _SYM_ACCIDENTAL
		case 'n', '♮':
			lval.acc = Natural
			return _SYM_ACCIDENTAL
		case 'a':
			if l.peek(0) == 'u' && l.peek(1) == 'g' {
				l.skip(2)
				return _SYM_AUG
			}
		case 'm':
			if l.peek(0) == 'a' && l.peek(1) == 'j' {
				l.skip(2)
				return _SYM_MAJ7
			} else if l.peek(0) == 'i' && l.peek(1) == 'n' {
				l.skip(2)
				return _SYM_MIN
			}
			return _SYM_MIN
		case 'd':
			if l.peek(0) == 'i' && l.peek(1) == 'm' {
				l.skip(2)
				return _SYM_DIM
			}
		case 'ø':
			return _SYM_HDIM
		case 'o':
			return _SYM_FDIM
		case '△', '∆':
			return _SYM_MAJ7
		case '1':
			if l.peek(0) == '1' {
				l.skip(1)
				lval.b = 11
				return _SYM_TONE
			} else if l.peek(0) == '3' {
				l.skip(1)
				lval.b = 13
				return _SYM_TONE
			}
		case '9':
			lval.b = int8(c) - '0'
			return _SYM_TONE
		}
	}
	return int(c)
}

func (l *chordLex) Error(s string) {
	l.err = errors.New(s)
}