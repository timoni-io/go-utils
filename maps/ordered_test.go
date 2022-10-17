package maps

import (
	"fmt"
	"testing"
)

func TestOrderedForEach(t *testing.T) {
	m := NewOrdered[string, string](nil, nil)
	if m == nil {
		t.Fail()
	}

	m.Set("1", "x")
	m.Set("2", "x")
	m.Set("3", "x")

	i := 0
	m.ForEach(func(k, v string) {
		i++

		if k != fmt.Sprint(i) {
			t.Errorf("%s != %d", k, i)
		}
	})
}

func TestOrderedKeys(t *testing.T) {
	m := NewOrdered[string, string](nil, nil)
	if m == nil {
		t.Fail()
	}

	m.Set("1", "a")
	m.Set("2", "b")
	m.Set("3", "c")

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

func TestOrderedValues(t *testing.T) {
	m := NewOrdered[string, string](nil, nil)
	if m == nil {
		t.Fail()
	}

	m.Set("1", "a")
	m.Set("2", "b")
	m.Set("3", "c")

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
