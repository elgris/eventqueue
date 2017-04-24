package eventqueue

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testEvent struct {
	sequence uint64
	content  int
}

// testEventCollection implements sort.Interface
type testEventCollection []*testEvent

func (c testEventCollection) Len() int           { return len(c) }
func (c testEventCollection) Less(i, j int) bool { return c[i].sequence < c[j].sequence }
func (c testEventCollection) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

func TestEventChannelEmitting(t *testing.T) {
	var comparator ComparatorFunc = func(a, b interface{}) bool {
		return a.(testEvent).sequence < b.(testEvent).sequence
	}
	ch := make(chan interface{}, 10)
	queue := NewEventQueue(5, ch, comparator)

	events := []testEvent{
		testEvent{
			sequence: 100,
			content:  10,
		},
		testEvent{
			sequence: 1,
			content:  12,
		},
		testEvent{
			sequence: 12,
			content:  15,
		},
		testEvent{
			sequence: 150,
			content:  151,
		},
	}

	queue.Push(events[0])
	queue.Push(events[1])
	queue.Push(events[2])
	queue.Push(events[3])
	queue.Flush()

	close(ch)

	expectedSequence := []int{1, 2, 0, 3}
	for i, expectedIndex := range expectedSequence {
		item := <-ch
		require.Equal(t, events[expectedIndex], item.(testEvent), "test case #%d", i)
	}
}
