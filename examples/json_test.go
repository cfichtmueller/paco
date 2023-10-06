package examples

import (
	_ "embed"
	"github.com/cfichtmueller/paco"
	"testing"
)

type JsonValue interface{}

func JsonValueFromString(s string) JsonValue {
	return s
}

func JsonValueFromSlice(v []JsonValue) JsonValue {
	return v
}

type JsonEntry struct {
	Key   string
	Value JsonValue
}

func NewJsonEntry(key string, value JsonValue) JsonEntry {
	return JsonEntry{
		Key:   key,
		Value: value,
	}
}

//go:embed json1.json
var object1Json string

//go:embed json2.json
var object2Json string

//go:embed json3.json
var object3Json string

//go:embed json4.json
var object4Json string

func Test_JSON(t *testing.T) {
	consumeWhitespace := paco.ConsumeWhile(paco.IsWhitespace)
	consumeWhitespaceOrNewline := paco.ConsumeWhile(paco.MatchAny(paco.IsWhitespace, paco.IsAnyOf('\n', '\r')))
	identifierParser := paco.Map(
		paco.AppendKeeping(
			paco.GetString(paco.ConsumeSome(paco.IsAsciiLetter)),
			paco.GetString(paco.ConsumeWhile(paco.MatchAny(paco.IsAsciiLetter, paco.IsDecimalDigit))),
		),
		func(s paco.Tuple[string, string]) string {
			return s.A + s.B
		})
	keyParser := paco.Between(paco.Exactly("\""), identifierParser, paco.Exactly("\""))

	key1, err := paco.Parse(keyParser, "\"name\"")
	if err != nil {
		t.Errorf("key parser didn't parse identifier")
	}
	if key1 != "name" {
		t.Errorf("expected key 'name', got %s", key1)
	}

	_, err = paco.Parse(keyParser, "name")
	if err == nil {
		t.Errorf("key parser parsed invalid key")
	}

	numberParser := paco.Map(paco.GetString(paco.ConsumeSome(paco.IsDecimalDigit)), JsonValueFromString)
	stringParser := paco.Map(paco.Between(paco.Exactly("\""), paco.GetString(paco.ConsumeWhile(paco.IsNoneOf('"', '\n', '\r'))), paco.Exactly("\"")), JsonValueFromString)
	arrayParser := paco.Map(paco.Between(
		paco.Exactly("["),
		paco.SepBy(
			paco.OneOf(numberParser, stringParser),
			paco.Between(consumeWhitespaceOrNewline, paco.Exactly(","), consumeWhitespaceOrNewline),
		),
		paco.Exactly("]"),
	), JsonValueFromSlice)
	trueParser := paco.Map(paco.Exactly("true"), func(empty paco.Empty) JsonValue { return true })
	falseParser := paco.Map(paco.Exactly("false"), func(empty paco.Empty) JsonValue { return false })
	nullParser := paco.Map(paco.Exactly("null"), func(empty paco.Empty) JsonValue { return nil })
	valueParser := paco.OneOf(numberParser, stringParser, arrayParser, trueParser, falseParser, nullParser)

	v1, err := paco.Parse(valueParser, "0815")
	if err != nil {
		t.Errorf("value parser didn't parse number value")
	}
	if v1 != "0815" {
		t.Errorf("expected 0815, got %s", v1)
	}

	v2, err := paco.Parse(valueParser, "\"hello world\"")
	if err != nil {
		t.Errorf("value parser didn't parse string value")
	}
	if v2 != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", v2)
	}

	oep1 := paco.StartKeeping(keyParser)
	oep2 := paco.AppendSkipping(oep1, paco.Between(consumeWhitespace, paco.Exactly(":"), consumeWhitespace))
	oep3 := paco.AppendKeeping(oep2, valueParser)
	objectEntryParser := paco.WithLabel(paco.MapT2(oep3, NewJsonEntry), "object entry")

	e, err := paco.Parse(objectEntryParser, "\"name\":\"jon\"")
	if err != nil {
		t.Errorf("objectEntryParser didn't parse string entry")
	}
	if e.Key != "name" {
		t.Errorf("expected key 'name', got '%s'", e.Key)
	}
	if e.Value != "jon" {
		t.Errorf("expected value 'jon', got '%s'", e.Value)
	}

	e, err = paco.Parse(objectEntryParser, "\"value\" : 128")
	if err != nil {
		t.Errorf("objectEntryParser didn't parse number entry")
	}
	if e.Value != "128" {
		t.Errorf("expected value 128, got '%s'", e.Value)
	}

	startObjectParser := paco.WithLabel(paco.AppendSkipping(
		paco.StartSkipping(paco.Exactly("{")),
		consumeWhitespaceOrNewline,
	), "start object")
	endObjectParser := paco.WithLabel(paco.AppendSkipping(
		paco.StartSkipping(consumeWhitespaceOrNewline),
		paco.Exactly("}"),
	), "end object")

	_, err = paco.Parse(endObjectParser, "\r\n}")
	if err != nil {
		t.Errorf("endObjectParser didn't parse end object with preceding newline: %v", err)
	}

	entriesParser := paco.WithLabel(paco.SepBy(
		objectEntryParser,
		paco.WithLabel(paco.Between(consumeWhitespaceOrNewline, paco.Exactly(","), consumeWhitespaceOrNewline), "entry separator"),
	), "entries")

	_, err = paco.Parse(entriesParser, "\"name\" : \"jon\"")
	if err != nil {
		t.Errorf("entries parser didn't parse single entry")
	}

	_, err = paco.Parse(entriesParser, "\"name\":\"jon\",\"age\":24")
	if err != nil {
		t.Errorf("entries parser didn't parse multiple entries: %v", err)
	}

	_, err = paco.Parse(entriesParser, "\"name\":\"jon\",\r\n\"age\":24")
	if err != nil {
		t.Errorf("entries parser didn't parse multiple entries (containing newline): %v", err)
	}

	objectParser := paco.Between(
		startObjectParser,
		entriesParser,
		endObjectParser,
	)

	entries, err := paco.Parse(objectParser, "{ \"name\": \"jon\"}")
	if err != nil {
		t.Errorf("object parser didn't parse object: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}

	entries, err = paco.Parse(objectParser, "{ \"name\" : \"jon\" ,\n \"age\": 24}")
	if err != nil {
		t.Errorf("object parser didn't parse multi-key object: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}

	mustParseObject := func(name string, input string, expectedEntryCount int) {
		entries, err = paco.Parse(objectParser, input)
		if err != nil {
			t.Errorf("object parser didn't parse %s: %v", name, err)
		}
		if len(entries) != expectedEntryCount {
			t.Errorf("got %d entries, expected %d", len(entries), expectedEntryCount)
		}
	}

	mustParseObject("object1.json", object1Json, 2)
	mustParseObject("object2.json", object2Json, 1)
	mustParseObject("object3.json", object3Json, 3)
	mustParseObject("object4.json", object4Json, 6)
}
