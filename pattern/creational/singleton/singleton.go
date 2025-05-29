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
	if singleInstance == nil {
		once.Do(func() {
			if singleInstance == nil {
				singleInstance = &single{}
				fmt.Println("create new instance")
			}
		})
	} else {
		fmt.Println("get instance")
	}
	return singleInstance
}
