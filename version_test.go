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

func rs(s string) []rune {
	if len(s) == 0 {
		return []rune{}
	}
	return []rune(s)
}

func TestNextSeprator(t *testing.T) {
	allCases := []struct {
		in        []rune
		len       int
		remaining []rune
	}{
		{rs("$$13"), 2, rs("13")},
		{rs("$13"), 1, rs("13")},
		{rs("13"), 0, rs("13")},
		{rs(""), 0, rs("")},
	}
	for _, c := range allCases {
		t.Run(string(c.in), func(t *testing.T) {
			sep, remaining := nextSeparator(c.in)
			equals(t, c.len, sep.len())
			equals(t, c.remaining, remaining)
		})
	}
}

func TestNextToken(t *testing.T) {
	allCases := []struct {
		in        []rune
		tok       token
		remaining []rune
	}{
		{rs("13"), token{numericType, rs("13")}, rs("")},
		{rs(""), token{sepType, rs("")}, rs("")},

		{rs("13ab"), token{numericType, rs("13")}, rs("ab")},
		{rs("ab"), token{alphaType, rs("ab")}, rs("")},
	}
	for _, c := range allCases {
		t.Run(string(c.in), func(t *testing.T) {
			tok, remaining := nextToken(c.in)
			equals(t, c.tok, tok)
			equals(t, c.remaining, remaining)
		})
	}
}

func TestRpmVerCmp(t *testing.T) {
	allCases := []struct {
		one      string
		two      string
		expected int
	}{
		// test comparison of just two numbers
		{"1", "2", -1},
		{"2", "1", 1},
		{"2", "2", 0},
		{"02", "2", 0},
		{"02", "2", 0},
		{"020", "2", 1},

		// test comparison of just alpha types
		{"a", "b", -1},
		{"b", "a", 1},
		{"b", "b", 0},
		{"b", "ba", -1},

		// test comparison of just separators
		{"$a", "#b", -1},
		{"$b", "#a", 1},
		{"$$a", "#a", 1},
		{"$a", "##a", -1},

		// test numeric vs. alpha
		{"1", "a", 1},
		{"a", "1", -1},

		// test a few composite cases
		{"1a01", "1a2", -1},
		{"1a1", "1a2", -1},
		{"1a$1", "1a$2", -1},
		{"1a$$1", "1a$2", 1},
	}

	for _, c := range allCases {
		t.Run(fmt.Sprintf("%q:%q", c.one, c.two), func(t *testing.T) {
			actual := rpmVerCmp(rs(c.one), rs(c.two))
			equals(t, c.expected, actual)
		})
	}
}

func TestVerCmp(t *testing.T) {
	allCases := []struct {
		one      string
		two      string
		expected int
	}{
		{"1.5", "1.4", 1},
		{"1.4", "1.5", -1},
		{"1.5", "1.5", 0},
		{"1.5-1", "1.5", 0},
		{"1.5", "1.5-1", 0},
		{"1.5-1", "1.5-2", -1},
		{"1.5-2", "1.5-1", 1},
	}

	for _, c := range allCases {
		t.Run(fmt.Sprintf("%q:%q", c.one, c.two), func(t *testing.T) {
			actual := VerCmp(c.one, c.two)
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
