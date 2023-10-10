package examples

import (
	_ "embed"
	"github.com/cfichtmueller/paco"
	"strings"
	"testing"
)

type TemplateContext map[string]string

type Template struct {
	Parts []TemplatePart
}

func (t Template) Render(c TemplateContext) string {
	result := strings.Builder{}
	for _, p := range t.Parts {
		result.WriteString(p(c))
	}
	return result.String()
}

type TemplatePart func(c TemplateContext) string

//go:embed template1.html
var template1 string

//go:embed template2.html
var template2 string

func Test_template_1(t *testing.T) {
	template, ok := expectParserParses(t, createTemplateParser(), template1)
	if !ok {
		return
	}
	expectParts(t, 1, template)
}

func Test_template_2(t *testing.T) {
	template, ok := expectParserParses(t, createTemplateParser(), template2)
	if !ok {
		return
	}
	expectParts(t, 3, template)
}

func Test_template_parser(t *testing.T) {
	template, ok := expectParserParses(t, createTemplateParser(), "my {{ $what}} is {{ $value }}")
	if !ok {
		return
	}
	expectParts(t, 4, template)

	expected := "my name is Smith"
	actual := template.Render(map[string]string{
		"what":  "name",
		"value": "Smith",
	})

	if actual != expected {
		t.Errorf("Expected '%s', got '%s'", expected, actual)
	}

	expected = "my name is "
	actual = template.Render(map[string]string{
		"what": "name",
	})

	if actual != expected {
		t.Errorf("Expected '%s', got '%s'", expected, actual)
	}
}

func createTemplateParser() paco.Parser[Template] {
	templateParser := paco.Map(
		paco.RepeatWhile(
			paco.OneOf(
				createLiteralParser(),
				createPlaceholderParser(),
			),
			func(part TemplatePart) bool { return true },
		),
		func(p []TemplatePart) Template {
			return Template{Parts: p}
		},
	)
	return templateParser
}

func Test_template_placeholder_parser(t *testing.T) {
	placeholder, err := paco.Parse(createPlaceholderParser(), "{{ $value }}")
	if err != nil {
		t.Errorf("placeholderParser didn't parse placeholder: %v", err)
	}
	actual := placeholder(map[string]string{"value": "23"})
	expected := "23"
	if actual != expected {
		t.Errorf("expected '%s', got '%s'", expected, actual)
	}
}

func createPlaceholderParser() paco.Parser[TemplatePart] {
	consumeWhitespace := paco.ConsumeWhile(paco.IsWhitespace)
	beginPlaceholderParser := paco.AppendSkipping(
		paco.StartSkipping(paco.Exactly("{{")),
		consumeWhitespace,
	)
	endPlaceholderParser := paco.AppendSkipping(
		paco.StartSkipping(consumeWhitespace),
		paco.Exactly("}}"),
	)
	placeholderParser := paco.Between(
		beginPlaceholderParser,
		paco.OneOf(createVariableParser()),
		endPlaceholderParser,
	)
	return placeholderParser
}

func Test_template_literal_parser(t *testing.T) {
	parser := createLiteralParser()
	_, err := paco.Parse(parser, "hello world")
	if err != nil {
		t.Errorf("literalParser didn't parse: %v", err)
	}

	_, next, err := parser(paco.State{
		Data:   "{{foo}}",
		Offset: 0,
	})
	if err == nil {
		t.Errorf("parser parsed erroneously")
	}
	if next.Offset != 0 {
		t.Errorf("Expected offset 0, got %d", next.Offset)
	}
}

func createLiteralParser() paco.Parser[TemplatePart] {
	literalParser := paco.Map(
		paco.GetString(paco.ConsumeSome(paco.IsNoneOf('{'))),
		func(literal string) TemplatePart { return func(c TemplateContext) string { return literal } },
	)
	return literalParser
}

func Test_template_variable_parser(t *testing.T) {
	parser := createVariableParser()

	part, err := paco.Parse(parser, "$name")
	if err != nil {
		t.Errorf("Parser didn't parse: %v", err)
	}
	actual := part(map[string]string{"name": "Jon"})
	expected := "Jon"
	if actual != expected {
		t.Errorf("Expected '%s', got '%s'", expected, actual)
	}
}

func createVariableParser() paco.Parser[TemplatePart] {
	variableParser := paco.Map(
		paco.Unpack(
			paco.AppendKeeping(
				paco.StartSkipping(paco.Exactly("$")),
				paco.GetString(paco.ConsumeWhile(paco.IsNoneOf(' ', '}'))),
			),
		),
		func(name string) TemplatePart {
			return func(c TemplateContext) string {
				val, ok := c[name]
				if !ok {
					return ""
				}
				return val
			}
		},
	)
	return variableParser
}

func expectParts(t *testing.T, expected int, template Template) {
	actual := len(template.Parts)
	if actual != expected {
		t.Errorf("Expected %d parts, got %d", expected, actual)
	}
}

func expectParserParses[T any](t *testing.T, parser paco.Parser[T], input string) (T, bool) {
	result, err := paco.Parse(parser, input)
	if err != nil {
		t.Errorf("parser didn't parse template: %v", err)
		var zero T
		return zero, false
	}
	return result, true
}
