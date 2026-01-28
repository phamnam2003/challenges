package main

import (
	"log"
	"sync"
)

func main() {
	var sMap = &sync.Map{}

	sMap.Store("key", "value")
	v, ok := sMap.Load("key")
	if ok {
		println(v.(string))
	}

	sMap.Delete("key")

	v, ok = sMap.Load("key")
	if !ok {
		println("key not found")
	}

	swapped := sMap.CompareAndSwap("key", "value", "value1")
	log.Println("swapped: ", swapped)

	sMap.Store("key", "value")
	deleted := sMap.CompareAndDelete("key", "value")
	log.Println("deleted: ", deleted)

	for i := range 10 {
		sMap.Store(i, i*i)
	}

	sMap.Range(func(key, value any) bool {
		log.Printf("key: %v, value: %v\n", key, value)

		// return true to continue iteration loop
		return true
	})

	prev, loaded := sMap.Swap(1, 123)
	log.Printf("previous value: %v, loaded: %v\n", prev, loaded)
	prev, loaded = sMap.Swap(2, 4)
	log.Printf("previous value: %v, loaded: %v\n", prev, loaded)

	val, loaded := sMap.LoadAndDelete(1)
	log.Printf("value: %v, loaded: %v\n", val, loaded)
}
