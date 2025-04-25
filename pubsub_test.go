package ibsync

import (
	"sync"
	"testing"
	"time"
)

// Test basic Publish and Subscribe
func TestPublishSubscribe(t *testing.T) {
	pubSub := NewPubSub()
	topic := "test_topic"
	msg := "test message"

	// Subscribe to a topic
	ch, unsubscribe := pubSub.Subscribe(topic)
	defer unsubscribe()

	// Publish a message to the topic
	pubSub.Publish(topic, msg)

	// Verify that the subscriber receives the message
	select {
	case received := <-ch:
		if received != msg {
			t.Errorf("Expected message %s, but got %s", msg, received)
		}
	case <-time.After(1 * time.Second):
		t.Error("Did not receive message on subscribed channel")
	}
}

// Test multiple subscribers
func TestMultipleSubscribers(t *testing.T) {
	pubSub := NewPubSub()
	topic := "multi_subscribers_topic"
	msg := "hello, subscribers!"

	// Subscribe multiple channels to the same topic
	ch1, unsubscribeCh1 := pubSub.Subscribe(topic)
	defer unsubscribeCh1()
	ch2, unsubscribeCh2 := pubSub.Subscribe(topic)
	defer unsubscribeCh2()

	// Publish a message to the topic
	pubSub.Publish(topic, msg)

	// Verify that both subscribers receive the message
	for _, ch := range []<-chan string{ch1, ch2} {
		select {
		case received := <-ch:

			if received != msg {
				t.Errorf("Expected message %s, but got %s", msg, received)
			}
		case <-time.After(1 * time.Second):

			t.Error("Did not receive message on subscribed channel")
		}
	}
}

// Test Unsubscribe
func TestUnsubscribe(t *testing.T) {
	pubSub := NewPubSub()
	topic := "unsubscribe_test"
	ch, _ := pubSub.Subscribe(topic)
	pubSub.Unsubscribe(topic, ch)

	select {
	case _, open := <-ch:
		if open {
			t.Error("Expected channel to be closed after unsubscribe")
		}
	default:
		// Success, channel was properly closed
	}
}

// Test UnsubscribeAll
func TestUnsubscribeAll(t *testing.T) {
	pubSub := NewPubSub()
	topic := "unsubscribe_all_test"
	ch1, _ := pubSub.Subscribe(topic)
	ch2, _ := pubSub.Subscribe(topic)

	pubSub.UnsubscribeAll(topic)

	// Verify that both channels are closed
	for _, ch := range []<-chan string{ch1, ch2} {
		select {
		case _, open := <-ch:
			if open {
				t.Error("Expected channel to be closed after UnsubscribeAll")
			}
		default:
			// Success, channel was properly closed
		}
	}
}

// Test Publish without subscribers
func TestPublishWithoutSubscribers(t *testing.T) {
	pubSub := NewPubSub()
	topic := "no_subscriber_topic"
	pubSub.Publish(topic, "no subscribers") // No channels subscribed, should proceed without errors
}

// Test Publish while unsubscribing in parallel
func TestPublishUnsubscribeParallel(t *testing.T) {
	pubSub := NewPubSub()
	topic := "parallel_publish_unsubscribe"
	msg := "parallel message"

	var wg sync.WaitGroup
	wg.Add(2)

	_, unsubscribe := pubSub.Subscribe(topic)

	go func() {
		defer wg.Done()
		pubSub.Publish(topic, msg)
	}()

	go func() {
		defer wg.Done()
		unsubscribe()
	}()

	wg.Wait()
}

func BenchmarkPubSub(b *testing.B) {
	pubSub := NewPubSub()
	reqID := 1
	eurusd := NewForex("EUR", "IDEALPRO", "USD")
	contractDetails := NewContractDetails()
	contractDetails.Contract = *eurusd

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ch, cancel := pubSub.Subscribe(reqID)
		pubSub.Publish(reqID, Encode(contractDetails))
		msg := <-ch
		var cd ContractDetails
		if err := Decode(&cd, msg); err != nil {
			return
		}
		cancel()
	}
}

func BenchmarkPubSubBuffered(b *testing.B) {
	pubSub := NewPubSub()
	reqID := 1
	eurusd := NewForex("EUR", "IDEALPRO", "USD")
	contractDetails := NewContractDetails()
	contractDetails.Contract = *eurusd

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ch, cancel := pubSub.Subscribe(reqID, 100)
		pubSub.Publish(reqID, Encode(contractDetails))
		msg := <-ch
		var cd ContractDetails
		if err := Decode(&cd, msg); err != nil {
			return
		}
		cancel()
	}
}
