package eventqueue

import (
	"container/heap"
	"sync"
)

// EventQueue allows to process out-of-order incoming events in
// ordered way. Basically this is a wrapper around a buffer which aggregates events
// and sorts the events as they come. Once buffer size hits treshold, EventQueue
// starts emitting the events through output channel.
// See NewEventQueue for details about instantiating EventQueue.
//
// Algorithm complexity of adding the event is O(logN).
// Algorithm complexity of popping the event out is also O(logN)
//
// Use case:
// You have a stream of events that arrive out of the order within some window
// (let's say, within bunch of 100 events they are out of order). But for
// each window N+1 events must be processed AFTER all events from window N are processed.
// It makes sense to aggregate 100 events, sort them, process them, then receive next 100.
// EventQueue simplifies this processing, especially when window size is not constant
// (but limited by some constant number which can be used as "emitThreshold" in EventQueue)
// and events arrive as an infinite stream and there is no clear separation between windows
type EventQueue struct {
	emitThreshold int
	output        chan<- interface{}
	queue         eventPriorityQueue

	lock sync.Mutex
}

// Comparator is an entity that helps to sort incoming events.
// Expect "a" anb "b" be events, so cast them to appropriate event type
// if needed
type Comparator interface {
	Less(a, b interface{}) bool
}

// ComparatorFunc implements Comparator interface.
// Useful for providing comparators as simple functions
type ComparatorFunc func(a, b interface{}) bool

func (f ComparatorFunc) Less(a, b interface{}) bool { return f(a, b) }

// NewEventQueue creates EventQueue
// - emitThreshold - this value sets number of events to accumulate before they start
// emitting through output channel (see Channel() method)
// - outputChannel - is an INPUT channel which will receive emitted events. The channel is injected
// by a client of EventQueue and controlled by it
// - comparator helps to sort your events
func NewEventQueue(emitThreshold int, outputChannel chan<- interface{}, comparator Comparator) *EventQueue {
	return &EventQueue{
		emitThreshold: emitThreshold,
		output:        outputChannel,

		queue: eventPriorityQueue{
			data:       make([]interface{}, 0, 10),
			comparator: comparator,
		},
	}
}

// Push adds an event to the queue in an ordered matter
func (es *EventQueue) Push(item interface{}) {
	es.lock.Lock()
	defer es.lock.Unlock()

	heap.Push(&es.queue, item)
	if es.queue.Len() >= es.emitThreshold {
		es.output <- es.popUnprotected()
	}
}

// Flush pushes the rest of the aggregated events to output channel.
// At the end the queue is empty
func (es *EventQueue) Flush() {
	es.lock.Lock()
	defer es.lock.Unlock()

	l := es.Len()
	for i := 0; i < l; i++ {
		es.output <- es.popUnprotected()
	}
}

// Len returns current length of the queue
func (es *EventQueue) Len() int { return es.queue.Len() }

func (es *EventQueue) popUnprotected() interface{} { return heap.Pop(&es.queue) }

// eventPriorityQueue is just an implementation of PriorityQueue (MinHeap) for events
type eventPriorityQueue struct {
	data       []interface{}
	comparator Comparator
}

func (pq eventPriorityQueue) Len() int           { return len(pq.data) }
func (pq eventPriorityQueue) Less(i, j int) bool { return pq.comparator.Less(pq.data[i], pq.data[j]) }
func (pq eventPriorityQueue) Swap(i, j int)      { pq.data[i], pq.data[j] = pq.data[j], pq.data[i] }

func (pq *eventPriorityQueue) Push(x interface{}) { pq.data = append(pq.data, x) }
func (pq *eventPriorityQueue) Pop() interface{} {
	old := pq.data
	n := len(old)
	item := old[n-1]
	pq.data = old[0 : n-1]
	return item
}
