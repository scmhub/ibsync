package ibsync

import (
	"fmt"
	"slices"
	"sync"
)

const defaultBufferSize = 5

// UnsubscribeFunc is a function type that can be used to unsubscribe from a topic.
type UnsubscribeFunc func()

// PubSub is a thread-safe publish-subscribe implementation.
// It manages topic subscriptions and message distribution.
type PubSub struct {
	mu     sync.RWMutex
	topics map[string][]chan string // Map of topics with a list of subscriber channels
}

// NewPubSub creates and initializes a new PubSub instance.
func NewPubSub() *PubSub {
	return &PubSub{
		topics: make(map[string][]chan string),
	}
}

// Subscribe creates a new subscriber for a topic and returns a channel to receive messages.
// It supports optional buffer size specification.
func (ps *PubSub) Subscribe(topic any, size ...int) (<-chan string, UnsubscribeFunc) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	t := fmt.Sprint(topic)

	buffSize := defaultBufferSize
	if len(size) > 0 {
		buffSize = size[0]
	}
	ch := make(chan string, buffSize)

	ps.topics[t] = append(ps.topics[t], ch)

	return ch, func() { ps.Unsubscribe(topic, ch) }
}

// Unsubscribe removes a specific subscriber channel from a topic.
// It closes the channel and removes the topic if no subscribers remain.
func (ps *PubSub) Unsubscribe(topic any, subscriberChan <-chan string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	t := fmt.Sprint(topic)

	subscribers, exists := ps.topics[t]
	if !exists {
		return
	}

	for i, ch := range subscribers {
		if ch == subscriberChan {
			ps.topics[t] = slices.Delete(subscribers, i, i+1)
			close(ch)
			if len(ps.topics[t]) == 0 {
				delete(ps.topics, t)
			}
			return
		}
	}
}

// UnsubscribeAll removes all subscribers from a topic.
// It closes all subscriber channels and deletes the topic from the topics map.
func (ps *PubSub) UnsubscribeAll(topic any) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	t := fmt.Sprint(topic)

	// If the topic exists, close all subscriber channels
	if subscribers, exists := ps.topics[t]; exists {
		for _, ch := range subscribers {
			close(ch) // Close each subscriber channel
		}
		delete(ps.topics, t) // Remove the topic from the map
	}
}

// Publish sends a message to all subscribers of a topic.
func (ps *PubSub) Publish(topic any, msg string) {
	ps.mu.RLock()
	t := fmt.Sprint(topic)

	subscribers, exists := ps.topics[t]
	if !exists {
		ps.mu.RUnlock()
		return
	}

	subsCopy := make([]chan string, len(subscribers))
	copy(subsCopy, subscribers)
	ps.mu.RUnlock()

	for _, ch := range subsCopy {
		ch <- msg // must be blocking or "end" msgs can get through before msgs and will close the channel too early
	}
}
