package paco

import (
	"strings"
	"testing"
)

func TestConsumeWhile(t *testing.T) {
	consumer := ConsumeWhile(IsAnyOf('a', 'b', 'c'))

	mustParse := func(input, expected string) {
		_, s, err := consumer(State{Data: input, Offset: 0})
		if err != nil {
			t.Errorf("parser didn't parse")
		}
		if s.Remaining() != expected {
			t.Errorf("expected '%s', got '%s'", expected, s.Remaining())
		}
	}

	mustParse("cba", "")
	mustParse("xxx", "xxx")
	mustParse("abcd", "d")
}

func TestExactly(t *testing.T) {
	initial := State{
		Data:   "beef",
		Offset: 0,
	}
	beefMatcher := Exactly("beef")

	_, next, err := beefMatcher(initial)
	if err != nil {
		t.Errorf("beef wasn't matched")
	}
	if next.Remaining() != "" {
		t.Errorf("parser didn't consume")
	}

	cheeseMatcher := Exactly("cheese")

	_, _, err = cheeseMatcher(initial)
	if err == nil {
		t.Errorf("cheese matched beef")
	}

	bitMatcher := Exactly("bit")
	_, next, err = bitMatcher(initial)
	if err == nil {
		t.Errorf("bit matched beef")
	}
	if next.Remaining() != "beef" {
		t.Errorf("parser consumed although it failed")
	}
}

func TestInfix(t *testing.T) {
	digitParser := GetString(ConsumeSome(IsDecimalDigit))
	operatorParser := GetString(OneOf(Exactly("+"), Exactly("-"), Exactly("*"), Exactly("/")))
	parser := Infix(digitParser, operatorParser, digitParser)

	mustParse := func(input, left, operator, right string) {
		r, err := Parse(parser, input)
		if err != nil {
			t.Errorf("parser didn't parse '%s': %v", input, err)
		}
		if r.B.A != left {
			t.Errorf("Expected left '%s', got '%s'", left, r.B.A)
		}
		if r.A != operator {
			t.Errorf("Expected operator '%s', got '%s'", operator, r.A)
		}
		if r.B.B != right {
			t.Errorf("Expected right '%s', got '%s'", right, r.B.B)
		}
	}

	mustNotParse := func(input string) {
		_, err := Parse(parser, input)
		if err == nil {
			t.Errorf("Parser parsed '%s', expected error", "a+3")
		}
	}

	mustParse("45+3", "45", "+", "3")
	mustParse("8/7", "8", "/", "7")

	mustNotParse("a+3")
	mustNotParse("3x5")
	mustNotParse("3+n")
}

func TestRepeatWhile_read_all(t *testing.T) {
	parser := RepeatWhile(
		Unpack(
			AppendSkipping(
				AppendKeeping(
					StartSkipping(ConsumeWhile(IsWhitespace)),
					GetString(ConsumeWhile(IsNotWhitespace)),
				),
				ConsumeWhile(IsWhitespace),
			),
		),
		func(s string) bool { return true },
	)

	tokens, err := Parse(parser, "hello my dear parser combinators")
	if err != nil {
		t.Errorf("parser didn't parse")
	}
	expected := []string{"hello", "my", "dear", "parser", "combinators"}
	if len(expected) != 5 {
		t.Errorf("Expected %v, got %v", expected, tokens)
	}
}

func TestRepeatWhile_read_partially(t *testing.T) {
	parser := RepeatWhile(
		GetString(ConsumeWhile(IsNotWhitespace)),
		func(s string) bool { return len(s) > 0 },
	)
	tokens, next, err := parser(State{
		Data:   "hello world",
		Offset: 0,
	})
	if err != nil {
		t.Errorf("parser didn't parse")
	}
	expectedOffset := 5
	if next.Offset != expectedOffset {
		t.Errorf("Expected offset %d, got %d", expectedOffset, next.Offset)
	}
	if len(tokens) != 1 {
		t.Errorf("Expected [hello], got %v", tokens)
	}
}

func Test_SepBy(t *testing.T) {
	sep := AppendSkipping(StartSkipping(Exactly(",")), ConsumeWhile(IsWhitespace))
	val := Between(Exactly("'"), ConsumeWhile(IsAsciiLetter), Exactly("'"))
	parser := SepBy(val, sep)

	list, err := Parse(parser, "'a', 'b','c'")
	if err != nil {
		t.Errorf("parser didn't parse")
	}
	if len(list) != 3 {
		t.Errorf("parser parsed %d elements, expected 3", len(list))
	}

	list, err = Parse(parser, "")
	if err != nil {
		t.Errorf("parser didn't parse empty input")
	}
	if len(list) != 0 {
		t.Errorf("parser parsed %d elements, expected 0", len(list))
	}
}

func Test_SepBy1(t *testing.T) {
	sep := AppendSkipping(StartSkipping(Exactly(",")), ConsumeWhile(IsWhitespace))
	val := Between(Exactly("'"), ConsumeWhile(IsAsciiLetter), Exactly("'"))
	parser := SepBy1(val, sep)

	list, err := Parse(parser, "'a', 'b','c'")
	if err != nil {
		t.Errorf("parser didn't parse")
	}
	if len(list) != 3 {
		t.Errorf("parser parsed %d elements, expected 3", len(list))
	}

	list, err = Parse(parser, "")
	if err == nil {
		t.Errorf("parser didn't parsed empty input")
	}
}

func Test_WithLabel(t *testing.T) {
	digits := GetString(ConsumeSome(IsDecimalDigit))
	parser := WithLabel(digits, "digits")

	val, err := Parse(parser, "123")
	if err != nil {
		t.Errorf("parser didn't parse: %v", err)
	}
	if val != "123" {
		t.Errorf("parser didn't return value of nested parser")
	}

	_, err = Parse(parser, "hello")
	if err == nil {
		t.Errorf("parser parsed invalid input")
	}
	if !strings.HasPrefix(err.Error(), "error parsing digit") {
		t.Errorf("parser didn't return label: %v", err)
	}
}
