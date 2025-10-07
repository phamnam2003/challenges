package main

import (
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"
)

func rolldice(w http.ResponseWriter, r *http.Request) {
	roll := 1 + rand.IntN(6)

	res := strconv.Itoa(roll) + "\n"
	if _, err := io.WriteString(w, res); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}
