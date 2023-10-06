package paco

import (
	"testing"
)

func TestState_Consume(t *testing.T) {
	s1 := State{Data: "abc", Offset: 0}
	s2 := s1.Consume(0)
	expectState(t, s2, "abc", 0)

	s2 = s1.Consume(2)
	expectState(t, s2, "c", 2)

	s3 := s2.Consume(1)
	expectState(t, s3, "", 3)

	if s3.HasRemaining() {
		t.Errorf("state has remaining, expected false")
	}
}

func expectState(t *testing.T, state State, remaining string, offset int) {
	if state.Remaining() != remaining {
		t.Errorf("expected remaining '%s', got '%s'", remaining, state.Remaining())
	}
	if state.Offset != offset {
		t.Errorf("expected offset %d, got %d", offset, state.Offset)
	}
}
