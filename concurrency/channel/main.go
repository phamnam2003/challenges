package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
)

func crawler() {
	// init max worker ur computer can create.
	maxWorker := 2
	// init array contains urls website need crawl data
	arrWebsCrawl := []string{"https://facebook.com", "https://youtube.com", "https://instagram.com", "https://go.dev", "https://github.com", "https://gitlab.com", "https://postgresql.org", "https://sqlc.dev", "https://redis.io", "https://rabbitmq.com"}
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
			for urlWebsite := range queuCrawl {
				res, err := http.Get(urlWebsite)
				if err != nil {
					log.Println(err)
					continue
				}
				// close ngay sau khi xử lý xong
				func() {
					defer res.Body.Close()

					bytesRes, err := io.ReadAll(res.Body)
					if err != nil {
						log.Println(err)
						return
					}

					// Trích domain từ URL
					parsedURL, err := url.Parse(urlWebsite)
					if err != nil {
						log.Println(err)
						return
					}
					domain := parsedURL.Hostname()

					// Ghi vào file tên là domain.txt
					fileName := fmt.Sprintf("%s.txt", domain)
					err = os.WriteFile(fileName, bytesRes, 0644)
					if err != nil {
						log.Println(err)
						return
					}
					log.Printf("Worker ghi nội dung website %s vào file %s\n", urlWebsite, fileName)
				}()
			}
			// when waitGroup is 0, finished process.
			waitGroup.Done()
		}(i)
	}
	// wait waitGroup is 0.
	waitGroup.Wait()
}

func main() {
	crawler()
}
