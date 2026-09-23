package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/kenchan0130/yahoo-realtime-search-feed/routers"
)

func main() {
	port := 8080
	if len(os.Args) > 1 {
		p, _ := strconv.Atoi(os.Args[1])
		port = p
	}

	router := routers.Init()
	err := router.Run(fmt.Sprintf(":%d", port))
	if err != nil {
		// Quote control characters in errors before writing them to the log.
		log.Fatalf("%q", err)
	}
}
