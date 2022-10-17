package channel

// import (
// 	"sync"
// 	"time"
// 	"utils/slice"
// )

// type Broker[T any] struct {
// 	Tx        chan<- T
// 	receivers []chan T
// 	mutex     sync.Mutex
// 	buffer    int
// }

// func NewBroker[T any](buffer int) *Broker[T] {

// 	tx := make(chan T, buffer)
// 	rx := &Broker[T]{
// 		Tx:        tx,
// 		receivers: []chan T{},
// 		mutex:     sync.Mutex{},
// 	}

// 	go func(tx chan T, rx *Broker[T]) {

// 		for v := range tx {
// 			rx.mutex.Lock()

// 			for _, c := range rx.receivers {

// 				go func(c chan T) {

// 					defer func() {
// 						// recover from 'send on closed channel
// 						if r := recover(); r != nil {
// 							// log.Println("Recovering from panic in broker:", r)
// 							return
// 						}
// 					}()

// 					timer := time.NewTimer(time.Second)
// 					defer timer.Stop()

// 					// send with 1s timeout
// 					select {
// 					case c <- v:
// 					case <-timer.C:
// 					}

// 				}(c)
// 			}

// 			rx.mutex.Unlock()
// 		}

// 		// When tx is closed
// 		for _, c := range rx.receivers {
// 			close(c)
// 		}

// 	}(tx, rx)

// 	return rx
// }

// func (b *Broker[T]) Close() {
// 	close(b.Tx)
// }

// func (b *Broker[T]) Copy() chan T {
// 	b.mutex.Lock()
// 	defer b.mutex.Unlock()

// 	x := make(chan T, b.buffer)
// 	b.receivers = append(b.receivers, x)

// 	return x
// }

// func (b *Broker[T]) Delete(c chan T) {
// 	b.mutex.Lock()
// 	defer b.mutex.Unlock()

// 	slice.Remove(&b.receivers, c)

// 	select {
// 	case <-c:
// 	default:
// 	}

// 	close(c)
// }

// func (b *Broker[T]) Debug() []chan T {
// 	return b.receivers
// }
