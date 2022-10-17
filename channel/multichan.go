package channel

import (
	"sync"
)

func Multi[T any](cs ...<-chan T) <-chan T {
	out := make(chan T)

	var wg sync.WaitGroup

	wg.Add(len(cs))

	for _, c := range cs {
		go func(c <-chan T) {
			for v := range c {
				out <- v
			}
			wg.Done()
		}(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
