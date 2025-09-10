package main

import "fmt"

// File is a leaf node in the composite pattern
type File struct {
	name string
}

// search searches for a keyword in the file
func (f *File) search(keyword string) {
	fmt.Printf("Searching for keyword %s in file %s\n", keyword, f.name)
}

// getName returns the name of the file
func (f *File) getName() string {
	return f.name
}
