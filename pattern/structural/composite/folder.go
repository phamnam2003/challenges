package main

import "fmt"

// Folder is a composite that can contain other components
type Folder struct {
	components []Component
	name       string
}

// search is a method that searches for a keyword in the folder and its components
func (f *Folder) search(keyword string) {
	fmt.Printf("Serching recursively for keyword %s in folder %s\n", keyword, f.name)
	for _, composite := range f.components {
		composite.search(keyword)
	}
}

// add adds a component to the folder
func (f *Folder) add(c Component) {
	f.components = append(f.components, c)
}
