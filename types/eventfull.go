package types

type Watcher[K comparable, V any] <-chan WatchMsg[K, V]

type WatchMsg[K comparable, V any] struct {
	Event EventType
	Item[K, V]
}
type EventType string

const (
	PutEvent    = "PUT"
	DeleteEvent = "DELETE"
)
