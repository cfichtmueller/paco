package paco

import (
	"fmt"
	"strings"
)

// Parse is the main parsing function. Provide a parser and an input and receive the parsing result.
func Parse[T any](parser Parser[T], data string) (T, error) {
	initial := State{
		Data:   data,
		Offset: 0,
	}
	result, final, err := parser(initial)
	if err != nil {
		var zero T
		return zero, err
	}
	if final.Offset < len(final.Data) {
		var zero T
		return zero, ErrUnconsumedInput
	}
	return result, nil
}

// AppendKeeping runs parserT and parserU in order and returns a Tuple containing the values of both parsers.
func AppendKeeping[T, U any](parserT Parser[T], parserU Parser[U]) Parser[Tuple[T, U]] {
	return func(initial State) (Tuple[T, U], State, error) {
		t, next, err := parserT(initial)
		if err != nil {
			var zero Tuple[T, U]
			return zero, initial, err
		}
		u, final, err := parserU(next)
		if err != nil {
			var zero Tuple[T, U]
			return zero, initial, err
		}
		return Tuple[T, U]{A: t, B: u}, final, nil
	}
}

// AppendSkipping runs parserT and parserU in order, discards the result of parserU and returns the result of parserT
func AppendSkipping[T, U any](parserT Parser[T], parserU Parser[U]) Parser[T] {
	return func(initial State) (T, State, error) {
		t, next, err := parserT(initial)
		if err != nil {
			var zero T
			return zero, initial, err
		}
		_, final, err := parserU(next)
		if err != nil {
			var zero T
			return zero, initial, err
		}
		return t, final, nil
	}
}

// Between runs parsers p1, p2, p3 in order and returns the result of p2
func Between[T, U, A any](p1 Parser[T], p2 Parser[A], p3 Parser[U]) Parser[A] {
	return func(initial State) (A, State, error) {
		var zero A
		_, next, err := p1(initial)
		if err != nil {
			return zero, initial, err
		}

		result, next, err := p2(next)
		if err != nil {
			return zero, initial, err
		}

		_, next, err = p3(next)
		if err != nil {
			return zero, initial, err
		}
		return result, next, nil
	}
}

// ConsumeIf consumes a rune if the condition holds true. If not it returns ErrNoMatch
func ConsumeIf(condition func(rune) bool) Parser[Empty] {
	return func(initial State) (Empty, State, error) {
		r, next := initial.NextRune()
		if condition(r) {
			return empty, next, nil
		}
		return empty, initial, ErrNoMatch
	}
}

// ConsumeSome consumes runes for as long as the condition holds true. It expects to consume at least one rune.
func ConsumeSome(condition func(rune) bool) Parser[Empty] {
	s := StartSkipping(ConsumeIf(condition))
	return AppendSkipping(s, ConsumeWhile(condition))
}

// ConsumeWhile consumes runes for as long as the condition holds true. May consume no runes.
func ConsumeWhile(condition func(rune) bool) Parser[Empty] {
	return func(initial State) (Empty, State, error) {
		current := initial
		for current.HasRemaining() {
			r, next := current.NextRune()
			if !condition(r) {
				return empty, current, nil
			}
			current = next
		}
		return empty, current, nil
	}
}

// Exactly consumes the given token. If it cans, it returns ErrNoMatch
func Exactly(token string) Parser[Empty] {
	return func(initial State) (Empty, State, error) {
		if !strings.HasPrefix(initial.Remaining(), token) {
			return empty, initial, ErrNoMatch
		}
		next := initial.Consume(len(token))
		return empty, next, nil
	}
}

// Fail fails parsing with ErrNoMatch
func Fail[T any](initial State) (T, State, error) {
	var zero T
	return zero, initial, ErrNoMatch
}

// FlatMap works like map but allows the mapper to decide whether to succeed or fail the operation.
func FlatMap[T, U any](parser Parser[T], mapper func(T) Parser[U]) Parser[U] {
	return func(initial State) (U, State, error) {
		t, next, err := parser(initial)
		if err != nil {
			var zero U
			return zero, initial, err
		}
		uParser := mapper(t)
		return uParser(next)
	}
}

