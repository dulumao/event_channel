package event_channel

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	is := assert.New(t)

	var wg sync.WaitGroup
	ec := New()

	event1 := ec.Once("event1", func(e *EventData) {
		fmt.Printf("%#v\n", e.Data)
	}, 1*time.Second)

	ec.On("event2", func(e *EventData) {
		fmt.Printf("%#v\n", e.Data)
	})

	wg.Add(1)

	go func() {
		select {
		case <-event1.Done:
			println("event1 end")
			wg.Done()
		case <-event1.Timeout.Done():
			println("event1 timeout done")
			wg.Done()
		}
	}()

	time.Sleep(2 * time.Second)

	is.Error(ec.Fire("event1", "event1 value"), "context deadline exceeded")
	is.Error(ec.Fire("event1", "event1 value"), "event not found")

	is.Nil(ec.Fire("event2", "event2 value"))

	is.True(ec.Remove("event2"))
	is.False(ec.Remove("event3"))

	is.Error(ec.Fire("event2", "event2 value"), "event not found")

	wg.Wait()
}