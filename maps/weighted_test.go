package maps

import (
	"fmt"
	"github.com/timoni-io/go-utils/types"
	"testing"
)

func TestWeightedForEach(t *testing.T) {
	m := NewWeighted(map[string]types.Weighted[string]{
		"1": {Value: "x", Weight: 3},
		"2": {Value: "x", Weight: 2},
		"3": {Value: "x", Weight: 1},
	})

	if m == nil {
		t.Fail()
	}

	i := 0
	m.ForEach(func(k, v string) {
		i++

		if k != fmt.Sprint(i) {
			t.Errorf("%s != %d", k, i)
		}
	})
}

func TestWeightedKeys(t *testing.T) {
	m := NewWeighted(map[string]types.Weighted[string]{
		"1": {Value: "x", Weight: 3},
		"2": {Value: "x", Weight: 2},
		"3": {Value: "x", Weight: 1},
	})
	if m == nil {
		t.Fail()
	}

	keys := m.Keys()

	if len(keys) != 3 {
		t.Fail()
	}

	expected := []string{
		"1",
		"2",
		"3",
	}

	for i, k := range keys {
		if expected[i] != k {
			t.Errorf("%s != %s", expected[i], k)
		}
	}
}

func TestWeightedValues(t *testing.T) {
	m := NewWeighted(map[string]types.Weighted[string]{
		"1": {Value: "a", Weight: 3},
		"2": {Value: "b", Weight: 2},
		"3": {Value: "c", Weight: 1},
	})
	if m == nil {
		t.Fail()
	}

	values := m.Values()

	if len(values) != 3 {
		t.Fail()
	}

	expected := []string{
		"a",
		"b",
		"c",
	}

	for i, v := range values {
		if expected[i] != v {
			t.Errorf("%s != %s", expected[i], v)
		}
	}
}