// GetString returns a string containing all characters consumed by the given parser
func GetString[T any](parser Parser[T]) Parser[string] {
	return func(initial State) (string, State, error) {
		start := initial.Offset
		_, next, err := parser(initial)
		if err != nil {
			return "", initial, err
		}
		end := next.Offset
		return initial.Data[start:end], next, nil
	}
}

// Infix parses infix operator notations. Returns a tuple containing the value of infix and another tuple with the
// values of left and right
func Infix[T1, U, T2 any](left Parser[T1], infix Parser[U], right Parser[T2]) Parser[Tuple[U, Tuple[T1, T2]]] {
	p1 := StartKeeping(left)
	p2 := AppendKeeping(p1, infix)
	p3 := AppendKeeping(p2, right)
	return Map(p3, func(t Tuple[Tuple[Tuple[Empty, T1], U], T2]) Tuple[U, Tuple[T1, T2]] {
		return Tuple[U, Tuple[T1, T2]]{
			A: t.A.B,
			B: Tuple[T1, T2]{
				A: t.A.A.B,
				B: t.B,
			},
		}
	})
}

// LeftAndRight parses left, sep, right and returns the values of left and right.
// Useful for infix operator parsing where the operator value isn't needed
func LeftAndRight[T1, U, T2 any](left Parser[T1], sep Parser[U], right Parser[T2]) Parser[Tuple[T1, T2]] {
	p1 := StartKeeping(left)
	p2 := AppendSkipping(p1, sep)
	p3 := AppendKeeping(p2, right)
	return Unpack2(p3)
}

// Map runs the given parser, then applies mapper to the result
func Map[T, U any](parser Parser[T], mapper func(T) U) Parser[U] {
	return func(initial State) (U, State, error) {
		t, next, err := parser(initial)
		if err != nil {
			var zero U
			return zero, initial, err
		}
		return mapper(t), next, nil
	}
}

// MapEmpty is a shortcut of Map for mapping results of empty parsers
func MapEmpty[T any](parser Parser[Empty], t T) Parser[T] {
	return Map(parser, func(e Empty) T { return t })
}

// MapT is a shortcut of map for mapping results of Tuple parsers
func MapT[T, A any](parser Parser[Tuple[Empty, T]], mapper func(T) A) Parser[A] {
	return func(initial State) (A, State, error) {
		tuple, next, err := parser(initial)
		if err != nil {
			var zero A
			return zero, initial, err
		}
		return mapper(tuple.B), next, nil
	}
}

// MapT2 is a shortcut of map for mapping results of Tuple[Tuple] parsers
func MapT2[T, U, A any](parser Parser[Tuple[Tuple[Empty, T], U]], mapper func(T, U) A) Parser[A] {
	return func(initial State) (A, State, error) {
		tuple, next, err := parser(initial)
		if err != nil {
			var zero A
			return zero, initial, err
		}
		return mapper(tuple.A.B, tuple.B), next, nil
	}
}

// MapT3 is a shortcut of map for mapping results of Tuple[Tuple[Tuple] parsers
func MapT3[T, U, V, A any](parser Parser[Tuple[Tuple[Tuple[Empty, T], U], V]], mapper func(T, U, V) A) Parser[A] {
	return func(initial State) (A, State, error) {
		tuple, next, err := parser(initial)
		if err != nil {
			var zero A
			return zero, initial, err
		}
		return mapper(tuple.A.A.B, tuple.A.B, tuple.B), next, nil
	}
}

// MapT4 is a shortcut of map for mapping results of Tuple[Tuple[Tuple[Tuple] parsers
func MapT4[T, U, V, W, A any](parser Parser[Tuple[Tuple[Tuple[Tuple[Empty, T], U], V], W]], mapper func(T, U, V, W) A) Parser[A] {
	return func(initial State) (A, State, error) {
		tuple, next, err := parser(initial)
		if err != nil {
			var zero A
			return zero, initial, err
		}
		return mapper(tuple.A.A.A.B, tuple.A.A.B, tuple.A.B, tuple.B), next, nil
	}
}

