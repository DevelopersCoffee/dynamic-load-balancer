package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Define the port as a command-line argument
	portPtr := flag.Int("port", 8080, "Port for the server to listen on")
	flag.Parse()

	port := *portPtr
	if port <= 0 {
		fmt.Println("Please provide a valid port number")
		os.Exit(1)
	}

	startServer(port)
}

func startServer(port int) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		msg := fmt.Sprintf("Received request on server running on port %d", port)
		fmt.Println(msg)
		w.Write([]byte(msg))
	})

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Server running on port %d\n", port)
	log.Fatal(http.ListenAndServe(addr, nil))
}
