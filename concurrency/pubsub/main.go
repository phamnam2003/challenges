package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type (
	// subscriber is a channel to receive messages.
	subscriber chan any

	// topicFunc is a function to filter messages.
	topicFunc func(v any) bool
)

type Publisher struct {
	// Atomic Lock Mutex
	m sync.RWMutex

	// queue buffer channel size
	buffer int

	// timeout for each publish
	timeout time.Duration

	// subscribers map has been subscribed to the publisher
	subscribers map[subscriber]topicFunc
}

// NewPublisher creates a new Publisher with a given timeout and buffer size.
func NewPublisher(publishTimeout time.Duration, buffer int) *Publisher {
	return &Publisher{
		timeout:     publishTimeout,
		buffer:      buffer,
		subscribers: make(map[subscriber]topicFunc),
	}
}

// SubscribeTopic subscribes to a topic with a filter function.
func (p *Publisher) SubscribeTopic(topic topicFunc) chan any {
	ch := make(chan any, p.buffer)
	p.m.Lock()
	// register the subscriber channel with the topic function
	p.subscribers[ch] = topic
	p.m.Unlock()

	return ch
}

// Subscribe subscribes to all topics.
func (p *Publisher) Subscribe() chan any {
	return p.SubscribeTopic(nil)
}

// Publish publishes a message to topics
func (p *Publisher) Publish(v any) {
	p.m.RLock()
	defer p.m.RUnlock()

	var wg sync.WaitGroup
	for sub, topic := range p.subscribers {
		wg.Add(1)
		go p.sendTopic(sub, topic, v, &wg)
	}

	wg.Wait()
}

func (p *Publisher) sendTopic(sub subscriber, topic topicFunc, v any, wg *sync.WaitGroup) {
	defer wg.Done()
	if topic != nil && !topic(v) {
		return
	}

	select {
	case sub <- v:
	case <-time.After(p.timeout):
	}
}

// Evict removes a subscriber from the publisher.
func (p *Publisher) Evict(sub chan any) {
	p.m.Lock()
	defer p.m.Unlock()

	delete(p.subscribers, sub)
	close(sub)
}

func (p *Publisher) Close() {
	p.m.Lock()
	defer p.m.Unlock()

	for sub := range p.subscribers {
		delete(p.subscribers, sub)
		close(sub)
	}
}

func main() {
	// initialize a new publisher
	p := NewPublisher(100*time.Millisecond, 10)
	defer p.Close()

	// subscribe to all topics
	all := p.Subscribe()

	// subscribe to topics that contain "golang"
	golang := p.SubscribeTopic(func(v any) bool {
		if s, ok := v.(string); ok {
			return strings.Contains(s, "golang")
		}
		return false
	})

	p.Publish("hello, world")
	p.Publish("hello, golang")

	go func() {
		for msg := range golang {
			fmt.Printf("[GOLANG] %s\n", msg)
		}
	}()

	go func() {
		for msg := range all {
			fmt.Printf("[ALL] %s\n", msg)
		}
	}()

	time.Sleep(1 * time.Second)
}