// OneOf runs all given parsers in order, returns the result of the first parser that doesn't return an error
func OneOf[T any](parsers ...Parser[T]) Parser[T] {
	return func(initial State) (T, State, error) {
		err := ErrNoMatch
		for _, p := range parsers {
			var result T
			var next State
			result, next, err = p(initial)
			if err == nil {
				return result, next, err
			}
		}
		var zero T
		return zero, initial, err
	}
}

// RepeatWhile repeatedly applies parser p while the predicate is satisfied
func RepeatWhile[T any](parser Parser[T], predicate func(T) bool) Parser[[]T] {
	return func(initial State) ([]T, State, error) {
		current := initial
		result := make([]T, 0)
		for current.HasRemaining() {
			r, next, err := parser(current)
			if err != nil {
				return nil, current, err
			}
			if !predicate(r) {
				break
			}
			current = next
			result = append(result, r)
		}
		return result, current, nil
	}
}

// WithLabel wraps the given parsers error messages with the given label
func WithLabel[T any](p Parser[T], label string) Parser[T] {
	return func(initial State) (T, State, error) {
		t, next, err := p(initial)
		if err != nil {
			return t, initial, fmt.Errorf("error parsing %s: %v", label, err)
		}
		return t, next, nil
	}
}

// SepBy parses zero or more p separated by sep
func SepBy[T, A any](p Parser[A], sep Parser[T]) Parser[[]A] {
	return func(initial State) ([]A, State, error) {
		current := initial
		result := make([]A, 0)
		for {
			if len(current.Remaining()) == 0 {
				return result, current, nil
			}
			val, afterVal, err := p(current)
			if err != nil {
				return nil, initial, err
			}
			result = append(result, val)
			_, afterSep, err := sep(afterVal)
			if err != nil {
				return result, afterVal, nil
			}
			current = afterSep
		}
	}
}

// SepBy1 parses one or more p separated by sep
func SepBy1[T, A any](p Parser[A], sep Parser[T]) Parser[[]A] {
	return func(initial State) ([]A, State, error) {
		current := initial
		result := make([]A, 0, 1)
		first := true
		for {
			val, next, err := p(current)
			if err != nil {
				return nil, initial, err
			}
			result = append(result, val)
			if !first && len(next.Remaining()) == 0 {
				return result, next, nil
			}
			_, next, err = sep(next)
			if err != nil {
				return nil, initial, err
			}
			first = false
			current = next
		}
	}
}

// StartKeeping returns a tuple with the result of the given parser
func StartKeeping[T any](parser Parser[T]) Parser[Tuple[Empty, T]] {
	return Map(parser, func(t T) Tuple[Empty, T] {
		return Tuple[Empty, T]{A: empty, B: t}
	})
}

// StartSkipping discards the result of the given parser
func StartSkipping[T any](parser Parser[T]) Parser[Empty] {
	return Map(parser, func(T) Empty { return empty })
}

// Succeed returns the given value
func Succeed[T any](value T) Parser[T] {
	return func(initial State) (T, State, error) {
		return value, initial, nil
	}
}

// Unpack unpacks the tuple result of the given parser.
func Unpack[T any](parser Parser[Tuple[Empty, T]]) Parser[T] {
	return func(initial State) (T, State, error) {
		tuple, next, err := parser(initial)
		if err != nil {
			var zero T
			return zero, initial, err
		}
		return tuple.B, next, nil
	}
}

// Unpack2 unpacks level 2 tuple nesting
func Unpack2[T, U any](parser Parser[Tuple[Tuple[Empty, T], U]]) Parser[Tuple[T, U]] {
	return func(initial State) (Tuple[T, U], State, error) {
		val, next, err := parser(initial)
		if err != nil {
			var zero Tuple[T, U]
			return zero, initial, err
		}
		return Tuple[T, U]{
			A: val.A.B,
			B: val.B,
		}, next, nil
	}
}
