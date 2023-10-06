package paco

import (
	"fmt"
)

var ErrNoMatch = fmt.Errorf("no match")
var ErrUnconsumedInput = fmt.Errorf("unconsumed input")

type Empty struct{}

var empty = Empty{}

type Parser[T any] func(state State) (T, State, error)

// Tuple is a tuple of two values
type Tuple[T, U any] struct {
	A T
	B U
}
