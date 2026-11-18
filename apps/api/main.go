package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	handler, err := NewHandler()
	if err != nil {
		log.Fatalf("load API fixtures: %v", err)
	}

	log.Printf("Perigee fixture API listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
