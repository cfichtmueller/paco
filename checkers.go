package paco

// MatchAny returns a predicate that tests the given predicates in order. Returns true if any predicate matches.
func MatchAny(c ...func(rune) bool) func(rune) bool {
	return func(r rune) bool {
		for _, current := range c {
			if current(r) {
				return true
			}
		}
		return false
	}
}

// IsAsciiLetter returns true if the given rune is in a-z or A-/
func IsAsciiLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

// IsDecimalDigit returns true if the given rune is in 0-9
func IsDecimalDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// IsWhitespace returns true if the given rune is tab or space
func IsWhitespace(r rune) bool {
	return r == '\t' || r == ' '
}

// IsNotWhitespace negates IsWhitespace
func IsNotWhitespace(r rune) bool {
	return !IsWhitespace(r)
}

// IsAnyOf returns true if the given rune is any of the allowed ones.
func IsAnyOf(allowed ...rune) func(rune) bool {
	return func(r rune) bool {
		for _, c := range allowed {
			if r == c {
				return true
			}
		}
		return false
	}
}

// IsNoneOf returns true if the given rune is not any of the forbidden ones.
func IsNoneOf(forbidden ...rune) func(rune) bool {
	return func(r rune) bool {
		for _, c := range forbidden {
			if r == c {
				return false
			}
		}
		return true
	}
}
