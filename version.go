package alpm

import "unicode"

// Note: this is modeled very closely on version.c in order to make it
// easy to compare the two implementations.

type parsed struct {
	epoch   []rune
	version []rune
	release []rune
}

func parseEVR(text string) parsed {
	epoch := []rune{}
	version := []rune{}
	release := []rune{}

	evr := []rune(text)
	l := len(evr)

	var s int // points to epoch terminator
	se := -1  // points to version terminator
	for i := l - 1; i >= 0; i-- {
		if evr[i] == '-' {
			se = i
			break
		}
	}

	for s = 0; s+1 < l && unicode.IsDigit(evr[s]); s++ {
	}

	var versionStart int
	if evr[s] == ':' {
		epoch = evr[:s]
		versionStart = s + 1
		if len(epoch) == 0 {
			epoch = []rune("0")
		}
	} else {
		epoch = []rune("0")
		versionStart = 0
	}
	if se != -1 {
		release = evr[se+1:]
		version = evr[versionStart:se]
	} else {
		version = evr[versionStart:]
	}

	return parsed{epoch, version, release}
}

type tokenType int

const (
	sepType     tokenType = iota
	alphaType             // segment composed of letters
	numericType           // segment composed of digits
)

type token struct {
	kind tokenType
	text []rune
}

func (t token) stripLeadingZeros() []rune {
	for i, r := range t.text {
		if r != '0' {
			return t.text[i:]
		}
	}
	return t.text[0:0]
}

func runeKind(r rune) tokenType {
	switch {
	case unicode.IsLetter(r):
		return alphaType
	case unicode.IsDigit(r):
		return numericType
	default:
		return sepType
	}
}

type separator int

func nextSeparator(source []rune) (sep separator, remaining []rune) {
	if len(source) == 0 {
		return 0, source
	}
	for i, r := range source {
		if runeKind(r) != sepType {
			return separator(i), source[i:]
		}
	}
	return separator(len(source)), source[0:0]
}

func (s separator) len() int {
	return int(s)
}

func nextToken(source []rune) (t token, remaining []rune) {
	if len(source) == 0 {
		return token{sepType, source}, source
	}
	kind := runeKind(source[0])
	for i, r := range source {
		k := runeKind(r)
		if k != kind {
			return token{kind, source[:i]}, source[i:]
		}
	}
	return token{kind, source}, source[0:0]
}

// like C strcmp, but for slices of runes
func runeCmp(one, two []rune) int {
	oneLen := len(one)
	twoLen := len(two)
	i := 0
	for {
		if i >= oneLen && i >= twoLen {
			// We have compared all runes and both runs ended at the
			// same time.
			return 0
		}
		if i >= oneLen || i >= twoLen {
			if twoLen > oneLen {
				return 1
			}
			return -1
		}

		rOne := one[i]
		rTwo := two[i]
		if rOne < rTwo {
			return 1
		}
		if rTwo < rOne {
			return -1
		}

		i++
	}
}

// rpmVerCmp compares epoch or version or release, deciding which is
// newer, similar to rpmvercmp in version.c.
//   0 if one and two are equal
//  -1 if one is older
//   1 if two is newer
//
// We expect to find segments (runs of letters or runs of digits)
// possibly separated by separators, which are characters that are not
// letters or digits.
//
// When comparing separators, we just compare length, with the longer
// one being newer.
//
// Digit segments are always newer than letter segments or empty
// segments.
//
// When comparing letter segments, we do the equivalent of C strcmp.
//
// When comparing digit segments, we strip leading zeros.  If the
// lengths are different, then the longest one is newer.  Otherwise
// treat as letter segment.
//
// If we have looped through all segments, we have special logic for
// the final pair:
//	 if one is empty and two is numeric, two is newer.
//	 if one is an alpha, two is newer.
//	 otherwise one is newer.
func rpmVerCmp(one, two []rune) int {
	for {
		var sepOne, sepTwo separator
		sepOne, one = nextSeparator(one)
		sepTwo, two = nextSeparator(two)

		// If we ran to the end of either, we are finished with the loop
		if len(one) == 0 || len(two) == 0 {
			break
		}

		// If the separator lengths were different, we are also finished
		if sepOne.len() != sepTwo.len() {
			if sepOne.len() < sepTwo.len() {
				return 1
			}
			return -1
		}

		var tokOne, tokTwo token
		tokOne, one = nextToken(one)
		tokTwo, two = nextToken(two)

		// this cannot happen, as we previously tested to make sure that
		// the first string has a non-null segment
		if tokOne.kind == sepType {
			return -1 // arbitrary
		}

		// take care of the case where the two version segments are
		// different types: one numeric, the other alpha.  Numeric
		// segments are always newer than alpha segments XXX See patch
		// #60884 (and details) from bugzilla #50977.
		if tokOne.kind != tokTwo.kind {
			if tokOne.kind == numericType {
				return -1
			}
			return 1
		}

		if tokOne.kind == numericType {
			// this used to be done by converting the digit segments
			// to ints using atoi() - it's changed because long
			// digit segments can overflow an int - this should fix that.

			// throw away any leading zeros - it's a number, right?
			tokOne.text = tokOne.stripLeadingZeros()
			tokTwo.text = tokTwo.stripLeadingZeros()

			// whichever number has more digits wins
			if len(tokOne.text) < len(tokTwo.text) {
				return 1
			}
			if len(tokTwo.text) < len(tokOne.text) {
				return -1
			}
		}

		rc := runeCmp(tokOne.text, tokTwo.text)
		if rc != 0 {
			return rc
		}
	}

	// This catches the case where all numeric and alpha segments have
	// compared identically and we ran out of more segments.
	if len(one) == 0 || len(two) == 0 {
		return 0
	}

	/* the final showdown. we never want a remaining alpha string to
	 * beat an empty string. the logic is a bit weird, but:
	 * - if one is empty and two is not an alpha, two is newer.
	 * - if one is an alpha, two is newer.
	 * - otherwise one is newer.
	 * */
	var tokOne, tokTwo token
	tokOne, _ = nextToken(one)
	tokTwo, _ = nextToken(two)
	if (tokOne.kind == sepType && tokTwo.kind != alphaType) || tokOne.kind == alphaType {
		return 1
	}
	return -1
}

// VerCmp compares two arch linux package version strings.  It returns:
//   0 if the versions are equal
//  -1 if a is older
//   1 if a is newer
//
// Note that this is not semantic versioning.
//
// Based on the contents of pacman/lib/libalpm/version.c, version
// v5.0.1.  This was apparently equivalent to rpmvercmp, in rpm
// version 4.8.1.
func VerCmp(a, b string) int {
	return 0
}
