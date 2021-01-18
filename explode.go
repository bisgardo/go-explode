package explode

import "fmt"

// Error is a syntax error of a brace expansion expression.
type Error struct {
	// Location of the character which triggered the error.
	Pos int
	// Missing match character that it expected to find.
	// If the triggering character is a separator, then this is 0.
	Missing rune
}

func (e Error) Error() string {
	if e.Missing == 0 {
		return fmt.Sprintf("%d: invalid separator", e.Pos)
	}
	return fmt.Sprintf("%d: no matching '%c' found", e.Pos, e.Missing)
}

// Explode interprets a brace expansion expression and returns the list of expanded strings.
// The expansion is computed as combinations of substrings separated by ',' and grouped recursively by '{}'
// (see the README of this library for more details).
// The result is ordered such that all expanded strings starting with a given substring in some group are generated
// before the next substring in the same group starts expanding.
// Examples:
//   fi{nd,ne,sh}                -> [find, fine, fish]
//   r{u,a}{,i}n                 -> [run, ruin, ran, rain]
//   s{{a,o}{il,lv},l{ee,o}p}ing -> [sailing, salving, soiling, solving, sleeping, sloping]
// An unmatched brace or a separator outside any brace pair is considered a syntax error.
// The function returns an appropriate Error immediately after such an error is encountered.
func Explode(expr string) ([]string, error) {
	// TODO Explain what (sub)context means.
	//      - Expansion context: Group delimited by braces.
	//      - Sub-context: separator-delimited component of a given expansion context. ...
	type context struct {
		// Pointer to the context inside of which the context is nested.
		parent *context
		// Index of the character which opened the context.
		offset int
		// Immutable set of prefixes that all substrings in the context will be appended to.
		prefixes []string
		// Mutable set of strings expanded locally in the context.
		// TODO Could prepend prefixes eagerly such that we can just assign head.result to result on close?
		//      Two issues:
		//      1. Looping over prefixes before substring changes the order of the result.
		//      2. It might result in more concatenations - consider "{a,b}{{c,d}e}":
		//         Now e is being appended to "c" and "d" only - not all of "ac", "ad", "bc", and "bd".
		result []string
	}

	// Set of expanded strings within the current sub-context.
	// Each component of an expansion group defines a new sub-context.
	// When a new context is opened, the value of 'results' comprise the set of prefixes that
	// the strings resolved in the context are to be combined with.
	// When the current sub-context is closed, 'result' is flushed into the local result of the containing context
	// and then reset to "empty".
	// When the current context is closed, its final local result is combined with the previous context's result
	// (restored from 'prefixes') to give the updated result.
	// This also means that once the last context has been closed,
	// the value contains the fully expanded set of strings (with the possible exception of a final suffix).
	var result []string

	// Start index of the substring that is currently being evaluated.
	offset := 0

	// TODO Implement escaping of '{', ',', and '}'.
	// TODO Make delimiters configurable.

	// Current context corresponding to the top element of the stack of contexts.
	var head *context
	for i, c := range expr {
		switch c {
		case '{':
			// Extract suffix.
			s := expr[offset:i]
			offset = i + 1

			// Append trailing suffix to all results.
			if s != "" {
				result = appendToAll(result, s)
			}

			// Push new expansion context.
			head = &context{
				offset:   i,
				parent:   head,
				prefixes: result,
				result:   nil,
			}
			result = nil
		case ',':
			if head == nil {
				// Comma is not valid outside a context.
				return nil, Error{Pos: i, Missing: 0}
			}

			// Extract suffix.
			s := expr[offset:i]
			offset = i + 1

			// Flush result with the suffix appended into the context's set of inner strings.
			if result != nil {
				for _, r := range result {
					head.result = append(head.result, r+s)
				}
				result = nil
			} else { // if head.result != nil || s != "" {
				// Not correct to say head.result = nil means []string{""}
				// because it may actually be empty and it may actually contain an empty string.
				head.result = append(head.result, s)
			}
		case '}':
			if head == nil {
				return nil, Error{Pos: i, Missing: '{'}
			}

			// Extract suffix.
			s := expr[offset:i]
			offset = i + 1

			// Flush result with the suffix appended to the context's set of inner strings.
			if result != nil {
				for _, r := range result {
					head.result = append(head.result, r+s)
				}
			} else if head.result != nil || s != "" {
				head.result = append(head.result, s)
			}

			// Close the context by prepending all its prefixes to all its inner strings.
			// TODO This might result in result=[]string{""}.
			result = combine(head.prefixes, head.result)

			// Pop current context from stack.
			head = head.parent
		}
	}

	// Append any trailing suffix to all results.
	s := expr[offset:]
	if s != "" {
		result = appendToAll(result, s)
	}
	// Handle any unclosed context.
	if head != nil {
		// TODO Consider enabling error recovery.
		return nil, Error{Pos: head.offset, Missing: '}'}
	}
	if result == nil {
		return []string{""}, nil
	}
	return result, nil
}

func appendToAll(s []string, suffix string) []string {
	if s == nil {
		return append(s, suffix)
	}
	if suffix != "" {
		for i := range s {
			s[i] += suffix
		}
	}
	return s
}

func combine(prefixes, s []string) []string {
	if prefixes == nil {
		return s
	}
	if s == nil {
		return prefixes
	}
	r := make([]string, 0, len(prefixes)*len(s))
	for _, p := range prefixes {
		for _, s := range s {
			r = append(r, p+s)
		}
	}
	return r
}
