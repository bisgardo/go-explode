package explode

// ReportErrorFunc is the type of a function which consumes errors encountered during Explode.
// TODO Make function return whether the algorithm should recover or not.
type ReportErrorFunc func(pos int, msg string)

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
// Errors are reported to the provided function.
func Explode(expr string, reportErr ReportErrorFunc) []string {
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
				if result == nil {
					result = append(result, s)
				} else {
					for i := range result {
						result[i] += s
					}
				}
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
				// Not inside context; comma is just a character like any other.
				reportErr(i, "invalid separator")
				// Otherwise just ignore.
				continue
			}

			// Extract suffix.
			s := expr[offset:i]
			offset = i + 1

			// Flush result with the suffix appended into the context's set of inner strings.
			if result == nil {
				head.result = append(head.result, s)
			} else {
				for _, r := range result {
					head.result = append(head.result, r+s)
				}
				result = nil
			}
		case '}':
			if head == nil {
				// Not inside an expansion context: Brace is unmatched.
				reportErr(i, "no matching '{'")
				// Recover by creating a dummy context.
				// Could alternatively restart the algorithm with a '{'
				// inserted in the beginning.
				head = &context{prefixes: nil}
			}

			// Extract suffix.
			s := expr[offset:i]
			offset = i + 1

			// Flush result with the suffix appended to the context's set of inner strings.
			if result == nil {
				head.result = append(head.result, s)
			} else {
				for _, r := range result {
					head.result = append(head.result, r+s)
				}
			}

			// Close the context by prepending all its prefixes to all its inner strings.
			if head.prefixes == nil {
				result = head.result
			} else {
				result = make([]string, 0, len(head.prefixes)*len(head.result))
				for _, p := range head.prefixes {
					for _, s := range head.result {
						result = append(result, p+s)
					}
				}
			}

			// Pop current context from stack.
			head = head.parent
		}
	}

	// Append any trailing suffix to all results.
	s := expr[offset:]
	if result == nil {
		result = append(result, s)
	} else if s != "" {
		for i := range result {
			result[i] += s
		}
	}
	// Handle unclosed contexts.
	for head != nil {
		reportErr(head.offset, "no matching '}'")

		// Fill current result into context.
		for _, r := range result {
			head.result = append(head.result, r)
		}

		// Close the context by prepending all its prefixes to all its inner strings.
		if head.prefixes == nil {
			result = head.result
		} else {
			result = make([]string, 0, len(head.prefixes)*len(head.result))
			for _, p := range head.prefixes {
				for _, s := range head.result {
					result = append(result, p+s)
				}
			}
		}
		// Pop current context from stack.
		head = head.parent
	}
	return result
}
