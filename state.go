package paco

import "unicode/utf8"

type State struct {
	Data   string
	Offset int
}

// HasRemaining returns true if the state has data left
func (s State) HasRemaining() bool {
	return len(s.Data) > s.Offset
}

// Remaining returns the remaining data
func (s State) Remaining() string {
	return s.Data[s.Offset:]
}

// Consume returns a new state with the offset advanced
func (s State) Consume(n int) State {
	s.Offset += n
	return s
}

func (s State) NextRune() (rune, State) {
	r, w := utf8.DecodeRuneInString(s.Remaining())
	return r, s.Consume(w)
}
