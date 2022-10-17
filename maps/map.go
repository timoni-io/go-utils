package maps

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/timoni-io/go-utils"
	"github.com/timoni-io/go-utils/channel"
	"github.com/timoni-io/go-utils/types"

	"github.com/fxamacker/cbor/v2"
)

type Map[K comparable, V any] struct {
	lock     *utils.Lock
	data     map[K]V
	readonly bool
	*channel.Hub[types.WatchMsg[K, V]]
}

func New[K comparable, V any](data map[K]V) *Map[K, V] {
	return &Map[K, V]{data: data}
}

func (m *Map[K, V]) init() error {
	if m == nil {
		return types.ErrNilMap
	}
	if m.data == nil {
		m.data = map[K]V{}
	}

	return nil
}

// retrun map with event chan
func (m *Map[K, V]) Eventfull(ctx context.Context, buf int) *Map[K, V] {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.Hub = channel.NewHub[types.WatchMsg[K, V]](ctx, buf)
	return m
}

// return ReadOnly Map
func (m *Map[K, V]) ReadOnly() *Map[K, V] {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.readonly = true
	return m
}

// return Safe Map
func (m *Map[K, V]) Safe() *Map[K, V] {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.lock = &utils.Lock{}
	return m
}

// return key existence
func (m *Map[K, V]) Exists(k K) bool {
	if m == nil {
		return false
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	_, exists := m.data[k]
	return exists
}

// return value for key
func (m *Map[K, V]) Get(k K) V {
	if m == nil {
		return *new(V)
	}
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.data[k]
}

// return value and existence of key
func (m *Map[K, V]) GetFull(k K) (obj V, exists bool) {
	if m == nil {
		return
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	obj, exists = m.data[k]
	return
}

// set value for key
func (m *Map[K, V]) Set(k K, v V) {
	if m.readonly || m.init() != nil {
		return
	}
	if m.Hub != nil {
		m.Hub.Broadcast(types.WatchMsg[K, V]{
			Event: types.PutEvent,
			Item: types.Item[K, V]{
				Key:   k,
				Value: v,
			},
		})
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	m.data[k] = v
}

// delete key from Map
func (m *Map[K, V]) Delete(k K) {
	if m == nil || m.readonly {
		return
	}
	if m.Hub != nil {
		m.Hub.Broadcast(types.WatchMsg[K, V]{
			Event: types.DeleteEvent,
			Item: types.Item[K, V]{
				Key:   k,
				Value: m.Get(k),
			},
		})
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.data, k)
}

// run function with direct access to Map
func (m *Map[K, V]) Commit(fn func(data map[K]V)) {
	if m.readonly || m.init() != nil {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	fn(m.data)
}

// return iterator for safe iterating over Map
func (m *Map[K, V]) Iter() types.Iterator[K, V] {
	if m == nil {
		return nil
	}

	m.lock.RLock()
	iter := make(chan types.Item[K, V], len(m.data))

	go func() {
		defer m.lock.RUnlock()
		for k, v := range m.data {
			iter <- types.Item[K, V]{Key: k, Value: v}
		}
		close(iter)
	}()

	return iter
}

// range over Map
func (m *Map[K, V]) ForEach(fn func(k K, v V)) {
	if m == nil {
		return
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	for k, v := range m.data {
		fn(k, v)
	}
}

// return all Map keys
func (m *Map[K, V]) Keys() (keys []K) {
	if m == nil {
		return
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	i := 0
	keys = make([]K, len(m.data))

	for k := range m.data {
		keys[i] = k
		i++
	}

	return
}

// return all Map values
func (m *Map[K, V]) Values() (values []V) {
	if m == nil {
		return
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	i := 0
	values = make([]V, len(m.data))

	for _, v := range m.data {
		values[i] = v
		i++
	}

	return
}

// return Map length
func (m *Map[K, V]) Len() int {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return len(m.data)
}

// return Map copy
func (m *Map[K, V]) Copy() *Map[K, V] {
	if m == nil {
		return nil
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	copy := utils.DeepCopy(m.data)
	if copy == nil {
		return nil
	}
	return New(*copy)
}

func (m *Map[K, V]) marshal(marsh types.MarshalFunc) ([]byte, error) {
	if m == nil {
		return nil, types.ErrNilMap
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	if err := m.init(); err != nil {
		return nil, err
	}

	return marsh(m.data)
}

func (m *Map[K, V]) unmarshal(unmarsh types.UnmarshalFunc, data []byte) error {
	if m == nil {
		return types.ErrNilMap
	}

	if m.readonly {
		return types.ErrReadOnlyMap
	}

	if err := m.init(); err != nil {
		return err
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	return unmarsh(data, &m.data)
}

func (m *Map[K, V]) MarshalJSON() ([]byte, error) {
	return m.marshal(json.Marshal)
}

func (m *Map[K, V]) UnmarshalJSON(data []byte) error {
	return m.unmarshal(json.Unmarshal, data)
}

func (m *Map[K, V]) MarshalCBOR() ([]byte, error) {
	return m.marshal(cbor.Marshal)
}

func (m *Map[K, V]) UnmarshalCBOR(data []byte) error {
	return m.unmarshal(cbor.Unmarshal, data)
}

func (m *Map[K, V]) String() string {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return fmt.Sprintf("{%v, Safe: %v}", m.data, m.lock != nil)
}
