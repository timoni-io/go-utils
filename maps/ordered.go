package maps

import (
	"github.com/timoni-io/go-utils"
	"github.com/timoni-io/go-utils/types"
	"sort"

	"golang.org/x/exp/constraints"
)

type OrderedMap[K constraints.Ordered, V any] struct {
	Map[K, V]
	lessFunc types.SortFunction[K]

	sorted     bool
	sortedKeys []K
}

func NewOrdered[K constraints.Ordered, V any](data map[K]V, lessFunc types.SortFunction[K]) *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		Map: Map[K, V]{
			data: data,
		},
		lessFunc: lessFunc,
	}
}

func (m *OrderedMap[K, V]) init() error {
	if m == nil {
		return types.ErrNilMap
	}

	if m.data == nil {
		m.data = map[K]V{}
	}

	if m.lessFunc == nil {
		m.lessFunc = func(data []K, i, j int) bool {
			return data[i] <= data[j]
		}
	}

	return nil
}

func (m *OrderedMap[K, V]) sort() {
	if m == nil {
		return
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	// skip sort if sorted
	if m.sorted {
		return
	}

	m.init()

	// get keys slice
	i := 0
	m.sortedKeys = make([]K, len(m.data))
	for k := range m.data {
		m.sortedKeys[i] = k
		i++
	}

	// sort keys
	sort.Slice(m.sortedKeys, func(i, j int) bool { return m.lessFunc(m.sortedKeys, i, j) })
	m.sorted = true
}

// return ReadOnly Map
func (m *OrderedMap[K, V]) ReadOnly() *OrderedMap[K, V] {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.readonly = true
	return m
}

// return Safe Map
func (m *OrderedMap[K, V]) Safe() *OrderedMap[K, V] {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.lock = &utils.Lock{}
	return m
}

// set value for key
func (m *OrderedMap[K, V]) Set(k K, v V) {
	if m.readonly || m.init() != nil {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	m.data[k] = v
	m.sorted = false
}

// delete key from Map
func (m *OrderedMap[K, V]) Delete(k K) {
	if m == nil || m.readonly {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.data, k)
	m.sorted = false
}

// run function with direct access to Map
func (m *OrderedMap[K, V]) Commit(fn func(data map[K]V)) {
	if m.readonly || m.init() != nil {
		return
	}

	fn(m.data)
	m.sorted = false
}

// return iterator for safe iterating over Map
func (m *OrderedMap[K, V]) Iter() types.Iterator[K, V] {
	if m == nil {
		return nil
	}

	// sort before returning
	m.sort()

	m.lock.RLock()
	iter := make(chan types.Item[K, V], len(m.sortedKeys))

	go func() {
		defer m.lock.RUnlock()

		for _, k := range m.sortedKeys {
			iter <- types.Item[K, V]{Key: k, Value: m.data[k]}
		}
		close(iter)
	}()

	return iter
}

// range over Map
func (m *OrderedMap[K, V]) ForEach(fn func(k K, v V)) {
	if m == nil {
		return
	}

	// sort before returning
	m.sort()

	m.lock.Lock()
	defer m.lock.Unlock()

	for _, k := range m.sortedKeys {
		fn(k, m.data[k])
	}
}

// return all Map keys
func (m *OrderedMap[K, V]) Keys() (keys []K) {
	if m == nil {
		return
	}

	// sort before returning
	m.sort()

	m.lock.Lock()
	defer m.lock.Unlock()

	return m.sortedKeys
}

// return all Map values
func (m *OrderedMap[K, V]) Values() (values []V) {
	if m == nil {
		return
	}

	// sort before returning
	m.sort()

	m.lock.Lock()
	defer m.lock.Unlock()

	i := 0
	values = make([]V, len(m.sortedKeys))

	for _, k := range m.sortedKeys {
		values[i] = m.data[k]
		i++
	}

	return
}

// return Map copy
func (m *OrderedMap[K, V]) Copy() *OrderedMap[K, V] {
	if m == nil {
		return nil
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	copy := utils.DeepCopy(m.data)
	if copy == nil {
		return nil
	}

	return NewOrdered(*copy, nil)
}
