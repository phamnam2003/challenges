package main

import (
	"fmt"
	"sync"
)

// single is the singleton struct
type single struct{}

// define instance and once variables
var (
	singleInstance *single
	once           sync.Once
)

// getInstance func returns the singleton instance
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
