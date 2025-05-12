package main

import (
	"fmt"
	"sync"
	"time"
)

func crawler() {
	// init max worker ur computer can create.
	maxWorker := 2
	// init array contains urls website need crawl data
	arrWebsCrawl := []string{"https://facebook.com", "https://youtube.com", "https://instagram.com", "https://go.dev", "https://github.com", "https://gitlab.com", "https://postgresql.org", "https://sqlc.dev", "https://redis.io", "https://rabbimq.com"}
	// make queue buffered channel with length array website
	queuCrawl := make(chan string, len(arrWebsCrawl))
	// create WaitGroup for calculate timing process in Goroutines
	waitGroup := new(sync.WaitGroup)

	// push url data into queue channel
	for _, urlWeb := range arrWebsCrawl {
		queuCrawl <- urlWeb
	}
	// when channel full, close channel (optional)
	close(queuCrawl)

	// each Worker, queue Channel will get data, activity will be action
	for i := range maxWorker {
		// add data to waitGroup, when value in waitGroup equals 0, block code is Done.
		waitGroup.Add(1)
		go func(idxCurrWorker int) {
			// get data in queue channel
			for url := range queuCrawl {
				fmt.Printf("Worker %d is crawling website: %s\n", idxCurrWorker, url)
				time.Sleep(time.Second)
			}
			// when waitGroup is 0, finished process.
			waitGroup.Done()
		}(i)
	}
	// wait waitGroup is 0.
	waitGroup.Wait()
}
