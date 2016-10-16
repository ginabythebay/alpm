package alpm

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func newParsed(e, v, r string) parsed {
	return parsed{[]rune(e), []rune(v), []rune(r)}
}

func TestParseEVR(t *testing.T) {
	allCases := []struct {
		in       string
		expected parsed
	}{
		{"13:79-foo", newParsed("13", "79", "foo")},
		{":79-foo", newParsed("0", "79", "foo")},
		{"79-foo", newParsed("0", "79", "foo")},
		{"79", newParsed("0", "79", "")},

		// cases that are not really normal, but we do handle
		{"abc-def", newParsed("0", "abc", "def")},
		{"abc:def-ghi", newParsed("0", "abc:def", "ghi")},
	}

	for _, c := range allCases {
		t.Run(c.in, func(t *testing.T) {
			actual := parseEVR(c.in)
			equals(t, c.expected, actual)
		})
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}
