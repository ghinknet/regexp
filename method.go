package regexp

import (
	"io"
	"regexp"
	"regexp/syntax"
	"strconv"
)

func compile(expr string, mode syntax.Flags) (*Regexp, error) {
	// Read lock
	l.RLock()

	// Construct regexp record
	r := record{
		Re:   expr,
		Mode: mode,
	}

	// Try to get existed regexp
	if re, ok := cached[r]; ok {
		l.RUnlock()
		return re, nil
	}
	l.RUnlock()

	// RW lock
	l.Lock()
	defer l.Unlock()

	// Compile a new one
	switch mode {
	case syntax.POSIX:
		compiled, err := regexp.CompilePOSIX(expr)
		if err != nil {
			return nil, err
		}
		cached[r] = compiled
	case syntax.Perl:
		fallthrough
	default:
		compiled, err := regexp.Compile(expr)
		if err != nil {
			return nil, err
		}
		cached[r] = compiled
	}
	return cached[r], nil
}

// Compile parses a regular expression and returns, if successful,
// a [Regexp] object that can be used to match against text.
//
// When matching against text, the regexp returns a match that
// begins as early as possible in the input (leftmost), and among those
// it chooses the one that a backtracking search would have found first.
// This so-called leftmost-first matching is the same semantics
// that Perl, Python, and other implementations use, although this
// package implements it without the expense of backtracking.
// For POSIX leftmost-longest matching, see [CompilePOSIX].
func Compile(expr string) (*Regexp, error) {
	return compile(expr, syntax.Perl)
}

// CompilePOSIX is like [Compile] but restricts the regular expression
// to POSIX ERE (egrep) syntax and changes the match semantics to
// leftmost-longest.
//
// That is, when matching against text, the regexp returns a match that
// begins as early as possible in the input (leftmost), and among those
// it chooses a match that is as long as possible.
// This so-called leftmost-longest matching is the same semantics
// that early regular expression implementations used and that POSIX
// specifies.
//
// However, there can be multiple leftmost-longest matches, with different
// submatch choices, and here this package diverges from POSIX.
// Among the possible leftmost-longest matches, this package chooses
// the one that a backtracking search would have found first, while POSIX
// specifies that the match be chosen to maximize the length of the first
// subexpression, then the second, and so on from left to right.
// The POSIX rule is computationally prohibitive and not even well-defined.
// See https://swtch.com/~rsc/regexp/regexp2.html#posix for details.
func CompilePOSIX(expr string) (*Regexp, error) {
	return compile(expr, syntax.POSIX)
}

func quote(s string) string {
	if strconv.CanBackquote(s) {
		return "`" + s + "`"
	}
	return strconv.Quote(s)
}

// MustCompile is like [Compile] but panics if the expression cannot be parsed.
// It simplifies safe initialization of global variables holding compiled regular
// expressions.
func MustCompile(str string) *Regexp {
	re, err := Compile(str)
	if err != nil {
		panic(`regexp: Compile(` + quote(str) + `): ` + err.Error())
	}
	return re
}

// MustCompilePOSIX is like [CompilePOSIX] but panics if the expression cannot be parsed.
// It simplifies safe initialization of global variables holding compiled regular
// expressions.
func MustCompilePOSIX(str string) *Regexp {
	re, err := CompilePOSIX(str)
	if err != nil {
		panic(`regexp: CompilePOSIX(` + quote(str) + `): ` + err.Error())
	}
	return re
}

// MatchReader reports whether the text returned by the [io.RuneReader]
// contains any match of the regular expression pattern.
// More complicated queries need to use [Compile] and the full [Regexp] interface.
func MatchReader(pattern string, r io.RuneReader) (matched bool, err error) {
	re, err := Compile(pattern)
	if err != nil {
		return false, err
	}
	return re.MatchReader(r), nil
}

// MatchString reports whether the string s
// contains any match of the regular expression pattern.
// More complicated queries need to use [Compile] and the full [Regexp] interface.
func MatchString(pattern string, s string) (matched bool, err error) {
	re, err := Compile(pattern)
	if err != nil {
		return false, err
	}
	return re.MatchString(s), nil
}

// Match reports whether the byte slice b
// contains any match of the regular expression pattern.
// More complicated queries need to use [Compile] and the full [Regexp] interface.
func Match(pattern string, b []byte) (matched bool, err error) {
	re, err := Compile(pattern)
	if err != nil {
		return false, err
	}
	return re.Match(b), nil
}

// QuoteMeta returns a string that escapes all regular expression metacharacters
// inside the argument text; the returned string is a regular expression matching
// the literal text.
func QuoteMeta(str string) string {
	return regexp.QuoteMeta(str)
}

// Clear all cached regexp
func Clear() {
	cached = make(map[record]*Regexp)
}

func release(expr string, mode syntax.Flags) {
	l.Lock()
	defer l.Unlock()

	// Construct regexp record
	r := record{
		Re:   expr,
		Mode: mode,
	}

	delete(cached, r)
}

// Release a specific Perl regexp
func Release(expr string) {
	release(expr, syntax.Perl)
}

// ReleasePOSIX a specific POSIX regexp
func ReleasePOSIX(expr string) {
	release(expr, syntax.POSIX)
}
