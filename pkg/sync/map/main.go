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
}
