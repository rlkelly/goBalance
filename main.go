package main

import (
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

func random(max float64) float64 {
	rand.Seed(time.Now().Unix())
	return rand.Float64() * max
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s <port>", os.Args[0])
	}
	if _, err := strconv.Atoi(os.Args[1]); err != nil {
		log.Fatalf("Invalid port: %s (%s)\n", os.Args[1], err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		println("--->", os.Args[1], req.URL.String(), math.Sqrt(random(999999999999)))
	})
	http.ListenAndServe(":"+os.Args[1], nil)
}
