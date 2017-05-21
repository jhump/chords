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
	p     Pitch
	t     Tone
	acc   Accidental
	triad triad
	tones []Tone
	b     byte
}

// any non-terminal which returns a value needs a type, which is
// really a field name in the above union struct
%type <ch>    chord fullChord
%type <p>     pitch
%type <t>     tone
%type <acc>   toneMod
%type <triad> triad sus
%type <tones> extras
%type <b>     toneVal

// same for terminals
%token <b>     SYM_NOTE SYM_TONE SYM_MAJ7 SYM_SUS
%token <acc>   SYM_ACCIDENTAL
%token <triad> SYM_MIN SYM_DIM SYM_HDIM SYM_FDIM SYM_AUG

%%

fullChord	: chord
		{
			$$ = $1
			chordlex.(*chordLex).res = &$$
		}
	| chord '/' pitch
		{
			$$ = $1
			$$.Bass = $3
			chordlex.(*chordLex).res = &$$
		}

// numerous formulations to prevent any ambiguity between an accidental
// adjusting the root pitch vs. the 7th
chord	: pitch extras
		{
			$$ = Chord{ Root: $1, ExtraTones: $2 }
		}
	| pitch '7' extras
		{
			$$ = Chord{ Root: $1, ExtraTones: append($3, Tone{Val: 7}) }
		}
	| pitch SYM_MAJ7 '7' extras
		{
			$$ = Chord{ Root: $1, ExtraTones: append($4, Tone{Val: 7, Acc: SHARP}) }
		}
	| pitch triad extras
		{
			if $2.susTone.Val != 0 {
				$$ = Chord{ Root: $1, Triad: $2.typ, ExtraTones: append($3, $2.susTone) }
			} else {
				$$ = Chord{ Root: $1, Triad: $2.typ, ExtraTones: $3 }
			}
		}
	| pitch SYM_MAJ7 extras
		{
			hasHighTone := false
			for _, tn := range $3 {
				if tn.Val > 7 {
					hasHighTone = true
					break
				}
			}
			if hasHighTone {
				$$ = Chord{ Root: $1, ExtraTones: append($3, Tone{Val:7, Acc:SHARP}) }
			} else {
				$$ = Chord{ Root: $1, ExtraTones: $3 }
			}
		}
	| pitch triad '7' extras
		{
			if $2.susTone.Val != 0 {
				$$ = Chord{ Root: $1, Triad: $2.typ, ExtraTones: append($4, Tone{Val: 7}, $2.susTone) }
			} else {
				$$ = Chord{ Root: $1, Triad: $2.typ, ExtraTones: append($4, Tone{Val: 7}) }
			}
		}
	| pitch triad SYM_MAJ7 '7' extras
		{
			if $2.susTone.Val != 0 {
				$$ = Chord{ Root: $1, Triad: $2.typ, ExtraTones: append($5, Tone{Val: 7, Acc: SHARP}, $2.susTone) }
			} else {
				$$ = Chord{ Root: $1, Triad: $2.typ, ExtraTones: append($5, Tone{Val: 7, Acc: SHARP}) }
			}
		}
	| pitch triad SYM_MAJ7 SYM_TONE extras
		{
			if $2.susTone.Val != 0 {
				$$ = Chord{ Root: $1, Triad: $2.typ, ExtraTones: append($5, Tone{Val: 7, Acc: SHARP}, Tone{Val: $4}, $2.susTone) }
			} else {
				$$ = Chord{ Root: $1, Triad: $2.typ, ExtraTones: append($5, Tone{Val: 7, Acc: SHARP}, Tone{Val: $4}) }
			}
		}
	| pitch triad SYM_ACCIDENTAL '7' extras
		{
			if $2.susTone.Val != 0 {
				$$ = Chord{ Root: $1, Triad: $2.typ, ExtraTones: append($5, Tone{Val: 7, Acc: $3}, $2.susTone) }
			} else {
				$$ = Chord{ Root: $1, Triad: $2.typ, ExtraTones: append($5, Tone{Val: 7, Acc: $3}) }
			}
		}

