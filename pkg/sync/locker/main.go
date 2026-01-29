package main

import (
	"fmt"
	"sync"
)

var (
	cache  = map[string]string{}
	rwLock sync.RWMutex
)

func read(id int, key string, wg *sync.WaitGroup) {
	defer wg.Done()

	rwLock.RLock()
	defer rwLock.RUnlock()

	fmt.Printf("Reader %d: READ_DONE cache[%q] → %q\n", id, key, cache[key])
}

func write(key, value string, wg *sync.WaitGroup) {
	defer wg.Done()

	rwLock.Lock()
	defer rwLock.Unlock()

	cache[key] = value
	fmt.Printf("Writer: WRITER_DONE cache[%q] = %q\n", key, value)
}

func main() {
	// RLock: nhiều reader đọc cùng lúc
	// Lock: writer phải độc quyền
	var wg sync.WaitGroup

	// Ghi dữ liệu trước
	cache["user"] = "Chris"

	// 5 readers đọc cùng lúc
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go read(i, "user", &wg)
	}
	wg.Wait()
	for _, v := range []string{"jupiter", "James", "Linda"} {
		wg.Add(1)
		go write("user", v, &wg)
	}
	wg.Wait()
}
