package explode

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		that("{b,c}", expandsTo("b", "c")),
		that("{{b,c}}", expandsTo("b", "c")),
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
	tests := []struct {
		expr string
		want Error
	}{
		{expr: ",", want: Error{Pos: 0, Missing: 0}},
		{expr: "a,b", want: Error{Pos: 1, Missing: 0}},
		{expr: "}", want: Error{Pos: 0, Missing: '{'}},
		{expr: "{", want: Error{Pos: 0, Missing: '}'}},
		{expr: "}{", want: Error{Pos: 0, Missing: '{'}},
		{expr: "{{", want: Error{Pos: 1, Missing: '}'}},
		{expr: "}}", want: Error{Pos: 0, Missing: '{'}},
		{expr: "}}{{", want: Error{Pos: 0, Missing: '{'}},
		{expr: "}{}{", want: Error{Pos: 0, Missing: '{'}},
		{expr: ",{a,b},", want: Error{Pos: 0, Missing: 0}},
		{expr: "{a,b},{c,d}", want: Error{Pos: 5, Missing: 0}},
		{expr: "a}b{c}d{e", want: Error{Pos: 1, Missing: '{'}},
		{expr: "ab}e{de},f}", want: Error{Pos: 2, Missing: '{'}},
	}

	for _, test := range tests {
		t.Run(test.expr, func(t *testing.T) {
			_, err := Explode(test.expr)
			assert.Equal(t, test.want, err)
		})
	}
}

/* RUNNER UTILS */

type spec struct {
	expr string
	want []string
}

func asser(t *testing.T, ss ...spec) {
	for _, s := range ss {
		t.Run(s.expr, func(t *testing.T) {
			res, err := Explode(s.expr)
			require.NoError(t, err)
			assert.Equal(t, s.want, res)
		})
	}
}

func that(expr string, want []string) spec {
	return spec{expr: expr, want: want}
}

func expandsTo(s ...string) []string {
	return s
}
