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
	type context struct {
		// Pointer to the context inside of which the context is nested.
		parent *context
		// Index of the character which opened the context.
		offset int
		// Immutable set of prefixes that all substrings in the context will be appended to.
		prefixes []string
		// Mutable set of strings expanded locally in the context.
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
	// TODO Consider letting nil represent the empty result ([""]).
	result := []string{""}

	// Start index of the substring that is currently being evaluated.
	offset := 0

	// TODO Make delimiters configurable.

	// TODO Substrings need to be escaped. Options:
	//      - Extract runes eagerly into a buffer (probably the best solution, but would like to keep slicing when possible).
	//      - Escape on a separate pass if needed (add extra var to track this).
	//      - Store indices of runes to skip on extract (sounds bad).
	inEscape := false

	// Current context corresponding to the top element of the stack of contexts.
	var head *context
	for i, c := range expr {
		switch c {
		case '\\':
			if inEscape {
				inEscape = false
				break
			}
			inEscape = true
		case '{':
			if inEscape {
				inEscape = false
				break
			}
			// Extract suffix.
			s := expr[offset:i] // TODO Needs to unescape.
			offset = i + 1

			// Append trailing suffix to all results.
			if s != "" {
				for i := range result {
					result[i] += s
				}
			}

			// Push new expansion context.
			head = &context{
				offset:   i,
				parent:   head,
				prefixes: result,
				result:   nil,
			}
			result = []string{""}
		case ',':
			if inEscape {
				inEscape = false
				break
			}
			if head == nil {
				// Comma is not valid outside a context.
				return nil, Error{Pos: i, Missing: 0}
			}

			// Extract suffix.
			s := expr[offset:i] // TODO Needs to unescape.
			offset = i + 1

			// Flush result with the suffix appended into the context's set of inner strings.
			for _, r := range result {
				head.result = append(head.result, r+s)
			}
			result = []string{""}
		case '}':
			if inEscape {
				inEscape = false
				break
			}
			if head == nil {
				return nil, Error{Pos: i, Missing: '{'}
			}

			// Extract suffix.
			s := expr[offset:i] // TODO Needs to unescape.
			offset = i + 1

			// Flush result with the suffix appended to the context's set of inner strings.
			for _, r := range result {
				head.result = append(head.result, r+s)
			}

			// Close the context by prepending all its prefixes to all its inner strings.
			result = make([]string, 0, len(head.prefixes)*len(head.result))
			for _, p := range head.prefixes {
				for _, s := range head.result {
					result = append(result, p+s)
				}
			}

			// Pop current context from stack.
			head = head.parent
		default:
			if inEscape {
				return nil, fmt.Errorf("invalid escape sequence '\\%c'", c)
			}
		}
	}
	if inEscape {
		return nil, fmt.Errorf("invalid trailing backslash")
	}

	// Append any trailing suffix to all results.
	s := expr[offset:] // TODO Needs to unescape.
	if s != "" {
		for i := range result {
			result[i] += s
		}
	}
	// Handle any unclosed context.
	if head != nil {
		// TODO Consider enabling error recovery.
		return nil, Error{Pos: head.offset, Missing: '}'}
	}
	return result, nil
}
