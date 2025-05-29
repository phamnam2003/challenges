package main

import (
	"fmt"
	"sync"
)

type single struct{}

var (
	singleInstance *single
	once           sync.Once
)

func getInstance() *single {
	once.Do(func() {
		if singleInstance == nil {
			singleInstance = &single{}
			fmt.Println("create new instance")
		}
	})
	return singleInstance
}
