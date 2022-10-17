package slice

import (
	"encoding/json"
	"github.com/timoni-io/go-utils"
	"github.com/timoni-io/go-utils/types"

	"github.com/fxamacker/cbor/v2"
)

type Slice[T any] struct {
	lock     *utils.Lock
	data     []T
	capacity int
}

func NewSlice[T any](capacity int) *Slice[T] {
	return &Slice[T]{
		data:     make([]T, 0, capacity),
		capacity: capacity,
	}
}

func NewSafeSlice[T any](capacity int) *Slice[T] {
	return &Slice[T]{
		lock:     &utils.Lock{},
		data:     make([]T, 0, capacity),
		capacity: capacity,
	}
}

func (s *Slice[T]) Add(x ...T) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.data = append(s.data, x...)
}

func (s *Slice[T]) GetAll() []T {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.data
}

func (s *Slice[T]) Len() int {
	s.lock.Lock()
	defer s.lock.Unlock()
	return len(s.data)
}

func (s *Slice[T]) Clear() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.data = make([]T, 0, s.capacity)
}

func (s *Slice[T]) Get(idx int) *T {
	s.lock.Lock()
	defer s.lock.Unlock()

	if idx < 0 || idx >= len(s.data) {
		return nil
	}

	return &s.data[idx]
}

func (s *Slice[T]) Commit(fn func(data *[]T, capacity int)) {
	s.lock.Lock()
	defer s.lock.Unlock()
	fn(&s.data, s.capacity)
}

func (s *Slice[T]) Take() []T {
	s.lock.Lock()
	defer s.lock.Unlock()
	v := s.data
	s.data = make([]T, 0, s.capacity)
	return v
}

func (s *Slice[T]) marshal(m types.MarshalFunc) ([]byte, error) {
	if s == nil {
		return nil, types.ErrNilSet
	}

	s.lock.RLock()
	defer s.lock.RUnlock()

	return m(s.data)
}

func (s *Slice[T]) unmarshal(um types.UnmarshalFunc, data []byte) error {
	if s == nil {
		return types.ErrNilSet
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	return um(data, &s.data)
}

func (s *Slice[T]) MarshalJSON() ([]byte, error) {
	return s.marshal(json.Marshal)
}

func (s *Slice[T]) UnmarshalJSON(data []byte) error {
	return s.unmarshal(json.Unmarshal, data)
}

func (s *Slice[T]) MarshalCBOR() ([]byte, error) {
	return s.marshal(cbor.Marshal)
}

func (s *Slice[T]) UnmarshalCBOR(data []byte) error {
	return s.unmarshal(cbor.Unmarshal, data)
}
