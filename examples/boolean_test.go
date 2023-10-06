package examples

import (
	"github.com/cfichtmueller/paco"
	"testing"
)

func Test_Boolean(t *testing.T) {
	falseParser := paco.MapEmpty(paco.Exactly("false"), false)
	trueParser := paco.MapEmpty(paco.Exactly("true"), true)
	boolParser := paco.OneOf(falseParser, trueParser)

	v, err := paco.Parse(boolParser, "false")
	if err != nil {
		t.Errorf("parser didn't parse 'false'")
	}
	if v {
		t.Errorf("expected false, got true")
	}

	v, err = paco.Parse(boolParser, "true")
	if err != nil {
		t.Errorf("parser didn't parse 'true'")
	}
	if !v {
		t.Errorf("expected true, got false")
	}

	_, err = paco.Parse(boolParser, "blue")
	if err == nil {
		t.Errorf("parser parsed erroneusly")
	}
}
