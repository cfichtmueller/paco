package paco

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

func IsAsciiLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func IsDecimalDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func IsWhitespace(r rune) bool {
	return r == '\t' || r == ' '
}

func IsNotWhitespace(r rune) bool {
	return !IsWhitespace(r)
}

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
