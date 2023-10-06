package examples

import (
	"github.com/cfichtmueller/paco"
	"testing"
)

func Test_ParseQuotedString(t *testing.T) {
	quoteParser := paco.Exactly("\"")
	stringParser := paco.GetString(paco.ConsumeWhile(paco.MatchAny(paco.IsAsciiLetter, paco.IsDecimalDigit, paco.IsWhitespace)))
	parser := paco.Between(quoteParser, stringParser, quoteParser)

	val, err := paco.Parse(parser, "\"hello world 123\"")
	if err != nil {
		t.Errorf("parser didn't parse")
	}
	if val != "hello world 123" {
		t.Errorf("parser parsed '%s', expected 'hello world 123'", val)
	}

	val, err = paco.Parse(parser, "hello world")
	if err == nil {
		t.Errorf("parser parsed invalid input")
	}
}
