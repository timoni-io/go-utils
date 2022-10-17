package maps

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/timoni-io/go-utils"
	"github.com/timoni-io/go-utils/set"
	"github.com/timoni-io/go-utils/types"

	"github.com/fxamacker/cbor/v2"
)

// WeightedMap is a map of Weighted values. Higher weight means higher priority (descending).
type WeightedMap[K comparable, V any] struct {
	Map[K, types.Weighted[V]]

	sorted     bool
	sortedKeys []K

	weightCount int
}

func NewWeighted[K comparable, V any](data map[K]types.Weighted[V]) *WeightedMap[K, V] {
	return &WeightedMap[K, V]{
		Map: Map[K, types.Weighted[V]]{
			data: data,
		},
	}
}

func NewWeightedMapFromSlice[K comparable, V any](keys []K, data []V) *WeightedMap[K, V] {
	m := map[K]types.Weighted[V]{}
	for i, value := range data {
		m[keys[i]] = types.Weighted[V]{Value: value, Weight: uint32(i)}
	}

	return NewWeighted(m)
}

func (m *WeightedMap[K, V]) init() error {
	if m == nil {
		return types.ErrNilMap
	}

	if m.data == nil {
		m.data = map[K]types.Weighted[V]{}
	}

	return nil
}

// sort keys
func (m *WeightedMap[K, V]) sort() {
	if m == nil {
		return
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	// skip sort if sorted
	if m.sorted {
		return
	}

	weights := set.New[uint32]()

	// get keys slice
	i := 0
	m.sortedKeys = make([]K, len(m.data))
	for k, v := range m.data {
		weights.Add(v.Weight)
		m.sortedKeys[i] = k
		i++
	}

	m.weightCount = weights.Length()

	// sort keys
	sort.Slice(m.sortedKeys, func(i, j int) bool {
		a := m.data[m.sortedKeys[i]]
		b := m.data[m.sortedKeys[j]]
		return a.Weight > b.Weight
	})

	m.sorted = true
}

func (m *WeightedMap[K, V]) ReadOnly() *WeightedMap[K, V] {
	m.lock.RLock()
	defer m.lock.RUnlock()

	m.readonly = true
	return m
}

func (m *WeightedMap[K, V]) Safe() *WeightedMap[K, V] {
	m.lock.RLock()
	defer m.lock.RUnlock()

	m.lock = &utils.Lock{}
	return m
}

// return value for key
func (m *WeightedMap[K, V]) Get(k K) V {
	if m == nil {
		return *new(V)
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.data[k].Value
}

// return value and existence of key
func (m *WeightedMap[K, V]) GetFull(k K) (obj V, exists bool) {
	if m == nil {
		return
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	weight, exists := m.data[k]
	return weight.Value, exists
}

// set value for key
func (m *WeightedMap[K, V]) Set(k K, v V) {
	if m.readonly || m.init() != nil {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	m.data[k] = types.Weighted[V]{Value: v, Weight: 0}
	m.sorted = false
}

// set value for key with weight
func (m *WeightedMap[K, V]) SetWeighted(k K, v types.Weighted[V]) {
	if m.readonly || m.init() != nil {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	m.data[k] = v
	m.sorted = false
}

// delete key from Map
func (m *WeightedMap[K, V]) Delete(k K) {
	if m == nil || m.readonly {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.data, k)
	m.sorted = false
}

// run function with direct access to Map
func (m *WeightedMap[K, V]) Commit(fn func(data map[K]types.Weighted[V])) {
	if m.readonly || m.init() != nil {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	fn(m.data)
	m.sorted = false
}

// return iterator for safe iterating over Map
func (m *WeightedMap[K, V]) Iter() types.Iterator[K, V] {
	if m == nil {
		return nil
	}

	// sort before returning
	m.sort()

	m.lock.RLock()
	iter := make(chan types.Item[K, V], len(m.data))

	go func() {
		defer m.lock.RUnlock()
		for _, k := range m.sortedKeys {
			iter <- types.Item[K, V]{Key: k, Value: m.data[k].Value}
		}
		close(iter)
	}()

	return iter
}

// WeightIter... Returned iterators are in descending order.
func (m *WeightedMap[K, V]) WeightIter() <-chan types.Iterator[K, V] {
	if m == nil {
		return nil
	}

	// sort before returning
	m.sort()

	m.lock.RLock()
	weightChan := make(chan types.Iterator[K, V], m.weightCount)

	go func() {
		defer m.lock.RUnlock()

		lastWeight := -1
		var iter chan types.Item[K, V]
		for _, k := range m.sortedKeys {
			v := m.data[k]
			if lastWeight != int(v.Weight) {
				// Close previous
				if iter != nil {
					close(iter)
				}

				iter = make(chan types.Item[K, V], len(m.data))
				weightChan <- iter
				lastWeight = int(v.Weight)
			}

			iter <- types.Item[K, V]{Key: k, Value: v.Value}
		}
		close(iter)
		close(weightChan)
	}()

	return weightChan
}

// range over Map
func (m *WeightedMap[K, V]) ForEach(fn func(k K, v V)) {
	if m == nil {
		return
	}

	// sort before returning
	m.sort()

	m.lock.Lock()
	defer m.lock.Unlock()

	for _, k := range m.sortedKeys {
		fn(k, m.data[k].Value)
	}
}

// return all Map keys
func (m *WeightedMap[K, V]) Keys() (keys []K) {
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
func (m *WeightedMap[K, V]) Values() (values []V) {
	if m == nil {
		return
	}

	// sort before returning
	m.sort()

	m.lock.Lock()
	defer m.lock.Unlock()

	i := 0
	values = make([]V, len(m.data))

	for _, k := range m.sortedKeys {
		values[i] = m.data[k].Value
		i++
	}

	return
}

// return Map copy
func (m *WeightedMap[K, V]) Copy() *WeightedMap[K, V] {
	if m == nil {
		return nil
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	copy := utils.DeepCopy(m.data)
	if copy == nil {
		return nil
	}
	return NewWeighted(*copy)
}

func (m *WeightedMap[K, V]) marshal(marsh types.MarshalFunc) ([]byte, error) {
	// init empty map
	if err := m.init(); err != nil {
		return nil, err
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	rawMap := map[K]V{}

	for k, v := range m.data {
		rawMap[k] = v.Value
	}

	return marsh(rawMap)
}

func (m *WeightedMap[K, V]) unmarshal(unmarsh types.UnmarshalFunc, data []byte) error {
	// init empty map
	if err := m.init(); err != nil {
		return err
	}

	if m.readonly {
		return types.ErrReadOnlyMap
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	// TODO: try unmarshal weighted values
	// var finalMap map[K]WeightedValue[V]
	// err := json.Unmarshal(data, &finalMap)
	// if err == nil {
	// 	// OK
	// 	m.data = finalMap
	// 	m.sort()
	// 	return nil
	// }

	// unmarshal weightless values
	var rawMap map[K]V
	err := unmarsh(data, &rawMap)
	if err != nil {
		return err
	}

	// assign existing weights
	finalMap := map[K]types.Weighted[V]{}
	for k, v := range rawMap {
		var weight uint32

		if x, exists := m.data[k]; exists {
			weight = x.Weight
		}

		finalMap[k] = types.Weighted[V]{Value: v, Weight: weight}
	}

	// save final map
	m.data = finalMap

	return nil
}

func (m *WeightedMap[K, V]) MarshalJSON() ([]byte, error) {
	return m.marshal(json.Marshal)
}

func (m *WeightedMap[K, V]) UnmarshalJSON(data []byte) error {
	return m.unmarshal(json.Unmarshal, data)
}

func (m *WeightedMap[K, V]) MarshalCBOR() ([]byte, error) {
	return m.marshal(cbor.Marshal)
}

func (m *WeightedMap[K, V]) UnmarshalCBOR(data []byte) error {
	return m.unmarshal(cbor.Unmarshal, data)
}

func (m *WeightedMap[K, V]) String() string {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return fmt.Sprint(m.data)
}
