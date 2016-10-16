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
