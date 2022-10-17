package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func DeepCopy[T any](src T) *T {
	data, err := json.Marshal(src)
	if err != nil {
		return nil
	}

	dst := new(T)
	err = json.Unmarshal(data, dst)
	if err != nil {
		return nil
	}
	return dst
}

// WaitWithTimeout waits for the waitgroup for the specified max timeout.
// Returns error if waiting timed out.
func WaitWithTimeout(wg *sync.WaitGroup, timeout time.Duration) error {
	c := make(chan struct{})
	go func() {
		wg.Wait()
		close(c)
	}()

	select {
	case <-c:
		return nil
	case <-time.After(timeout):
		return errors.New("timeout")
	}
}

func PanicHandler() error {
	if err := recover(); err != nil {
		return fmt.Errorf("%v", err)
	}

	return nil
}

func All[T any](val []T, fn func(x T) bool) bool {
	for _, v := range val {
		if !fn(v) {
			return false
		}
	}

	return true
}

func Any[T any](val []T, fn func(x T) bool) bool {
	for _, v := range val {
		if fn(v) {
			return true
		}
	}

	return false
}

func Must[T any](out T, err error) T {
	if err != nil {
		panic(err)
	}
	return out
}

func First[T any](in []T) T {
	if len(in) > 0 {
		return in[0]
	}
	return *new(T)
}

func PanicOnNil(v any) {
	if v == nil {
		panic("value is nil")
	}
}

func Ternary[T any](cond bool, t, f T) T {
	if cond {
		return t
	}
	return f
}
