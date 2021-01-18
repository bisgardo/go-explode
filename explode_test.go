package explode

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

/* TRIVIAL CASES */

func Test__simple_string_expands_to_itself(t *testing.T) {
	asser(t,
		that("", expandsTo("")),
		that("a", expandsTo("a")),
		that("ab", expandsTo("ab")),
		that("abc", expandsTo("abc")),
	)
}

func Test__groups_of_empty_strings_expand_to_empty_strings(t *testing.T) {
	asser(t,
		that("{}", expandsTo("")),
		that("{{}}", expandsTo("")),
		that("{,}", expandsTo("", "")),
		that("{{},{}}", expandsTo("", "")),
		that("{,}{,}", expandsTo("", "", "", "")),
		that("{{},{}}{}{{},{}}", expandsTo("", "", "", "")),
	)
}

func Test__empty_groups_expand_to_empty(t *testing.T) {
	asser(t,
		that("a{}", expandsTo("a")),
		that("{}b", expandsTo("b")),
		that("a{}b", expandsTo("ab")),
		that("{}a{}b{}", expandsTo("ab")),
	)
}

func Test__singleton_groups_expand_to_substring(t *testing.T) {
	asser(t,
		that("{a}", expandsTo("a")),
		that("{ab}", expandsTo("ab")),
		that("{{a}}", expandsTo("a")),
		that("a{b}", expandsTo("ab")),
		that("{a}b", expandsTo("ab")),
		that("{a}{b}", expandsTo("ab")),
		that("{a}b{c}", expandsTo("abc")),
		that("a{b}c", expandsTo("abc")),
		that("ab{cd}ef", expandsTo("abcdef")),
	)
}

/* SIMPLE CASES */

func Test__single_nonempty_group_expands_to_all_substrings(t *testing.T) {
	asser(t,
		that("{,a}", expandsTo("", "a")),
		that("{a,}", expandsTo("a", "")),
		that("{a,b}", expandsTo("a", "b")),
		that("{{a,b}}", expandsTo("a", "b")),
		that("a{b,c}", expandsTo("ab", "ac")),
		that("{a,b}c", expandsTo("ac", "bc")),
		that("a{b,c}d", expandsTo("abd", "acd")),
		that("{a{b,c}d}", expandsTo("abd", "acd")),
	)
}

func Test__single_nonempty_group_with_nested_singleton_groups_expands_to_all_substrings(t *testing.T) {
	asser(t,
		that("{a{b},c{d},e{f}}", expandsTo("ab", "cd", "ef")),
		that("{{a}b,{c}d,{e}f}", expandsTo("ab", "cd", "ef")),
		that("{{a,b},{c,d}}", expandsTo("a", "b", "c", "d")),
	)
}

func Test__group_pair_expands_to_all_combinations(t *testing.T) {
	asser(t,
		that("{a,b}{c,d}", expandsTo("ac", "ad", "bc", "bd")),
		that("{{a,b}{c,d}}", expandsTo("ac", "ad", "bc", "bd")),
		that("{abc,def}{ghi,jkl}", expandsTo("abcghi", "abcjkl", "defghi", "defjkl")),
		that("a{b,c}e{f,g}h", expandsTo("abefh", "abegh", "acefh", "acegh")),
	)
}

/* COMPLEX CASES */

func Test__nested_groups_expand_locally(t *testing.T) {
	asser(t,
		that("a{b{c,d}e,f{g,h}i}j", expandsTo("abcej", "abdej", "afgij", "afhij")),
		that("a{{b,c}{d,e},{f,g}{h,i}}j", expandsTo("abdj", "abej", "acdj", "acej", "afhj", "afij", "aghj", "agij")),
	)
}

func Test__group_triple_expands_to_all_combinations(t *testing.T) {
	asser(t,
		that("{a,b}{c,d}{e,f}", expandsTo("ace", "acf", "ade", "adf", "bce", "bcf", "bde", "bdf")),
		that("{a,b}c{d,e}f{g,h}", expandsTo("acdfg", "acdfh", "acefg", "acefh", "bcdfg", "bcdfh", "bcefg", "bcefh")),
		that("{a,b}{{c,d}{e,f},{g,h}{i,j}}{k,l}", expandsTo(
			"acek", "acel", "acfk", "acfl", "adek", "adel", "adfk", "adfl",
			"agik", "agil", "agjk", "agjl", "ahik", "ahil", "ahjk", "ahjl",
			"bcek", "bcel", "bcfk", "bcfl", "bdek", "bdel", "bdfk", "bdfl",
			"bgik", "bgil", "bgjk", "bgjl", "bhik", "bhil", "bhjk", "bhjl",
		)),
	)
}

func Test__errors_are_reported_at_correct_location(t *testing.T) {
	asser(t,
		that(",", expandsTo(",")).with(error(0, 0)),
		that("a,b", expandsTo("a,b")).with(error(1, 0)),
		that("}", expandsTo("")).with(error(0, '{')),
		that("{", expandsTo("")).with(error(0, '}')),
		that("}{", expandsTo("")).with(error(0, '{'), error(1, '}')),
		that("{{", expandsTo("")).with(error(1, '}'), error(0, '}')),
		that("}}", expandsTo("")).with(error(0, '{'), error(1, '{')),
		that("}}{{", expandsTo("")).with(error(0, '{'), error(1, '{'), error(3, '}'), error(2, '}')),
		that("}{}{", expandsTo("")).with(error(0, '{'), error(3, '}')),
		that(",{a,b},", expandsTo(",a,", ",b,")).with(error(0, 0), error(6, 0)),
		that("{a,b},{c,d}", expandsTo("a,c", "a,d", "b,c", "b,d")).with(error(5, 0)),
		that("a}b{c}d{e", expandsTo("abcde")).with(error(1, '{'), error(7, '}')),
		that("a,d}e{fg},h}", expandsTo("a,defg,h")).with(error(1, 0), error(3, '{'), error(9, 0), error(11, '{')),
	)
}

/* RUNNER UTILS */

func testReportError(t *testing.T, expectedErrs []wantError) (ReportErrorFunc, *int) {
	errCount := 0
	return func(pos int, msg string) {
		if errCount >= len(expectedErrs) {
			t.Errorf("unexpected error: %d: %s", pos, msg)
			return
		}

		w := expectedErrs[errCount]
		errCount++
		assert.Equal(t, w.pos, pos)
		if w.unmatched == 0 {
			assert.Equal(t, "invalid separator", msg)
		} else {
			assert.Equal(t, fmt.Sprintf("no matching '%c'", w.unmatched), msg)
		}
	}, &errCount
}

type wantError struct {
	pos       int
	unmatched rune
}

func error(pos int, unmatched rune) wantError {
	return wantError{
		pos:       pos,
		unmatched: unmatched,
	}
}

type spec struct {
	expr     string
	wantRes  []string
	wantErrs []wantError
}

func (s spec) with(errs ...wantError) spec {
	return spec{
		expr:     s.expr,
		wantRes:  s.wantRes,
		wantErrs: errs,
	}
}

func asser(t *testing.T, ss ...spec) {
	for _, s := range ss {
		t.Run(s.expr, func(t *testing.T) {
			reportError, errCount := testReportError(t, s.wantErrs)
			res := Explode(s.expr, reportError)
			assert.Equal(t, s.wantRes, res)
			assert.Equal(t, len(s.wantErrs), *errCount)
		})
	}
}

func that(expr string, want []string) spec {
	return spec{expr: expr, wantRes: want}
}

func expandsTo(s ...string) []string {
	return s
}
