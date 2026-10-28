package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/kenchan0130/yahoo-realtime-search-feed/routers"
)

func main() {
	port := 8080
	if len(os.Args) > 1 {
		p, _ := strconv.Atoi(os.Args[1])
		port = p
	}

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           routers.Init(),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	log.Printf("Listening and serving HTTP on %s", server.Addr)
	err := server.ListenAndServe()
	if err != nil {
		// Quote control characters in errors before writing them to the log.
		log.Fatalf("%q", err)
	}
}
