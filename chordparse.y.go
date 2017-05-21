//line chordparse.y:2
package chords

import __yyfmt__ "fmt"

//line chordparse.y:3
import (
	"errors"
)

//line chordparse.y:13
type chordSymType struct {
	yys   int
	ch    Chord
	p     Pitch
	t     Tone
	acc   Accidental
	triad triad
	tones []Tone
	b     byte
}

const SYM_NOTE = 57346
const SYM_TONE = 57347
const SYM_MAJ7 = 57348
const SYM_SUS = 57349
const SYM_ACCIDENTAL = 57350
const SYM_MIN = 57351
const SYM_DIM = 57352
const SYM_HDIM = 57353
const SYM_FDIM = 57354
const SYM_AUG = 57355

var chordToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"SYM_NOTE",
	"SYM_TONE",
	"SYM_MAJ7",
	"SYM_SUS",
	"SYM_ACCIDENTAL",
	"SYM_MIN",
	"SYM_DIM",
	"SYM_HDIM",
	"SYM_FDIM",
	"SYM_AUG",
	"'/'",
	"'7'",
	"'-'",
	"'+'",
	"'2'",
	"'4'",
	"'5'",
	"'6'",
}
var chordStatenames = [...]string{}

const chordEofCode = 1
const chordErrCode = 2
const chordInitialStackSize = 16

//line chordparse.y:233

/*  start  of  programs  */

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
	return &chordLex{input: r}
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
	if l.pos+dist >= len(l.input) {
		return lexEOF
	}
	return l.input[l.pos+dist]
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
		switch c {
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

//line yacctab:1
var chordExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const chordNprod = 38
const chordPrivate = 57344

var chordTokenNames []string
var chordStates []string

const chordLast = 90

var chordAct = [...]int{

	6, 49, 50, 48, 5, 28, 4, 18, 30, 34,
	35, 39, 22, 8, 21, 27, 11, 13, 14, 15,
	16, 9, 7, 12, 17, 23, 24, 25, 26, 22,
	37, 19, 38, 20, 44, 10, 47, 45, 1, 36,
	31, 32, 23, 24, 25, 26, 46, 51, 52, 53,
	22, 3, 40, 27, 2, 0, 22, 29, 0, 27,
	33, 31, 32, 23, 24, 25, 26, 31, 32, 23,
	24, 25, 26, 22, 42, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 41, 43, 23, 24, 25, 26,
}
var chordPact = [...]int{

	2, -1000, -10, 7, -3, 2, -1000, 51, 45, 24,
	51, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	68, 66, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, 51, -1000, -1000, 51, 31, -12, -1000,
	-1000, -1000, -17, -1000, -1000, -1000, 51, 51, 51, -1000,
	-1000, -1000, -1000, -1000,
}
var chordPgo = [...]int{

	0, 54, 38, 51, 35, 33, 21, 7, 0, 31,
}
var chordR1 = [...]int{

	0, 2, 2, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 3, 3, 6, 6, 6, 6, 6, 6,
	6, 6, 7, 7, 7, 7, 8, 8, 4, 4,
	9, 9, 9, 9, 9, 5, 5, 5,
}
var chordR2 = [...]int{

	0, 1, 3, 2, 3, 4, 3, 3, 4, 5,
	5, 5, 1, 2, 1, 1, 1, 1, 1, 1,
	1, 1, 2, 3, 2, 3, 0, 2, 1, 2,
	1, 1, 1, 1, 1, 1, 1, 1,
}
var chordChk = [...]int{

	-1000, -2, -1, -3, 4, 14, -8, 15, 6, -6,
	-4, 9, 16, 10, 11, 12, 13, 17, -7, -9,
	-5, 7, 5, 18, 19, 20, 21, 8, 8, -3,
	-8, 16, 17, 15, -8, -8, 15, 6, 8, -8,
	-9, 18, 8, 19, -8, -8, 15, 5, 15, 18,
	19, -8, -8, -8,
}
var chordDef = [...]int{

	0, -2, 1, 26, 12, 0, 3, 26, 26, 26,
	26, 14, 15, 16, 17, 18, 19, 20, 21, 28,
	0, 0, 30, 31, 32, 33, 34, 37, 13, 2,
	4, 35, 36, 26, 7, 6, 26, 0, 37, 27,
	29, 22, 0, 24, 5, 8, 26, 26, 26, 23,
	25, 9, 10, 11,
}
var chordTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 17, 3, 16, 3, 14, 3, 3,
	18, 3, 19, 20, 21, 15,
}
var chordTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13,
}
var chordTok3 = [...]int{
	0,
}

var chordErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	chordDebug        = 0
	chordErrorVerbose = false
)

type chordLexer interface {
	Lex(lval *chordSymType) int
	Error(s string)
}

type chordParser interface {
	Parse(chordLexer) int
	Lookahead() int
}

type chordParserImpl struct {
	lval  chordSymType
	stack [chordInitialStackSize]chordSymType
	char  int
}

func (p *chordParserImpl) Lookahead() int {
	return p.char
}

func chordNewParser() chordParser {
	return &chordParserImpl{}
}

const chordFlag = -1000

func chordTokname(c int) string {
	if c >= 1 && c-1 < len(chordToknames) {
		if chordToknames[c-1] != "" {
			return chordToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func chordStatname(s int) string {
	if s >= 0 && s < len(chordStatenames) {
		if chordStatenames[s] != "" {
			return chordStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func chordErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !chordErrorVerbose {
		return "syntax error"
	}

	for _, e := range chordErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + chordTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := chordPact[state]
	for tok := TOKSTART; tok-1 < len(chordToknames); tok++ {
		if n := base + tok; n >= 0 && n < chordLast && chordChk[chordAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if chordDef[state] == -2 {
		i := 0
		for chordExca[i] != -1 || chordExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; chordExca[i] >= 0; i += 2 {
			tok := chordExca[i]
			if tok < TOKSTART || chordExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if chordExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += chordTokname(tok)
	}
	return res
}

func chordlex1(lex chordLexer, lval *chordSymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = chordTok1[0]
		goto out
	}
	if char < len(chordTok1) {
		token = chordTok1[char]
		goto out
	}
	if char >= chordPrivate {
		if char < chordPrivate+len(chordTok2) {
			token = chordTok2[char-chordPrivate]
			goto out
		}
	}
	for i := 0; i < len(chordTok3); i += 2 {
		token = chordTok3[i+0]
		if token == char {
			token = chordTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = chordTok2[1] /* unknown char */
	}
	if chordDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", chordTokname(token), uint(char))
	}
	return char, token
}

func chordParse(chordlex chordLexer) int {
	return chordNewParser().Parse(chordlex)
}

func (chordrcvr *chordParserImpl) Parse(chordlex chordLexer) int {
	var chordn int
	var chordVAL chordSymType
	var chordDollar []chordSymType
	_ = chordDollar // silence set and not used
	chordS := chordrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	chordstate := 0
	chordrcvr.char = -1
	chordtoken := -1 // chordrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		chordstate = -1
		chordrcvr.char = -1
		chordtoken = -1
	}()
	chordp := -1
	goto chordstack

ret0:
	return 0

ret1:
	return 1

chordstack:
	/* put a state and value onto the stack */
	if chordDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", chordTokname(chordtoken), chordStatname(chordstate))
	}

	chordp++
	if chordp >= len(chordS) {
		nyys := make([]chordSymType, len(chordS)*2)
		copy(nyys, chordS)
		chordS = nyys
	}
	chordS[chordp] = chordVAL
	chordS[chordp].yys = chordstate

chordnewstate:
	chordn = chordPact[chordstate]
	if chordn <= chordFlag {
		goto chorddefault /* simple state */
	}
	if chordrcvr.char < 0 {
		chordrcvr.char, chordtoken = chordlex1(chordlex, &chordrcvr.lval)
	}
	chordn += chordtoken
	if chordn < 0 || chordn >= chordLast {
		goto chorddefault
	}
	chordn = chordAct[chordn]
	if chordChk[chordn] == chordtoken { /* valid shift */
		chordrcvr.char = -1
		chordtoken = -1
		chordVAL = chordrcvr.lval
		chordstate = chordn
		if Errflag > 0 {
			Errflag--
		}
		goto chordstack
	}

chorddefault:
	/* default state action */
	chordn = chordDef[chordstate]
	if chordn == -2 {
		if chordrcvr.char < 0 {
			chordrcvr.char, chordtoken = chordlex1(chordlex, &chordrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if chordExca[xi+0] == -1 && chordExca[xi+1] == chordstate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			chordn = chordExca[xi+0]
			if chordn < 0 || chordn == chordtoken {
				break
			}
		}
		chordn = chordExca[xi+1]
		if chordn < 0 {
			goto ret0
		}
	}
	if chordn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			chordlex.Error(chordErrorMessage(chordstate, chordtoken))
			Nerrs++
			if chordDebug >= 1 {
				__yyfmt__.Printf("%s", chordStatname(chordstate))
				__yyfmt__.Printf(" saw %s\n", chordTokname(chordtoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for chordp >= 0 {
				chordn = chordPact[chordS[chordp].yys] + chordErrCode
				if chordn >= 0 && chordn < chordLast {
					chordstate = chordAct[chordn] /* simulate a shift of "error" */
					if chordChk[chordstate] == chordErrCode {
						goto chordstack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if chordDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", chordS[chordp].yys)
				}
				chordp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if chordDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", chordTokname(chordtoken))
			}
			if chordtoken == chordEofCode {
				goto ret1
			}
			chordrcvr.char = -1
			chordtoken = -1
			goto chordnewstate /* try again in the same state */
		}
	}

	/* reduction by production chordn */
	if chordDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", chordn, chordStatname(chordstate))
	}

	chordnt := chordn
	chordpt := chordp
	_ = chordpt // guard against "declared and not used"

	chordp -= chordR2[chordn]
	// chordp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if chordp+1 >= len(chordS) {
		nyys := make([]chordSymType, len(chordS)*2)
		copy(nyys, chordS)
		chordS = nyys
	}
	chordVAL = chordS[chordp+1]

	/* consult goto table to find next state */
	chordn = chordR1[chordn]
	chordg := chordPgo[chordn]
	chordj := chordg + chordS[chordp].yys + 1

	if chordj >= chordLast {
		chordstate = chordAct[chordg]
	} else {
		chordstate = chordAct[chordj]
		if chordChk[chordstate] != -chordn {
			chordstate = chordAct[chordg]
		}
	}
	// dummy call; replaced with literal code
	switch chordnt {

	case 1:
		chordDollar = chordS[chordpt-1 : chordpt+1]
		//line chordparse.y:41
		{
			chordVAL.ch = chordDollar[1].ch
			chordlex.(*chordLex).res = &chordVAL.ch
		}
	case 2:
		chordDollar = chordS[chordpt-3 : chordpt+1]
		//line chordparse.y:46
		{
			chordVAL.ch = chordDollar[1].ch
			chordVAL.ch.Bass = chordDollar[3].p
			chordlex.(*chordLex).res = &chordVAL.ch
		}
	case 3:
		chordDollar = chordS[chordpt-2 : chordpt+1]
		//line chordparse.y:55
		{
			chordVAL.ch = Chord{Root: chordDollar[1].p, ExtraTones: chordDollar[2].tones}
		}
	case 4:
		chordDollar = chordS[chordpt-3 : chordpt+1]
		//line chordparse.y:59
		{
			chordVAL.ch = Chord{Root: chordDollar[1].p, ExtraTones: append(chordDollar[3].tones, Tone{Val: 7})}
		}
	case 5:
		chordDollar = chordS[chordpt-4 : chordpt+1]
		//line chordparse.y:63
		{
			chordVAL.ch = Chord{Root: chordDollar[1].p, ExtraTones: append(chordDollar[4].tones, Tone{Val: 7, Acc: SHARP})}
		}
	case 6:
		chordDollar = chordS[chordpt-3 : chordpt+1]
		//line chordparse.y:67
		{
			if chordDollar[2].triad.susTone.Val != 0 {
				chordVAL.ch = Chord{Root: chordDollar[1].p, Triad: chordDollar[2].triad.typ, ExtraTones: append(chordDollar[3].tones, chordDollar[2].triad.susTone)}
			} else {
				chordVAL.ch = Chord{Root: chordDollar[1].p, Triad: chordDollar[2].triad.typ, ExtraTones: chordDollar[3].tones}
			}
		}
	case 7:
		chordDollar = chordS[chordpt-3 : chordpt+1]
		//line chordparse.y:75
		{
			hasHighTone := false
			for _, tn := range chordDollar[3].tones {
				if tn.Val > 7 {
					hasHighTone = true
					break
				}
			}
			if hasHighTone {
				chordVAL.ch = Chord{Root: chordDollar[1].p, ExtraTones: append(chordDollar[3].tones, Tone{Val: 7, Acc: SHARP})}
			} else {
				chordVAL.ch = Chord{Root: chordDollar[1].p, ExtraTones: chordDollar[3].tones}
			}
		}
	case 8:
		chordDollar = chordS[chordpt-4 : chordpt+1]
		//line chordparse.y:90
		{
			if chordDollar[2].triad.susTone.Val != 0 {
				chordVAL.ch = Chord{Root: chordDollar[1].p, Triad: chordDollar[2].triad.typ, ExtraTones: append(chordDollar[4].tones, Tone{Val: 7}, chordDollar[2].triad.susTone)}
			} else {
				chordVAL.ch = Chord{Root: chordDollar[1].p, Triad: chordDollar[2].triad.typ, ExtraTones: append(chordDollar[4].tones, Tone{Val: 7})}
			}
		}
	case 9:
		chordDollar = chordS[chordpt-5 : chordpt+1]
		//line chordparse.y:98
		{
			if chordDollar[2].triad.susTone.Val != 0 {
				chordVAL.ch = Chord{Root: chordDollar[1].p, Triad: chordDollar[2].triad.typ, ExtraTones: append(chordDollar[5].tones, Tone{Val: 7, Acc: SHARP}, chordDollar[2].triad.susTone)}
			} else {
				chordVAL.ch = Chord{Root: chordDollar[1].p, Triad: chordDollar[2].triad.typ, ExtraTones: append(chordDollar[5].tones, Tone{Val: 7, Acc: SHARP})}
			}
		}
	case 10:
		chordDollar = chordS[chordpt-5 : chordpt+1]
		//line chordparse.y:106
		{
			if chordDollar[2].triad.susTone.Val != 0 {
				chordVAL.ch = Chord{Root: chordDollar[1].p, Triad: chordDollar[2].triad.typ, ExtraTones: append(chordDollar[5].tones, Tone{Val: 7, Acc: SHARP}, Tone{Val: chordDollar[4].b}, chordDollar[2].triad.susTone)}
			} else {
				chordVAL.ch = Chord{Root: chordDollar[1].p, Triad: chordDollar[2].triad.typ, ExtraTones: append(chordDollar[5].tones, Tone{Val: 7, Acc: SHARP}, Tone{Val: chordDollar[4].b})}
			}
		}
	case 11:
		chordDollar = chordS[chordpt-5 : chordpt+1]
		//line chordparse.y:114
		{
			if chordDollar[2].triad.susTone.Val != 0 {
				chordVAL.ch = Chord{Root: chordDollar[1].p, Triad: chordDollar[2].triad.typ, ExtraTones: append(chordDollar[5].tones, Tone{Val: 7, Acc: chordDollar[3].acc}, chordDollar[2].triad.susTone)}
			} else {
				chordVAL.ch = Chord{Root: chordDollar[1].p, Triad: chordDollar[2].triad.typ, ExtraTones: append(chordDollar[5].tones, Tone{Val: 7, Acc: chordDollar[3].acc})}
			}
		}
	case 12:
		chordDollar = chordS[chordpt-1 : chordpt+1]
		//line chordparse.y:123
		{
			chordVAL.p = Pitch{N: Note(chordDollar[1].b)}
		}
	case 13:
		chordDollar = chordS[chordpt-2 : chordpt+1]
		//line chordparse.y:127
		{
			chordVAL.p = Pitch{N: Note(chordDollar[1].b), Acc: chordDollar[2].acc}
		}
	case 14:
		chordDollar = chordS[chordpt-1 : chordpt+1]
		//line chordparse.y:132
		{
			chordVAL.triad = triad{typ: MIN3}
		}
	case 15:
		chordDollar = chordS[chordpt-1 : chordpt+1]
		//line chordparse.y:136
		{
			chordVAL.triad = triad{typ: MIN3}
		}
	case 16:
		chordDollar = chordS[chordpt-1 : chordpt+1]
		//line chordparse.y:140
		{
			chordVAL.triad = triad{typ: DIM3}
		}
	case 17:
		chordDollar = chordS[chordpt-1 : chordpt+1]
		//line chordparse.y:144
		{
			chordVAL.triad = triad{typ: HDIM}
		}
	case 18:
		chordDollar = chordS[chordpt-1 : chordpt+1]
		//line chordparse.y:148
		{
			chordVAL.triad = triad{typ: FDIM}
		}
	case 19:
		chordDollar = chordS[chordpt-1 : chordpt+1]
		//line chordparse.y:152
		{
			chordVAL.triad = triad{typ: AUG3}
		}
	case 20:
		chordDollar = chordS[chordpt-1 : chordpt+1]
		//line chordparse.y:156
		{
			chordVAL.triad = triad{typ: AUG3}
		}
	case 21:
		chordDollar = chordS[chordpt-1 : chordpt+1]
		//line chordparse.y:160
		{
			chordVAL.triad = chordDollar[1].triad
		}
	case 22:
		chordDollar = chordS[chordpt-2 : chordpt+1]
		//line chordparse.y:165
		{
			chordVAL.triad = triad{typ: SUS, susTone: Tone{Val: 2}}
		}
	case 23:
		chordDollar = chordS[chordpt-3 : chordpt+1]
		//line chordparse.y:169
		{
			chordVAL.triad = triad{typ: SUS, susTone: Tone{Val: 2, Acc: chordDollar[2].acc}}
		}
	case 24:
		chordDollar = chordS[chordpt-2 : chordpt+1]
		//line chordparse.y:173
		{
			chordVAL.triad = triad{typ: SUS, susTone: Tone{Val: 4}}
		}
	case 25:
		chordDollar = chordS[chordpt-3 : chordpt+1]
		//line chordparse.y:177
		{
			chordVAL.triad = triad{typ: SUS, susTone: Tone{Val: 4, Acc: chordDollar[2].acc}}
		}
	case 26:
		chordDollar = chordS[chordpt-0 : chordpt+1]
		//line chordparse.y:182
		{
			chordVAL.tones = nil
		}
	case 27:
		chordDollar = chordS[chordpt-2 : chordpt+1]
		//line chordparse.y:186
		{
			chordVAL.tones = append([]Tone{chordDollar[1].t}, chordDollar[2].tones...)
		}
	case 28:
		chordDollar = chordS[chordpt-1 : chordpt+1]
		//line chordparse.y:191
		{
			chordVAL.t = Tone{Val: chordDollar[1].b}
		}
	case 29:
		chordDollar = chordS[chordpt-2 : chordpt+1]
		//line chordparse.y:195
		{
			chordVAL.t = Tone{Val: chordDollar[2].b, Acc: chordDollar[1].acc}
		}
	case 30:
		chordDollar = chordS[chordpt-1 : chordpt+1]
		//line chordparse.y:200
		{
			chordVAL.b = chordDollar[1].b
		}
	case 31:
		chordDollar = chordS[chordpt-1 : chordpt+1]
		//line chordparse.y:204
		{
			chordVAL.b = 2
		}
	case 32:
		chordDollar = chordS[chordpt-1 : chordpt+1]
		//line chordparse.y:208
		{
			chordVAL.b = 4
		}
	case 33:
		chordDollar = chordS[chordpt-1 : chordpt+1]
		//line chordparse.y:212
		{
			chordVAL.b = 5
		}
	case 34:
		chordDollar = chordS[chordpt-1 : chordpt+1]
		//line chordparse.y:216
		{
			chordVAL.b = 6
		}
	case 35:
		chordDollar = chordS[chordpt-1 : chordpt+1]
		//line chordparse.y:221
		{
			chordVAL.acc = FLAT
		}
	case 36:
		chordDollar = chordS[chordpt-1 : chordpt+1]
		//line chordparse.y:225
		{
			chordVAL.acc = SHARP
		}
	case 37:
		chordDollar = chordS[chordpt-1 : chordpt+1]
		//line chordparse.y:229
		{
			chordVAL.acc = chordDollar[1].acc
		}
	}
	goto chordstack /* stack new state and value */
}