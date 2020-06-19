package event_channel

import (
	"context"
	"errors"
	"sync"
	"time"
)

type eventFunc func(e *EventData)

type EventChannel struct {
	events map[string]*Event
	l      sync.RWMutex
}

type EventData struct {
	Data interface{}
}

type Event struct {
	f       eventFunc
	Timeout context.Context
	Done    chan struct{}
	isOnce  bool
}

func New() *EventChannel {
	return &EventChannel{
		events: make(map[string]*Event),
	}
}

func (ec *EventChannel) On(eventName string, eventFunc eventFunc) {
	ec.l.Lock()
	defer ec.l.Unlock()

	ec.events[eventName] = &Event{
		f: eventFunc,
	}
}

func (ec *EventChannel) Once(eventName string, eventFunc eventFunc, timeout time.Duration) *Event {
	ec.l.Lock()
	defer ec.l.Unlock()

	ctx, _ := context.WithTimeout(context.TODO(), timeout)

	ec.events[eventName] = &Event{
		f:       eventFunc,
		Done:    make(chan struct{}, 0),
		Timeout: ctx,
		isOnce:  true,
	}

	return ec.events[eventName]
}

func (ec *EventChannel) Has(eventName string) bool {
	ec.l.RLock()
	defer ec.l.RUnlock()

	if _, ok := ec.events[eventName]; ok {
		return true
	}

	return false
}

func (ec *EventChannel) Remove(eventName string) bool {
	ec.l.Lock()
	defer ec.l.Unlock()

	if _, ok := ec.events[eventName]; ok {
		delete(ec.events, eventName)

		return true
	}

	return false
}

func (ec *EventChannel) Fire(eventName string, data interface{}) error {
	ec.l.Lock()
	defer ec.l.Unlock()

	if e, ok := ec.events[eventName]; ok {
		if e.isOnce {
			delete(ec.events, eventName)

			if e.Timeout.Err() != nil {
				return e.Timeout.Err()
			}

			e.f(&EventData{
				Data: data,
			})

			close(e.Done)

			return nil
		}

		e.f(&EventData{
			Data: data,
		})

		return nil
	}

	return errors.New("event not found")
}