pitch	: SYM_NOTE
		{
			$$ = Pitch{ N: Note($1) }
		}
	| SYM_NOTE SYM_ACCIDENTAL
		{
			$$ = Pitch{ N: Note($1), Acc: $2 }
		}

triad	: SYM_MIN
		{
			$$ = triad{ typ: MIN3 }
		}
	| '-'
		{
			$$ = triad{ typ: MIN3 }
		}
	| SYM_DIM
		{
			$$ = triad{ typ: DIM3 }
		}
	| SYM_HDIM
		{
			$$ = triad{ typ: HDIM }
		}
	| SYM_FDIM
		{
			$$ = triad{ typ: FDIM }
		}
	| SYM_AUG
		{
			$$ = triad{ typ: AUG3 }
		}
	| '+'
		{
			$$ = triad{ typ: AUG3 }
		}
	| sus
		{
			$$ = $1
		}

sus	: SYM_SUS '2'
		{
			$$ = triad{ typ: SUS, susTone: Tone{ Val: 2 } }
		}
	| SYM_SUS SYM_ACCIDENTAL '2'
		{
			$$ = triad{ typ: SUS, susTone: Tone{ Val: 2, Acc: $2 } }
		}
	| SYM_SUS '4'
		{
			$$ = triad{ typ: SUS, susTone: Tone{ Val: 4 } }
		}
	| SYM_SUS SYM_ACCIDENTAL '4'
		{
			$$ = triad{ typ: SUS, susTone: Tone{ Val: 4, Acc: $2 } }
		}

extras	: /* empty */
		{
			$$ = nil
		}
	| tone extras
		{
			$$ = append([]Tone{$1}, $2...)
		}

tone	: toneVal
		{
			$$ = Tone{ Val: $1 }
		}
	| toneMod toneVal
		{
			$$ = Tone{ Val: $2, Acc: $1 }
		}

toneVal	: SYM_TONE
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
			$$ = FLAT
		}
	| '+'
		{
			$$ = SHARP
		}
	| SYM_ACCIDENTAL
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
		lval.b = byte(c)
		return SYM_NOTE
	} else {
		switch (c) {
		case 's':
			if l.peek(0) == 'u' && l.peek(1) == 's' {
				l.skip(2)
				return SYM_SUS
			}
		case '#', '♯':
			lval.acc = SHARP
			return SYM_ACCIDENTAL
		case 'x', '𝄪':
			lval.acc = DBL_SHARP
			return SYM_ACCIDENTAL
		case '♭':
			lval.acc = FLAT
			return SYM_ACCIDENTAL
		case '𝄫':
			lval.acc = DBL_FLAT
			return SYM_ACCIDENTAL
		case 'b':
			if l.peek(0) == 'b' {
				l.skip(1)
				lval.acc = DBL_FLAT
				return SYM_ACCIDENTAL
			}
			lval.acc = FLAT
			return SYM_ACCIDENTAL
		case 'n', '♮':
			lval.acc = NATURAL
			return SYM_ACCIDENTAL
		case 'a':
			if l.peek(0) == 'u' && l.peek(1) == 'g' {
				l.skip(2)
				return SYM_AUG
			}
		case 'm':
			if l.peek(0) == 'a' && l.peek(1) == 'j' {
				l.skip(2)
				return SYM_MAJ7
			} else if l.peek(0) == 'i' && l.peek(1) == 'n' {
				l.skip(2)
				return SYM_MIN
			}
			return SYM_MIN
		case 'd':
			if l.peek(0) == 'i' && l.peek(1) == 'm' {
				l.skip(2)
				return SYM_DIM
			}
		case 'ø':
			return SYM_HDIM
		case 'o':
			return SYM_FDIM
		case '△', '∆':
			return SYM_MAJ7
		case '1':
			if l.peek(0) == '1' {
				l.skip(1)
				lval.b = 11
				return SYM_TONE
			} else if l.peek(0) == '3' {
				l.skip(1)
				lval.b = 13
				return SYM_TONE
			}
		case '9':
			lval.b = byte(c) - '0'
			return SYM_TONE
		}
	}
	return int(c)
}

func (l *chordLex) Error(s string) {
	l.err = errors.New(s)
}