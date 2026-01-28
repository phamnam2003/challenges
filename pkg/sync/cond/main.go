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

type Queue struct {
	items  []string
	lock   sync.Locker
	cond   *sync.Cond
	closed bool
}

func newQ() *Queue {
	lock := &sync.Mutex{}
	return &Queue{
		items: make([]string, 0),
		lock:  lock,
		cond:  sync.NewCond(lock),
	}
}

func (q *Queue) Consume(size int) {
	for range size {
		go func() {
			for {
				q.lock.Lock()
				for len(q.items) == 0 && !q.closed {
					log.Println("consume sleeping")
					q.cond.Wait()
				}
				if q.closed && len(q.items) == 0 {
					q.lock.Unlock()
					return
				}
				item := q.items[0]
				q.items = q.items[1:]
				q.lock.Unlock()

				log.Println("Consumed item:", item)
				time.Sleep(3 * time.Second)
			}
		}()
	}
}

func (q *Queue) Enqueue(item string) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.closed {
		return
	}
	q.items = append(q.items, item)
	q.cond.Signal()
}

func (q *Queue) Close() {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.closed = true
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
