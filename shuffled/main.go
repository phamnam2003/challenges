package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func main() {
	inputFile := "words.txt"
	numFiles := 10
	outputDir := "shuffled"

	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		log.Fatalf("Cannot create output directory: %v", err)
	}

	words, err := readWords(inputFile)
	if err != nil {
		log.Fatalf("Failed to read words: %v", err)
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var wg sync.WaitGroup
	for i := 1; i <= numFiles; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// Copy slice và xáo trộn
			shuffled := make([]string, len(words))
			copy(shuffled, words)
			r.Shuffle(len(shuffled), func(i, j int) {
				shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
			})

			outputPath := filepath.Join(outputDir, fmt.Sprintf("shuffled_%02d.txt", index))
			if err := writeWords(outputPath, shuffled); err != nil {
				log.Printf("Failed to write file %s: %v", outputPath, err)
			} else {
				log.Printf("Generated: %s", outputPath)
			}
		}(i)
	}

	wg.Wait()
	log.Println("Done generating shuffled files.")
}

func readWords(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if text != "" {
			words = append(words, text)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return words, nil
}

func writeWords(filePath string, words []string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, word := range words {
		_, err := writer.WriteString(word + "\n")
		if err != nil {
			return err
		}
	}
	return writer.Flush()
}
