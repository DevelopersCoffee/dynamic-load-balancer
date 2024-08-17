package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"com.developerscoffee/dynamic-load-balancer/common"
	STRATEGY "com.developerscoffee/dynamic-load-balancer/strategy"
)

type LB struct {
	backends []*common.Backend
	strategy STRATEGY.BalancingStrategy
}

func InitLB() *LB {
	backends := []*common.Backend{
		{Host: "localhost", Port: 8081, IsHealthy: true},
		{Host: "localhost", Port: 8082, IsHealthy: true},
		{Host: "localhost", Port: 8083, IsHealthy: true},
		{Host: "localhost", Port: 8084, IsHealthy: true},
	}

	lb := &LB{
		backends: backends,
		strategy: STRATEGY.NewRRBalancingStrategy(backends), // Using Round-Robin by default
	}
	return lb
}

func (lb *LB) proxyHTTP(w http.ResponseWriter, req *http.Request) {
	// Select a backend using the configured balancing strategy
	backend := lb.strategy.GetNextBackend(common.IncomingReq{})
	log.Printf("Request -> Routing to backend: %s", backend.String())

	// Try to connect to the selected backend
	backendURL := fmt.Sprintf("http://%s:%d%s", backend.Host, backend.Port, req.RequestURI)
	resp, err := http.Get(backendURL)
	if err != nil {
		http.Error(w, "Backend is currently unavailable", http.StatusServiceUnavailable)
		log.Printf("Error connecting to backend %s: %s", backend.String(), err.Error())
		return
	}
	defer resp.Body.Close()

	// Copy the response from the backend to the client
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	lb := InitLB()

	// Start HTTP server to handle incoming HTTP requests
	http.HandleFunc("/", lb.proxyHTTP)
	log.Println("Load Balancer is listening on HTTP port 9090")
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}
