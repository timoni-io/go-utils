package types

import "errors"

var (
	ErrNilSet      = errors.New("set is nil")
	ErrNilMap      = errors.New("map is nil")
	ErrReadOnlyMap = errors.New("map is readonly")
)

type Iterator[K comparable, V any] <-chan Item[K, V]


type Item[K comparable, V any] struct {
	Key   K
	Value V
}

// encoding/json Marshal function type
type MarshalFunc func(v any) ([]byte, error)

// encoding/json Unmarshal function type
type UnmarshalFunc func(data []byte, v any) error
