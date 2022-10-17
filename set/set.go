package set

import (
	"encoding/json"
	"fmt"

	"github.com/timoni-io/go-utils"
	"github.com/timoni-io/go-utils/types"

	"github.com/fxamacker/cbor/v2"
)

// --- Set implemented using map ---

type void struct{}

type Set[T comparable] struct {
	lock *utils.Lock
	data map[T]void
}

func New[T comparable](data ...T) *Set[T] {
	set := &Set[T]{}
	set.Add(data...)
	return set
}

func NewSafe[T comparable](data ...T) *Set[T] {
	set := &Set[T]{lock: &utils.Lock{}}
	set.Add(data...)
	return set
}

func (set *Set[T]) init() error {
	if set == nil {
		return types.ErrNilSet
	}
	if set.data == nil {
		set.data = map[T]void{}
	}

	return nil
}

func (set *Set[T]) Add(values ...T) {
	if set.init() != nil {
		return
	}

	set.lock.Lock()
	defer set.lock.Unlock()

	for _, value := range values {
		set.data[value] = void{}
	}
}

func (set *Set[T]) Delete(values ...T) {
	if set == nil {
		return
	}

	set.lock.Lock()
	defer set.lock.Unlock()

	for _, v := range values {
		delete(set.data, v)
	}
}

func (set *Set[T]) Remove(values ...T) {
	if set.init() != nil {
		return
	}

	set.lock.Lock()
	defer set.lock.Unlock()

	for _, value := range values {
		delete(set.data, value)
	}
}

func (set *Set[T]) Contains(value T) bool {
	if set == nil {
		return false
	}

	set.lock.RLock()
	defer set.lock.RUnlock()

	_, exists := set.data[value]
	return exists
}

func (set *Set[T]) Iter() <-chan T {
	out := make(chan T, len(set.data))

	set.lock.RLock()
	go func() {
		defer set.lock.RUnlock()
		for value := range set.data {
			out <- value
		}
	}()

	return out
}

func (set *Set[T]) List() (list []T) {
	if set == nil {
		return
	}

	set.lock.RLock()
	defer set.lock.RUnlock()

	i := 0
	list = make([]T, len(set.data))
	for value := range set.data {
		list[i] = value
		i++
	}
	return
}

func (set *Set[T]) Length() int {
	if set == nil {
		return 0
	}

	set.lock.RLock()
	defer set.lock.RUnlock()

	return len(set.data)
}

func (set *Set[T]) String() string {
	if set == nil {
		return "[]"
	}

	return fmt.Sprint(set.List())
}

func (set *Set[T]) marshal(m types.MarshalFunc) ([]byte, error) {
	if set == nil {
		return nil, types.ErrNilSet
	}

	set.lock.RLock()
	defer set.lock.RUnlock()

	return m(set.List())
}

func (set *Set[T]) unmarshal(um types.UnmarshalFunc, data []byte) error {
	if set == nil {
		return types.ErrNilSet
	}

	set.lock.Lock()
	defer set.lock.Unlock()

	var values []T
	err := um(data, &values)
	if err != nil {
		return err
	}

	set.data = nil
	set.Add(values...)

	return nil
}

func (set *Set[T]) MarshalJSON() ([]byte, error) {
	return set.marshal(json.Marshal)
}

func (set *Set[T]) UnmarshalJSON(data []byte) error {
	return set.unmarshal(json.Unmarshal, data)
}

func (set *Set[T]) MarshalCBOR() ([]byte, error) {
	return set.marshal(cbor.Marshal)
}

func (set *Set[T]) UnmarshalCBOR(data []byte) error {
	return set.unmarshal(cbor.Unmarshal, data)
}
