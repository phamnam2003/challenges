package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Queue represents a thread-safe queue with condition variable support.
type Queue struct {
	items  []string
	lock   sync.Locker
	cond   *sync.Cond
	closed bool
}

// newQ initializes and returns a new Queue instance.
func newQ() *Queue {
	lock := &sync.Mutex{}
	return &Queue{
		items: make([]string, 0),
		lock:  lock,
		cond:  sync.NewCond(lock),
	}
}

// Consume starts consuming items from the queue using the specified number of goroutines.
func (q *Queue) Consume(size int) {
	// create go routines with limit size, but this is not recommended for production use
	// because it may lead to resource exhaustion.
	for range size {
		go func() {
			// loop infinitely to consume items
			for {
				q.lock.Lock()
				for len(q.items) == 0 && !q.closed {
					log.Println("consume sleeping")
					// move go routine to g-parking state with empty queue and not closed
					q.cond.Wait()
				}
				if q.closed && len(q.items) == 0 {
					q.lock.Unlock()
					return
				}
				// get message from queue
				item := q.items[0]
				// assign new slice to queue items
				q.items = q.items[1:]
				q.lock.Unlock()

				log.Println("Consumed item:", item)
				time.Sleep(3 * time.Second)
			}
		}()
	}
}

// Enqueue adds an item to the queue and signals a waiting consumer.
func (q *Queue) Enqueue(item string) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.closed {
		return
	}
	q.items = append(q.items, item)

	// signal one consumer to handle message
	q.cond.Signal()
}

// Close marks the queue as closed and broadcasts to all waiting consumers.
func (q *Queue) Close() {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.closed = true

	// wake up all consumers to shutdown go routines
	q.cond.Broadcast()
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	q := newQ()
	defer q.Close()

	for i := range 5 {
		go func(idx int) {
			q.Enqueue(fmt.Sprintf("value index : %d", idx))
		}(i)
	}

	q.Consume(2)

	<-ctx.Done()
	log.Println("Shutting down gracefully...")
}
