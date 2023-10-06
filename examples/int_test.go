package examples

import (
	"github.com/cfichtmueller/paco"
	"strconv"
	"testing"
)

func Test_Integers(t *testing.T) {
	intParser := paco.FlatMap(
		paco.GetString(paco.ConsumeSome(paco.IsDecimalDigit)),
		func(digits string) paco.Parser[int] {
			if len(digits) > 1 && digits[0] == '0' {
				return paco.Fail[int]
			}
			v, err := strconv.Atoi(digits)
			if err != nil {
				return paco.Fail[int]
			}
			return paco.Succeed(v)
		},
	)

	mustParse := func(input string, expected int) {
		d, err := paco.Parse(intParser, input)
		if err != nil {
			t.Errorf("parser didn't parse")
		}
		if d != expected {
			t.Errorf("expected %d, got %d", expected, d)
		}
	}

	mustNotParse := func(input string) {
		_, err := paco.Parse(intParser, input)
		if err == nil {
			t.Errorf("parser parsed '%s' erroneously", input)
		}
	}

	mustParse("0", 0)
	mustParse("1", 1)
	mustParse("10", 10)
	mustParse("123", 123)

	mustNotParse("00")
	mustNotParse("abc")
	mustNotParse("12c")
}
