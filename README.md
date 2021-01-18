# go-explode

Tiny, zero-dependency Go library for performing brace expansion on a string expression.

Brace expansion defines a simple language for describing groups of alternating strings that "explode"
into all combinations of these strings:

Groups are delimited by `{` and `}` and the alternations separated by `,`.
Whitespace around these symbols is not trimmed.
Groups may be nested inside other groups to make the expansion recursive.

The result is ordered such that all expanded strings starting with a given substring in some group are generated
before the next substring in the same group starts expanding.
This appears to be the standard expansion ordering and is also the most intuitive when considering
left- and rightmost groups more and less "significant", respectively (i.e. like numbers).

The idea of brace expansion originates form the shell languages and is probably best known from
[bash](https://www.gnu.org/software/bash/manual/html_node/Brace-Expansion.html).
Compared to the bash implementation, this one doesn't (currently) support `{1..3}` style sequences.

While other implementations are easy to find (including in Go),
this one is production ready by virtue of being polished, fully tested, and extremely fast:
The algorithm is non-recursive and believed to be at least very close to optimal in terms of performance and memory usage.

## Examples

- `fi{nd,ne,sh}` expands to `[find, fine, fish]`
- `r{u,a}{,i}n` expands to `[run, ruin, ran, rain]`
- `s{{a,o}{il,lv},l{ee,o}p}ing` expands to `[sailing, salving, soiling, solving, sleeping, sloping]`

## Usage

#### Command "go get" (legacy)

```
go get github.com/bisgardo/go-explode
```

#### Glide (legacy)

```
glide get github.com/bisgardo/go-explode
```

#### Command "go mod"

Coming soon.
