// Command chordreader is a command-line program that spells chords. The
// chord names are given as command-line args. The program fails if an
// invalid chord name is given.
//
// The program parses the chord names, computes a canonical name, and then
// spells the chord, printing out all of its constituent tones.
//
// Valid chord names must first indicate their root tone as: 'A'-'G' (must be
// capital) followed by an optional 'n', '♮', '#', '♯', 'b', '♭', 'x', '𝄪',
// 'bb', or '𝄫'. The root tone may be followed by a triad indicator (major if
// omitted): '-', 'm', or 'min' for minor; 'aug' or '+' for augmented; 'dim' for
// diminished (fully-diminished if 7th is present); 'ø' for half-diminished
// (implies the 7th); or 'o' for fully diminished (implies a flat 7th, aka minor
// 6th/13th).
//
// For four+ part chords, the next symbol is usually a '7' with an optional
// modifier indicating a major/sharp 7th: 'maj', '∆', '△', '#', or '♯'. This may
// be followed by additional tones, '2', '4', '5', '6', '9', '11', and/or '13',
// each of which may be preceded by an accidental. Presence of such a subsequent
// tone that is greater than 7 (e.g 9, 11, 13) implies presence of the 7th.
//
// A 'sus' can be used in place of a triad indicator to mean that the 3rd is
// omitted. The 'sus' is followed by a '2' or '4', with an optional sharp (for
// 4) or flat (for 2) modifier in between, to indicate which note replaces the
// 3rd.
//
// A chord name can end with a bass tone, indicated by a '/' followed by the
// bass tone (same syntax as the chord's root tone).
package main

import (
	"fmt"
	"os"
	"path"

	"github.com/jhump/chords"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Usage:")
		fmt.Printf("  %s chord...\n", path.Base(os.Args[0]))
		fmt.Println(`
Each argument is a chord. Each chord will be spelled out and its canonical name
printed.

Valid chords must first indicate their root tone as: 'A'-'G' (must be capital)
followed by an optional 'n', '♮', '#', '♯', 'b', '♭', 'x', '𝄪', 'bb', or '𝄫'.
The root tone may be followed by a triad indicator (major if omitted): '-',
'm', or 'min' for minor; 'aug' or '+' for augmented; 'dim' for diminished
(fully-diminished if 7th is present); 'ø' for half-diminished (implies the
7th); or 'o' for fully diminished (implies a flat 7th, aka minor 6th/13th).

For four+ part chords, the next symbol is usually a '7' with an optional
modifier indicating a major/sharp 7th: 'maj', '∆', '△', '#', or '♯'. This may
be followed by additional tones, '2', '4', '5', '6', '9', '11', and/or '13',
each of which may be preceded by an accidental. Presence of such a subsequent
tone that is greater than 7 (e.g 9, 11, 13) implies presence of the 7th.

A 'sus' can be used in place of a triad indicator to mean that the 3rd is
omitted. The 'sus' is followed by a '2' or '4', with an optional sharp (for 4)
or flat (for 2) modifier in between, to indicate which note replaces the 3rd.

A chord can end with a bass tone, indicated by a '/' followed by the bass tone
(same syntax as the chord's root tone).`)
	}

	chs := map[string]*chords.Chord{}
	for _, s := range args {
		ch, err := chords.ParseChord(s)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse %q as a chord: %v\n", s, err)
			os.Exit(1)
		}
		err = ch.Validate()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse %q as a chord: %v\n", s, err)
			os.Exit(1)
		}
		chs[s] = ch
		ch.Canonicalize()
		fmt.Printf("%s => %v: %v\n", s, ch, ch.Spell())
	}
}
