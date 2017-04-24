# eventqueue

A simple queue that helps you to process out-of-ordered events.

## Use case

You have a stream of events that arrive out of the order within some window
(let's say, within bunch of 100 events they are out of order). But for
each window N+1 events must be processed AFTER all events from window N are processed.
It makes sense to aggregate 100 events, sort them, process them, then receive next 100.
EventQueue simplifies this processing, especially when window size is not constant
(but limited by some constant number which can be used as "emitThreshold" in EventQueue)
and events arrive as an infinite stream and there is no clear separation between windows

![eventqueue](https://cloud.githubusercontent.com/assets/1905821/25341176/4a110756-2908-11e7-92b7-dde82ed01373.png)

## Usage

Check `eventqueue_test.go`

## Licence

MIT
