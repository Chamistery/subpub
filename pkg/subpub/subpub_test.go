package subpub

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestPublishSubscribe(t *testing.T) {
	bus := NewSubPub()
	defer bus.Close(context.Background())

	var mu sync.Mutex
	received := []interface{}{}

	sub, err := bus.Subscribe("topic", func(msg interface{}) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, msg)
	})
	if err != nil {
		t.Fatalf("subscribe error: %v", err)
	}

	messages := []string{"a", "b", "c"}
	for _, m := range messages {
		bus.Publish("topic", m)
	}

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	if len(received) != len(messages) {
		t.Errorf("expected %d messages, got %d", len(messages), len(received))
	}
	for i, m := range messages {
		if received[i] != m {
			t.Errorf("message %d: expected %v, got %v", i, m, received[i])
		}
	}
	mu.Unlock()

	sub.Unsubscribe()
}

func TestSlowSubscriberDoesNotBlock(t *testing.T) {
	bus := NewSubPub()
	defer bus.Close(context.Background())

	fastDone := make(chan struct{})
	_, _ = bus.Subscribe("slow", func(msg interface{}) {
		// slow handler sleeps
		time.Sleep(100 * time.Millisecond)
	})
	_, _ = bus.Subscribe("slow", func(msg interface{}) {
		close(fastDone)
	})

	bus.Publish("slow", "event")

	select {
	case <-fastDone:
		// ok
	case <-time.After(50 * time.Millisecond):
		t.Error("fast subscriber was blocked by slow one")
	}
}
