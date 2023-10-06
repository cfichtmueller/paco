package paco

import "unicode/utf8"

type State struct {
	Data   string
	Offset int
}

func (s State) HasRemaining() bool {
	return len(s.Data) > s.Offset
}

func (s State) Remaining() string {
	return s.Data[s.Offset:]
}

func (s State) Consume(n int) State {
	s.Offset += n
	return s
}

func (s State) NextRune() (rune, State) {
	r, w := utf8.DecodeRuneInString(s.Remaining())
	return r, s.Consume(w)
}
