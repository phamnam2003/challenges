package main

import (
	"log"
	"os"
	"sync"
)

func main() {
	var onceFunc = sync.OnceFunc(func() {
		log.Println("run once this func")
	})
	onceValue := sync.OnceValue(func() int {
		sum := 0
		for i := 0; i < 1000; i++ {
			sum += i
		}
		log.Printf("Computed once: [sum: %d]", sum)
		return sum
	})

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			onceFunc()
		}()
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			const want = 499500
			got := onceValue()
			if got != want {
				log.Println("want", want, "got", got)
			}
		}()
	}
	onceValues := sync.OnceValues(func() ([]byte, error) {
		log.Println("Reading file once")
		return os.ReadFile("D:\\nampham2003\\challenges\\tech\\di\\main.go")
	})
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			data, err := onceValues()
			if err != nil {
				log.Println("error:", err)
			}
			_ = data // Ignore the data for this example
			done <- true
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
	wg.Wait()
}
